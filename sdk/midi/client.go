package midi

import (
	"github.com/leandrodaf/midi/sdk/contracts"
)

// NewMIDIClient creates a new MIDI client with the specified options.
// It applies default options and initializes the client.
//
// opts ...contracts.Option: A variadic list of option functions to customize the client configuration.
//
// Returns:
//   - contracts.ClientMIDI: An instance of the MIDI client.
//   - error: An error, if any occurred during the creation of the client.
func NewMIDIClient(opts ...contracts.Option) (contracts.ClientMIDI, error) {
	options, err := applyDefaultOptions(opts...)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(&options)
	if err != nil {
		return nil, err
	}

	return client, nil
}
