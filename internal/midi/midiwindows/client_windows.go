//go:build windows
// +build windows

package midiwindows

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/leandrodaf/midi/v2/sdk/contracts"
	"golang.org/x/sys/windows"
)

var (
	ErrNoMIDIDevices       = errors.New("no MIDI devices found")
	ErrInvalidDeviceHandle = errors.New("invalid MIDI device handle")
	ErrOpenFailed          = errors.New("failed to open MIDI device")
	ErrStartCaptureFailed  = errors.New("failed to start MIDI capture")
	ErrStopCaptureFailed   = errors.New("failed to stop MIDI capture")
	ErrCloseDeviceFailed   = errors.New("failed to close MIDI device")
)

type HMIDIIN windows.Handle

const (
	CALLBACK_FUNCTION = 0x00030000
	MIDI_IO_STATUS    = 0x00000020
)

const (
	MIM_OPEN      = 0x3C1
	MIM_CLOSE     = 0x3C2
	MIM_DATA      = 0x3C3
	MIM_ERROR     = 0x3C5
	MIM_LONGERROR = 0x3C6
	MIM_MOREDATA  = 0x3CC
)

type midiInCaps struct {
	wMid           uint16
	wPid           uint16
	vDriverVersion uint32
	szPname        [32]uint16
	dwSupport      uint32
}

type ClientMid struct {
	logger            contracts.Logger
	eventChannel      atomic.Value
	handle            HMIDIIN
	portConn          bool
	mu                sync.Mutex
	callback          uintptr
	midiEventFilter   *contracts.MIDIEventFilter
	coreMIDIConfig    *contracts.CoreMIDIConfig
	channelBufferSize int
	outCh             chan contracts.MIDI
	closeChOnce       sync.Once
}

var (
	winmm                = windows.NewLazySystemDLL("winmm.dll")
	procMidiInGetNumDevs = winmm.NewProc("midiInGetNumDevs")
	procMidiInGetDevCaps = winmm.NewProc("midiInGetDevCapsW")
	procMidiInOpen       = winmm.NewProc("midiInOpen")
	procMidiInStart      = winmm.NewProc("midiInStart")
	procMidiInStop       = winmm.NewProc("midiInStop")
	procMidiInClose      = winmm.NewProc("midiInClose")
)

func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Info("MIDI client created for Windows")
	return &ClientMid{
		logger:            options.Logger,
		midiEventFilter:   options.MIDIEventFilter,
		coreMIDIConfig:    options.CoreMIDIConfig,
		channelBufferSize: options.ChannelBufferSize,
	}, nil
}

func (m *ClientMid) ListDevices() ([]contracts.DeviceInfo, error) {
	r0, _, _ := procMidiInGetNumDevs.Call()
	if uint32(r0) == 0 {
		m.logger.Warn(ErrNoMIDIDevices.Error())
		return nil, ErrNoMIDIDevices
	}
	numDevices := uint32(r0)
	devices := make([]contracts.DeviceInfo, numDevices)
	for i := uint32(0); i < numDevices; i++ {
		var caps midiInCaps
		r1, _, _ := procMidiInGetDevCaps.Call(uintptr(i), uintptr(unsafe.Pointer(&caps)), unsafe.Sizeof(caps))
		if r1 != 0 {
			m.logger.Warn(fmt.Sprintf("Failed to get information for MIDI device %d", i))
			continue
		}
		name := windows.UTF16ToString(caps.szPname[:])
		devices[i] = contracts.DeviceInfo{
			Name:         name,
			EntityName:   name,
			Manufacturer: fmt.Sprintf("MID: %d PID: %d", caps.wMid, caps.wPid),
		}
	}
	return devices, nil
}

func (m *ClientMid) SelectDevice(deviceID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.portConn {
		if err := m.stopCapture(); err != nil {
			return fmt.Errorf("failed to stop previous MIDI capture: %w", err)
		}
	}

	m.callback = windows.NewCallback(midiInCallback)
	r1, _, _ := procMidiInOpen.Call(
		uintptr(unsafe.Pointer(&m.handle)),
		uintptr(deviceID),
		m.callback,
		uintptr(unsafe.Pointer(m)),
		uintptr(CALLBACK_FUNCTION|MIDI_IO_STATUS),
	)
	if r1 != 0 {
		return fmt.Errorf("%w %d: return code %d", ErrOpenFailed, deviceID, r1)
	}
	m.portConn = true
	m.logger.Info(fmt.Sprintf("MIDI device %d connected", deviceID))
	return nil
}

