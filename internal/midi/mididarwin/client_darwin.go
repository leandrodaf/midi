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

// Error definitions for MIDI connection and handling issues.
var (
	ErrNoMIDIDevices        = errors.New("no MIDI devices found")
	ErrInvalidMIDIDevice    = errors.New("invalid MIDI device")
	ErrMIDIConnectionError  = errors.New("error connecting to MIDI device")
	ErrCreateInputPort      = errors.New("error creating input port")
	ErrIncompleteMIDIPacket = errors.New("incomplete MIDI packet")
)

// internalPortConnection is an interface for handling disconnection from a MIDI port.
type internalPortConnection interface {
	Disconnect()
}

// ClientMid manages MIDI operations on Darwin (macOS) systems.
// This struct handles connections to MIDI devices, manages event capturing,
// and ensures safe concurrency handling.
type ClientMid struct {
	logger          contracts.Logger
	eventChannel    atomic.Value               // Atomic storage for the event channel to ensure thread safety.
	client          coremidi.Client            // CoreMIDI client instance for MIDI operations.
	inputPort       coremidi.InputPort         // Input port for receiving MIDI events.
	portConn        internalPortConnection     // Connection to the MIDI port.
	midiEventFilter *contracts.MIDIEventFilter // Filter for specific MIDI events.
	coreMIDIConfig  *contracts.CoreMIDIConfig  // Configuration for MIDI client.
	mu              sync.Mutex                 // Mutex for thread safety on shared resources.
	capturing       bool                       // Indicates if event capturing is currently active.
	wg              sync.WaitGroup             // WaitGroup for managing concurrent MIDI event processing.
	stopOnce        sync.Once                  // Ensures Stop() is executed only once.
}

// NewMIDIClient initializes a new ClientMid for handling MIDI events on macOS.
// Applies logging and configurations based on the provided options.
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

// ListDevices retrieves and returns available MIDI devices.
// If no devices are found, an error is logged and returned.
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

// SelectDevice selects a MIDI device by ID and connects to it.
// If a device is already connected, it disconnects first.
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
// If an event channel is valid and the message meets filter criteria, it is sent to the channel.
// Adds to WaitGroup to ensure safe concurrent processing.
func (m *ClientMid) handleMIDIMessage(source coremidi.Source, packet coremidi.Packet) {
	m.wg.Add(1)
	defer m.wg.Done()

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

// isCommandAllowed verifies if a MIDI command is allowed based on the event filter configuration.
func isCommandAllowed(command byte, allowedCommands []contracts.MIDICommand) bool {
	for _, allowedCommand := range allowedCommands {
		if command == byte(allowedCommand) {
			return true
		}
	}
	return false
}

// StartCapture begins capturing MIDI events by storing the event channel and marking capturing as active.
// Ensures any ongoing capture is stopped before starting a new one.
func (m *ClientMid) StartCapture(eventChannel chan contracts.MIDI) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if eventChannel == nil {
		m.logger.Error("StartCapture called with nil eventChannel")
		return
	}

	if m.capturing {
		m.logger.Warn("Capture already started; attempting to stop existing capture")
		if err := m.Stop(); err != nil {
			m.logger.Error("Failed to stop existing capture", m.logger.Field().Error("error", err))
		}
	}

	m.logger.Info("Starting MIDI event capture")
	m.eventChannel.Store(eventChannel)
	m.capturing = true
}

// Stop halts MIDI event capturing, disconnects from the device, and waits for ongoing processing to complete.
// This function ensures it only executes once, even if called multiple times.
func (m *ClientMid) Stop() error {
	m.stopOnce.Do(func() {
		m.logger.Info("Stopping MIDI capture")
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.capturing {
			m.capturing = false

			if m.portConn != nil {
				m.portConn.Disconnect()
				m.portConn = nil
			}

			// Store a closed dummy channel to prevent further writes and avoid any panic.
			dummyChannel := make(chan contracts.MIDI)
			m.eventChannel.Store(dummyChannel)

			m.logger.Info("MIDI capture stopped")
			m.wg.Wait() // Wait for all ongoing MIDI event processing to complete
		}
	})
	return nil
}
