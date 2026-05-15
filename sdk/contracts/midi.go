package contracts

import "context"

// MIDI represents a MIDI event.
type MIDI struct {
	Timestamp uint64
	Command   byte
	Note      byte
	Velocity  byte
}

// DeviceEventType describes what happened to a MIDI device.
type DeviceEventType int

const (
	// DeviceAdded is sent when a new MIDI device becomes available.
	DeviceAdded DeviceEventType = iota
	// DeviceRemoved is sent when a MIDI device is disconnected or deactivated.
	DeviceRemoved
)

// DeviceEvent is emitted by WatchDevices when the set of MIDI devices changes.
type DeviceEvent struct {
	Type   DeviceEventType
	Device DeviceInfo
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
	// WatchDevices returns a channel that emits DeviceEvent values whenever a
	// MIDI device is connected or disconnected. The channel is closed when ctx
	// is cancelled. Implementations may use OS-level notifications (macOS
	// CoreMIDI) or periodic polling (Linux, Windows).
	WatchDevices(ctx context.Context) (<-chan DeviceEvent, error)
}
