package contracts

// MIDICommand represents the types of MIDI commands for event filtering.
type MIDICommand byte

const (
	// NoteOn is the MIDI command for a Note On event (0x90).
	NoteOn MIDICommand = 0x90
	// NoteOff is the MIDI command for a Note Off event (0x80).
	NoteOff MIDICommand = 0x80
)

// MIDIEventFilter allows users to specify which MIDI commands to capture.
type MIDIEventFilter struct {
	Commands []MIDICommand // List of MIDI commands to filter.
}

// CoreMIDIConfig holds configuration for CoreMIDI.
type CoreMIDIConfig struct {
	ClientName string // Name of the MIDI client.
}

// ClientOptions defines the configuration options for the MIDI client.
type ClientOptions struct {
	Logger          Logger           // Logger for logging events and errors.
	LogLevel        LogLevel         // Level of logging to use.
	LogFilePath     string           // File path for logging if file logging is enabled.
	MIDIEventFilter *MIDIEventFilter // Optional filter for MIDI events to capture.
	CoreMIDIConfig  *CoreMIDIConfig  // Configuration specific to CoreMIDI.
}

// Option is a function that modifies ClientOptions.
type Option func(*ClientOptions)

// WithLogger sets the logger for the MIDI client.
func WithLogger(l Logger) Option {
	return func(opts *ClientOptions) {
		opts.Logger = l
	}
}

// WithLogLevel sets the logging level for the MIDI client.
func WithLogLevel(level LogLevel) Option {
	return func(opts *ClientOptions) {
		opts.LogLevel = level
	}
}

// WithMIDIEventFilter sets the MIDI event filter for the MIDI client.
func WithMIDIEventFilter(filter MIDIEventFilter) Option {
	return func(opts *ClientOptions) {
		opts.MIDIEventFilter = &filter
	}
}

// WithCoreMIDIConfig sets the CoreMIDI configuration for the MIDI client.
func WithCoreMIDIConfig(config CoreMIDIConfig) Option {
	return func(opts *ClientOptions) {
		opts.CoreMIDIConfig = &config
	}
}
