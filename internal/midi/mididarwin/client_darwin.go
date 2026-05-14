//go:build darwin
// +build darwin

package mididarwin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/leandrodaf/midi/v2/internal/coremidi"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

var (
	ErrNoMIDIDevices        = errors.New("no MIDI devices found")
	ErrInvalidMIDIDevice    = errors.New("invalid MIDI device")
	ErrMIDIConnectionError  = errors.New("error connecting to MIDI device")
	ErrCreateInputPort      = errors.New("error creating input port")
	ErrIncompleteMIDIPacket = errors.New("incomplete MIDI packet")
)

type internalPortConnection interface {
	Disconnect()
}

type ClientMid struct {
	logger            contracts.Logger
	eventChannel      atomic.Value
	client            coremidi.Client
	inputPort         coremidi.InputPort
	portConn          internalPortConnection
	midiEventFilter   *contracts.MIDIEventFilter
	coreMIDIConfig    *contracts.CoreMIDIConfig
	mu                sync.Mutex
	capturing         bool
	wg                sync.WaitGroup
	channelBufferSize int
	outCh             chan contracts.MIDI
	closeChOnce       sync.Once
}

func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	if options.CoreMIDIConfig == nil {
		options.CoreMIDIConfig = &contracts.CoreMIDIConfig{ClientName: "GO MIDI Client"}
	}
	client, err := coremidi.NewClient(options.CoreMIDIConfig.ClientName)
	if err != nil {
		return nil, err
	}
	options.Logger.Info("MIDI client successfully created")
	return &ClientMid{
		logger:            options.Logger,
		client:            client,
		midiEventFilter:   options.MIDIEventFilter,
		coreMIDIConfig:    options.CoreMIDIConfig,
		channelBufferSize: options.ChannelBufferSize,
	}, nil
}

func (m *ClientMid) ListDevices() ([]contracts.DeviceInfo, error) {
	sources, err := coremidi.AllSources()
	if err != nil {
		return nil, fmt.Errorf("error listing MIDI sources: %w", err)
	}
	if len(sources) == 0 {
		m.logger.Warn(ErrNoMIDIDevices.Error())
		return nil, ErrNoMIDIDevices
	}
	devices := make([]contracts.DeviceInfo, len(sources))
	for i, source := range sources {
		entity := source.Entity()
		devices[i] = contracts.DeviceInfo{
			Name:         source.Name(),
			EntityName:   entity.Name(),
			Manufacturer: entity.Manufacturer(),
		}
	}
	return devices, nil
}

func (m *ClientMid) SelectDevice(deviceID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sources, err := coremidi.AllSources()
	if err != nil {
		return fmt.Errorf("error retrieving MIDI sources: %w", err)
	}
	if deviceID < 0 || deviceID >= len(sources) {
		m.logger.Error(ErrInvalidMIDIDevice.Error())
		return ErrInvalidMIDIDevice
	}
	if m.portConn != nil {
		m.portConn.Disconnect()
		m.portConn = nil
	}
	source := sources[deviceID]
	m.logger.Info("MIDI device selected",
		contracts.IntField("deviceID", deviceID),
		contracts.StringField("deviceName", source.Name()))

	m.inputPort, err = coremidi.NewInputPort(m.client, "Input Port", m.handleMIDIMessage)
	if err != nil {
		m.logger.Error(ErrCreateInputPort.Error())
		return fmt.Errorf("%w: %v", ErrCreateInputPort, err)
	}
	m.portConn, err = m.inputPort.Connect(source)
	if err != nil {
		m.logger.Error(ErrMIDIConnectionError.Error())
		return fmt.Errorf("%w: %v", ErrMIDIConnectionError, err)
	}
	m.logger.Info("MIDI device successfully connected")
	return nil
}

func (m *ClientMid) handleMIDIMessage(source coremidi.Source, packet coremidi.Packet) {
	m.wg.Add(1)
	defer m.wg.Done()

	eventChannel, _ := m.eventChannel.Load().(chan contracts.MIDI)
	if eventChannel == nil {
		return
	}
	if len(packet.Data) < 3 {
		m.logger.Warn(ErrIncompleteMIDIPacket.Error())
		return
	}
	event := contracts.MIDI{
		Timestamp: uint64(time.Now().UTC().UnixNano()),
		Command:   packet.Data[0],
		Note:      packet.Data[1],
		Velocity:  packet.Data[2],
	}
	if !contracts.IsCommandAllowed(event.Command, m.midiEventFilter) {
		return
	}
	select {
	case eventChannel <- event:
	default:
		m.logger.Warn("Event buffer full; dropping MIDI event")
	}
}

// stopLocked performs stop cleanup. Caller MUST hold m.mu.
func (m *ClientMid) stopLocked() {
	if !m.capturing {
		return
	}
	m.capturing = false
	if m.portConn != nil {
		m.portConn.Disconnect()
		m.portConn = nil
	}
	m.eventChannel.Store(make(chan contracts.MIDI)) // unbuffered dummy; default case in handleMIDIMessage prevents blocking
	m.logger.Info("MIDI capture stopped")
}

func (m *ClientMid) closeOutCh() {
	m.closeChOnce.Do(func() {
		if m.outCh != nil {
			close(m.outCh)
		}
	})
}

func (m *ClientMid) Stop() error {
	m.mu.Lock()
	if !m.capturing {
		m.mu.Unlock()
		return nil
	}
	m.logger.Info("Stopping MIDI capture")
	m.stopLocked()
	m.mu.Unlock()

	m.wg.Wait()
	m.closeOutCh()
	return nil
}

func (m *ClientMid) StartCapture(ctx context.Context) (<-chan contracts.MIDI, error) {
	if err := m.Stop(); err != nil {
		return nil, err
	}

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
	m.mu.Unlock()

	m.logger.Info("Starting MIDI event capture")

	go func() {
		<-ctx.Done()
		_ = m.Stop()
	}()

	return ch, nil
}
