package midi

import (
	"github.com/leandrodaf/midi/internal/logger"
	"github.com/leandrodaf/midi/sdk/contracts"
)

// applyDefaultOptions sets default values for ClientOptions if not explicitly provided.
//
// opts ...contracts.Option: A variadic list of option functions that can modify ClientOptions.
//
// Returns:
//   - contracts.ClientOptions: A structure containing the finalized client options with defaults applied.
//   - error: An error if there was an issue applying the options.
func applyDefaultOptions(opts ...contracts.Option) (contracts.ClientOptions, error) {
	options := &contracts.ClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Set defaults if options are not provided
	if options.Logger == nil {
		options.Logger = logger.NewZapLogger() // Default to a standard logger
	}
	if options.LogLevel == 0 {
		options.LogLevel = contracts.InfoLevel // Default log level to InfoLevel
	}

	if options.CoreMIDIConfig == nil {
		options.CoreMIDIConfig = &contracts.CoreMIDIConfig{ClientName: "GO MIDI Client"} // Default CoreMIDI config
	}

	options.Logger.SetLevel(options.LogLevel) // Set the logger to the specified log level
	return *options, nil
}
