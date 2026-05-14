package contracts

// ClientOptions holds all configuration for the MIDI client.
type ClientOptions struct {
	Logger            Logger
	LogLevel          LogLevel
	logLevelSet       bool // true when WithLogLevel was called explicitly
	LogDestination    LogDestination
	LogFilePath       string
	MIDIEventFilter   *MIDIEventFilter
	CoreMIDIConfig    *CoreMIDIConfig
	ChannelBufferSize int
}

// LogLevelIsSet reports whether WithLogLevel was explicitly called.
func (o *ClientOptions) LogLevelIsSet() bool {
	return o.logLevelSet
}

// Option is a function that modifies ClientOptions.
type Option func(*ClientOptions)

// WithLogger sets a custom logger.
func WithLogger(l Logger) Option {
	return func(opts *ClientOptions) {
		opts.Logger = l
	}
}

// WithLogLevel sets the minimum logging level.
func WithLogLevel(level LogLevel) Option {
	return func(opts *ClientOptions) {
		opts.LogLevel = level
		opts.logLevelSet = true
	}
}

// WithLogDestination directs log output to console or file.
// When dest is FileLog, filePath must be provided.
func WithLogDestination(dest LogDestination, filePath ...string) Option {
	return func(opts *ClientOptions) {
		opts.LogDestination = dest
		if len(filePath) > 0 {
			opts.LogFilePath = filePath[0]
		}
	}
}

// WithMIDIEventFilter sets a command allowlist. Only listed commands are delivered.
// Pass nil or omit to receive all commands.
func WithMIDIEventFilter(filter MIDIEventFilter) Option {
	return func(opts *ClientOptions) {
		opts.MIDIEventFilter = &filter
	}
}

// WithCoreMIDIConfig sets CoreMIDI-specific configuration (macOS only).
func WithCoreMIDIConfig(config CoreMIDIConfig) Option {
	return func(opts *ClientOptions) {
		opts.CoreMIDIConfig = &config
	}
}

// WithChannelBufferSize sets the buffer size of the event channel returned by StartCapture.
// Default is 100.
func WithChannelBufferSize(n int) Option {
	return func(opts *ClientOptions) {
		opts.ChannelBufferSize = n
	}
}