func (m *ClientMid) closeOutCh() {
	m.closeChOnce.Do(func() {
		if m.outCh != nil {
			close(m.outCh)
		}
	})
}

func (m *ClientMid) StartCapture(ctx context.Context) (<-chan contracts.MIDI, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.portConn {
		return nil, fmt.Errorf("no MIDI device selected")
	}

	size := m.channelBufferSize
	if size <= 0 {
		size = 100
	}
	ch := make(chan contracts.MIDI, size)
	m.outCh = ch
	m.closeChOnce = sync.Once{}
	m.eventChannel.Store((chan contracts.MIDI)(ch))

	if m.handle == 0 {
		return nil, ErrInvalidDeviceHandle
	}
	r1, _, _ := procMidiInStart.Call(uintptr(m.handle))
	if r1 != 0 {
		m.eventChannel.Store(make(chan contracts.MIDI))
		return nil, fmt.Errorf("%w: return code %d", ErrStartCaptureFailed, r1)
	}

	m.logger.Info("MIDI capture started")

	go func() {
		<-ctx.Done()
		_ = m.Stop()
	}()

	return ch, nil
}

func midiInCallback(hMidiIn uintptr, wMsg uint32, dwInstance uintptr, dwParam1 uintptr, dwParam2 uintptr) uintptr {
	m := (*ClientMid)(unsafe.Pointer(dwInstance))

	switch wMsg {
	case MIM_OPEN:
		m.logger.Info("MIDI device opened")
	case MIM_CLOSE:
		m.logger.Info("MIDI device closed")
	case MIM_DATA:
		if dwParam2 == 0 {
			return 0
		}
		status := byte(dwParam1 & 0xFF)
		data1 := byte((dwParam1 >> 8) & 0xFF)
		data2 := byte((dwParam1 >> 16) & 0xFF)

		command := status & 0xF0
		channel := status & 0x0F

		event := contracts.MIDI{
			Timestamp: uint64(time.Now().UTC().UnixNano()),
			Command:   command,
			Note:      data1,
			Velocity:  data2,
		}

		if !contracts.IsCommandAllowed(event.Command, m.midiEventFilter) {
			m.logger.Debug(fmt.Sprintf("MIDI command 0x%X filtered out", command))
			return 0
		}

		if command == byte(contracts.NoteOn) && event.Velocity == 0 || command == byte(contracts.NoteOff) {
			m.logger.Debug(fmt.Sprintf("Note Off: Channel %d, Note %d", channel+1, event.Note))
		} else if command == byte(contracts.NoteOn) {
			m.logger.Debug(fmt.Sprintf("Note On: Channel %d, Note %d, Velocity %d", channel+1, event.Note, event.Velocity))
		}

		if ch, ok := m.eventChannel.Load().(chan contracts.MIDI); ok && ch != nil {
			select {
			case ch <- event:
			default:
				m.logger.Warn("MIDI event channel is full; event discarded")
			}
		}
	case MIM_ERROR, MIM_LONGERROR:
		m.logger.Error(fmt.Sprintf("MIDI error: msg=0x%X", wMsg))
	case MIM_MOREDATA:
		m.logger.Debug("Received MIM_MOREDATA message; ignored")
	default:
		m.logger.Warn(fmt.Sprintf("Unknown MIDI message: 0x%X", wMsg))
	}
	return 0
}

func (m *ClientMid) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.portConn {
		return nil
	}
	if err := m.stopCapture(); err != nil {
		return fmt.Errorf("failed to stop MIDI capture: %w", err)
	}
	m.logger.Info("MIDI capture stopped and device closed")
	m.closeOutCh()
	return nil
}

// WatchDevices returns a channel that emits a DeviceEvent whenever a MIDI
// device is connected or disconnected. On Windows, this is implemented by
// polling midiInGetNumDevs every 2 seconds and diffing against the previous list.
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


	if m.handle == 0 {
		return ErrInvalidDeviceHandle
	}
	r1, _, _ := procMidiInStop.Call(uintptr(m.handle))
	if r1 != 0 {
		return fmt.Errorf("%w: return code %d", ErrStopCaptureFailed, r1)
	}
	r1, _, _ = procMidiInClose.Call(uintptr(m.handle))
	if r1 != 0 {
		return fmt.Errorf("%w: return code %d", ErrCloseDeviceFailed, r1)
	}
	m.portConn = false
	m.handle = 0
	m.eventChannel.Store(make(chan contracts.MIDI))
	return nil
}
