//go:build linux && cgo
// +build linux,cgo

package midilinux

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/leandrodaf/midi/v2/internal/alsa"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
	"golang.org/x/sys/unix"
)

var (
	ErrNoMIDIDevices     = errors.New("no MIDI devices found")
	ErrInvalidMIDIDevice = errors.New("invalid MIDI device")
	ErrDeviceNotSelected = errors.New("no MIDI device selected: call SelectDevice first")
)

// ClientMid is the Linux ALSA raw MIDI client.
type ClientMid struct {
	logger            contracts.Logger
	eventChannel      atomic.Value
	midiEventFilter   *contracts.MIDIEventFilter
	channelBufferSize int

	mu                 sync.Mutex
	devices            []alsa.DeviceInfo
	selectedDeviceAddr string
	capturing          bool
	outCh              chan contracts.MIDI
	closeChOnce        sync.Once
	wg                 sync.WaitGroup

	// cancel pipe: write end is closed/written to signal the read goroutine to stop.
	cancelPipeR int
	cancelPipeW int
}

// NewMIDIClient creates a Linux ALSA raw-MIDI client. No hardware is opened
// at construction time; call SelectDevice to choose an input device and
// StartCapture to begin receiving events.
func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Info("MIDI client created for Linux (ALSA)")
	return &ClientMid{
		logger:            options.Logger,
		midiEventFilter:   options.MIDIEventFilter,
		channelBufferSize: options.ChannelBufferSize,
		cancelPipeR:       -1,
		cancelPipeW:       -1,
	}, nil
}

// ListDevices enumerates ALSA raw-MIDI input devices and caches the result
// so that SelectDevice can map a stable integer index to a hardware address.
func (m *ClientMid) ListDevices() ([]contracts.DeviceInfo, error) {
	devs, err := alsa.EnumerateInputs()
	if err != nil {
		return nil, fmt.Errorf("error listing ALSA MIDI inputs: %w", err)
	}
	if len(devs) == 0 {
		m.logger.Warn(ErrNoMIDIDevices.Error())
		return nil, ErrNoMIDIDevices
	}

	m.mu.Lock()
	m.devices = devs
	m.mu.Unlock()

	result := make([]contracts.DeviceInfo, len(devs))
	for i, d := range devs {
		result[i] = contracts.DeviceInfo{
			Name:         d.Name,
			EntityName:   d.SubdeviceName,
			Manufacturer: d.HWAddr,
		}
	}
	return result, nil
}

// SelectDevice records the hardware address of the ALSA device at index
// deviceID. If capture is already running it is stopped first.
func (m *ClientMid) SelectDevice(deviceID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.capturing {
		m.mu.Unlock()
		_ = m.Stop()
		m.mu.Lock()
	}

	if m.devices == nil {
		devs, err := alsa.EnumerateInputs()
		if err != nil {
			return fmt.Errorf("error retrieving ALSA MIDI inputs: %w", err)
		}
		m.devices = devs
	}

	if deviceID < 0 || deviceID >= len(m.devices) {
		m.logger.Error(ErrInvalidMIDIDevice.Error())
		return ErrInvalidMIDIDevice
	}

	dev := m.devices[deviceID]
	m.selectedDeviceAddr = dev.HWAddr

	m.logger.Info("MIDI device selected",
		contracts.IntField("deviceID", deviceID),
		contracts.StringField("deviceName", dev.Name),
		contracts.StringField("hwAddr", dev.HWAddr))
	return nil
}

// closeOutCh closes the output channel exactly once.
func (m *ClientMid) closeOutCh() {
	m.closeChOnce.Do(func() {
		if m.outCh != nil {
			close(m.outCh)
		}
	})
}

// Stop halts capture. Safe to call concurrently.
func (m *ClientMid) Stop() error {
	m.mu.Lock()
	if !m.capturing {
		m.mu.Unlock()
		return nil
	}
	m.logger.Info("Stopping MIDI capture")
	m.capturing = false
	m.eventChannel.Store(make(chan contracts.MIDI))

	// Signal the read goroutine via the cancel pipe.
	cancelW := m.cancelPipeW
	m.mu.Unlock()

	if cancelW >= 0 {
		_, _ = unix.Write(cancelW, []byte{0})
	}

	m.wg.Wait()
	m.closeOutCh()
	return nil
}

// StartCapture opens the previously selected ALSA device, creates a cancel
// pipe pair for cooperative shutdown, and launches a goroutine that calls
// readLoop to parse incoming MIDI bytes. The returned channel is closed when
// ctx is cancelled or Stop is called.
func (m *ClientMid) StartCapture(ctx context.Context) (<-chan contracts.MIDI, error) {
	if err := m.Stop(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	if m.selectedDeviceAddr == "" {
		m.mu.Unlock()
		return nil, ErrDeviceNotSelected
	}
	addr := m.selectedDeviceAddr
	m.mu.Unlock()

	// Open the ALSA device fresh for this capture session.
	raw, err := alsa.OpenInput(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to open ALSA device %q: %w", addr, err)
	}

	midifd, err := raw.FD()
	if err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("failed to get ALSA fd: %w", err)
	}

	// Create a cancel pipe for this capture session.
	pipeFds := make([]int, 2)
	if err := unix.Pipe(pipeFds); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("failed to create cancel pipe: %w", err)
	}
	cancelR, cancelW := pipeFds[0], pipeFds[1]

	size := m.channelBufferSize
	if size <= 0 {
		size = 100
	}
	ch := make(chan contracts.MIDI, size)

	m.mu.Lock()
	m.outCh = ch
	m.closeChOnce = sync.Once{}
	m.eventChannel.Store((chan contracts.MIDI)(ch))
	m.capturing = true
	m.cancelPipeR = cancelR
	m.cancelPipeW = cancelW
	m.mu.Unlock()

	m.logger.Info("Starting MIDI event capture (ALSA)",
		contracts.StringField("device", addr))

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer raw.Close()
		defer unix.Close(cancelR)
		defer unix.Close(cancelW)
		m.readLoop(raw, midifd, cancelR)
	}()

	go func() {
		<-ctx.Done()
		_ = m.Stop()
	}()

	return ch, nil
}

