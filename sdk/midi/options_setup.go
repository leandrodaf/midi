package midi

import (
	"github.com/leandrodaf/midi/internal/logger"
	"github.com/leandrodaf/midi/sdk/contracts"
)

func applyDefaultOptions(opts ...contracts.Option) (contracts.ClientOptions, error) {
	options := &contracts.ClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Logger == nil {
		options.Logger = logger.NewLogger()
	}
	if !options.LogLevelIsSet() {
		options.LogLevel = contracts.InfoLevel
	}
	if options.CoreMIDIConfig == nil {
		options.CoreMIDIConfig = &contracts.CoreMIDIConfig{ClientName: "GO MIDI Client"}
	}
	if options.ChannelBufferSize <= 0 {
		options.ChannelBufferSize = 100
	}

	options.Logger.SetLevel(options.LogLevel)
	if options.LogDestination != "" {
		options.Logger.SetDestination(options.LogDestination, options.LogFilePath)
	}

	return *options, nil
}
