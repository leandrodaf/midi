package contracts

// MIDI represents a MIDI event with a timestamp, command, note, and velocity.
type MIDI struct {
	Timestamp uint64 // Timestamp indicates the time the event occurred.
	Command   byte   // Command specifies the type of MIDI event (e.g., Note On, Note Off).
	Note      byte   // Note represents the MIDI note number (0-127).
	Velocity  byte   // Velocity indicates the strength of the note being played (0-127).
}

// ClientMIDI defines an interface for MIDI client operations.
type ClientMIDI interface {
	Stop() error                         // Stops the MIDI client and releases resources.
	ListDevices() ([]DeviceInfo, error)  // Lists all available MIDI devices.
	SelectDevice(deviceID int) error     // Selects a MIDI device by its ID for communication.
	StartCapture(eventChannel chan MIDI) // Starts capturing MIDI events and sends them to the specified channel.
}