// readLoop reads raw MIDI bytes using poll(2) with a cancel pipe so that
// Stop() wakes it immediately without blocking.
func (m *ClientMid) readLoop(raw *alsa.RawMIDI, midifd, cancelfd int) {
	buf := make([]byte, 64)
	var parser midiParser

	pfds := []unix.PollFd{
		{Fd: int32(midifd), Events: unix.POLLIN},
		{Fd: int32(cancelfd), Events: unix.POLLIN},
	}

	for {
		_, err := unix.Poll(pfds, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return
		}

		// Cancel pipe signalled.
		if pfds[1].Revents&unix.POLLIN != 0 {
			return
		}

		if pfds[0].Revents&unix.POLLIN == 0 {
			continue
		}

		n, err := raw.Read(buf)
		if err != nil || n == 0 {
			return
		}

		for _, b := range buf[:n] {
			cmd, note, vel, ok := parser.feed(b)
			if !ok {
				continue
			}
			event := contracts.MIDI{
				Timestamp: uint64(time.Now().UTC().UnixNano()),
				Command:   cmd,
				Note:      note,
				Velocity:  vel,
			}
			if !contracts.IsCommandAllowed(event.Command, m.midiEventFilter) {
				m.logger.Debug(fmt.Sprintf("MIDI command 0x%X filtered out", cmd))
				continue
			}
			eventChannel, _ := m.eventChannel.Load().(chan contracts.MIDI)
			if eventChannel == nil {
				return
			}
			select {
			case eventChannel <- event:
			default:
				m.logger.Warn("Event buffer full; dropping MIDI event")
			}
		}
	}
}

// WatchDevices returns a channel that emits a DeviceEvent whenever a MIDI
// device is connected or disconnected. On Linux, this is implemented by
// polling ListDevices every 2 seconds and diffing against the previous list.
func (m *ClientMid) WatchDevices(ctx context.Context) (<-chan contracts.DeviceEvent, error) {
	evCh := make(chan contracts.DeviceEvent, 16)

	prev, _ := m.ListDevices()

	go func() {
		defer close(evCh)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				curr, _ := m.ListDevices()
				diffDevices(prev, curr, evCh)
				prev = curr
			}
		}
	}()

	return evCh, nil
}

// diffDevices compares two device lists and sends DeviceAdded / DeviceRemoved
// events to evCh for each difference.
func diffDevices(prev, curr []contracts.DeviceInfo, evCh chan<- contracts.DeviceEvent) {
	for _, d := range curr {
		if !containsDevice(prev, d) {
			select {
			case evCh <- contracts.DeviceEvent{Type: contracts.DeviceAdded, Device: d}:
			default:
			}
		}
	}
	for _, d := range prev {
		if !containsDevice(curr, d) {
			select {
			case evCh <- contracts.DeviceEvent{Type: contracts.DeviceRemoved, Device: d}:
			default:
			}
		}
	}
}

func containsDevice(list []contracts.DeviceInfo, d contracts.DeviceInfo) bool {
	for _, item := range list {
		if item.Name == d.Name && item.Manufacturer == d.Manufacturer {
			return true
		}
	}
	return false
}
// midiParser reassembles raw MIDI bytes into 3-byte channel-voice messages
// using the running-status rule. SysEx and real-time messages are discarded.
type midiParser struct {
	status  byte
	data    [2]byte
	dataPos int
}

// feed processes one raw MIDI byte. It returns (cmd, note, vel, true) when a
// complete 3-byte channel-voice message has been assembled, otherwise false.
func (p *midiParser) feed(b byte) (cmd, note, vel byte, ok bool) {
	if b >= 0x80 {
		switch {
		case b >= 0xF8: // real-time messages (1 byte) — skip
			return 0, 0, 0, false
		case b == 0xF0: // SysEx start — reset until 0xF7
			p.status = 0
			p.dataPos = 0
			return 0, 0, 0, false
		case b == 0xF7: // SysEx end
			p.status = 0
			p.dataPos = 0
			return 0, 0, 0, false
		case b >= 0xF0: // other system-common messages — skip
			p.status = 0
			p.dataPos = 0
			return 0, 0, 0, false
		default: // channel voice status (0x80–0xEF)
			p.status = b & 0xF0
			p.dataPos = 0
			return 0, 0, 0, false
		}
	}

	// Data byte
	if p.status == 0 {
		return 0, 0, 0, false
	}
	p.data[p.dataPos] = b
	p.dataPos++
	if p.dataPos < 2 {
		return 0, 0, 0, false
	}
	// Running status: keep p.status, reset only dataPos.
	p.dataPos = 0
	return p.status, p.data[0], p.data[1], true
}
