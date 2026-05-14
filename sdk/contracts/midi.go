package contracts

import "context"

// MIDI represents a MIDI event.
type MIDI struct {
	Timestamp uint64
	Command   byte
	Note      byte
	Velocity  byte
}

// ClientMIDI defines the interface for MIDI client operations.
type ClientMIDI interface {
	// Stop halts MIDI event capture and releases resources.
	Stop() error
	// ListDevices returns all available MIDI input devices.
	ListDevices() ([]DeviceInfo, error)
	// SelectDevice selects a MIDI device by its index for capture.
	SelectDevice(deviceID int) error
	// StartCapture begins capturing MIDI events. It returns a read-only channel
	// that receives events. The channel is closed when the context is cancelled
	// or Stop() is called. The channel buffer size is controlled by WithChannelBufferSize.
	StartCapture(ctx context.Context) (<-chan MIDI, error)
}
