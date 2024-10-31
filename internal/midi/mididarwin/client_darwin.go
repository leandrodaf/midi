//go:build darwin
// +build darwin

package mididarwin

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/leandrodaf/midi/sdk/contracts"
	"github.com/youpy/go-coremidi"
)

// Error definitions
var (
	ErrNoMIDIDevices        = errors.New("no MIDI devices found")
	ErrInvalidMIDIDevice    = errors.New("invalid MIDI device")
	ErrMIDIConnectionError  = errors.New("error connecting to MIDI device")
	ErrCreateInputPort      = errors.New("error creating input port")
	ErrIncompleteMIDIPacket = errors.New("incomplete MIDI packet")
)

// internalPortConnection handles port disconnection.
type internalPortConnection interface {
	Disconnect()
}

// ClientMid manages MIDI on Darwin systems.
type ClientMid struct {
	logger          contracts.Logger
	eventChannel    atomic.Value
	client          coremidi.Client
	inputPort       coremidi.InputPort
	portConn        internalPortConnection
	midiEventFilter *contracts.MIDIEventFilter
	coreMIDIConfig  *contracts.CoreMIDIConfig
	mu              sync.Mutex
	capturing       bool
}

// NewMIDIClient initializes a new MIDI client for Darwin with applied options.
func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	client, err := coremidi.NewClient(options.CoreMIDIConfig.ClientName)
	if err != nil {
		return nil, err
	}
	options.Logger.Info("MIDI client successfully created")

	return &ClientMid{
		logger:          options.Logger,
		client:          client,
		midiEventFilter: options.MIDIEventFilter,
		coreMIDIConfig:  options.CoreMIDIConfig,
	}, nil
}

// ListDevices lists available MIDI devices.
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
		sourceEntity := source.Entity()
		devices[i] = contracts.DeviceInfo{
			Name:         source.Name(),
			EntityName:   sourceEntity.Name(),
			Manufacturer: sourceEntity.Manufacturer(),
		}
	}
	return devices, nil
}

// SelectDevice selects a MIDI device by its ID.
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
		m.logger.Field().Int("deviceID", deviceID),
		m.logger.Field().String("deviceName", source.Name()))

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

// handleMIDIMessage processes incoming MIDI messages and applies filtering.
func (m *ClientMid) handleMIDIMessage(source coremidi.Source, packet coremidi.Packet) {
	eventChannel, _ := m.eventChannel.Load().(chan contracts.MIDI)
	if eventChannel == nil {
		m.logger.Warn("eventChannel not initialized or of invalid type")
		return
	}

	if len(packet.Data) >= 3 {
		event := contracts.MIDI{
			Timestamp: uint64(time.Now().UTC().UnixNano()),
			Command:   packet.Data[0],
			Note:      packet.Data[1],
			Velocity:  packet.Data[2],
		}

		if m.midiEventFilter != nil && !isCommandAllowed(event.Command, m.midiEventFilter.Commands) {
			return
		}
		select {
		case eventChannel <- event:
		default:
			m.logger.Warn("Event buffer full; dropping MIDI event")
		}
	} else {
		m.logger.Warn(ErrIncompleteMIDIPacket.Error())
	}
}

// isCommandAllowed checks if a command is in the allowed commands list.
func isCommandAllowed(command byte, allowedCommands []contracts.MIDICommand) bool {
	for _, allowedCommand := range allowedCommands {
		if command == byte(allowedCommand) {
			return true
		}
	}
	return false
}

// StartCapture initializes MIDI event capturing.
func (m *ClientMid) StartCapture(eventChannel chan contracts.MIDI) {
	if eventChannel == nil {
		m.logger.Error("StartCapture called with nil eventChannel")
		return
	}

	// Check if capture is already started
	if m.capturing {
		m.logger.Warn("Capture already started; attempting to stop existing capture")
		if err := m.Stop(); err != nil {
			m.logger.Error("Failed to stop existing capture", m.logger.Field().Error("error", err))
		}
	}

	m.logger.Info("Starting MIDI event capture")
	m.eventChannel.Store(eventChannel)
	m.capturing = true // Atualiza o estado para capturando
}

// Stop stops MIDI event capturing and disconnects the device.
func (m *ClientMid) Stop() error {
	m.logger.Info("Stopping MIDI capture")
	return m.stopCapture()
}

// stopCapture stops capturing MIDI events and releases resources.
func (m *ClientMid) stopCapture() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.portConn != nil {
		m.portConn.Disconnect()
		m.portConn = nil
	}

	if m.eventChannel.Load() != nil {
		m.eventChannel.Store(nil)
		m.capturing = false
	}

	m.logger.Info("MIDI capture stopped")
	return nil
}
