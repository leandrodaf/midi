package midi

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/leandrodaf/midi/v2/internal/midi/mididarwin"
	"github.com/leandrodaf/midi/v2/internal/midi/midilinux"
	"github.com/leandrodaf/midi/v2/internal/midi/midiwindows"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

// ErrUnsupportedOS is returned when the operating system is not supported by the MIDI client.
var ErrUnsupportedOS = errors.New("unsupported operating system")

// clientInitializers maps OS names to corresponding MIDI client initializers.
var clientInitializers = map[string]func(*contracts.ClientOptions) (contracts.ClientMIDI, error){
	"darwin":  mididarwin.NewMIDIClient,  // macOS — CoreMIDI via CGo.
	"windows": midiwindows.NewMIDIClient, // Windows — winmm.dll via syscalls.
	"linux":   midilinux.NewMIDIClient,   // Linux — ALSA raw MIDI via CGo.
}

// NewClient initializes a MIDI client based on the current operating system.
// It supports macOS (Darwin), Windows, and Linux, returning ErrUnsupportedOS if the OS is unsupported.
//
// opts *contracts.ClientOptions: Configuration options for the MIDI client.
//
// Returns:
//   - contracts.ClientMIDI: An instance of the MIDI client.
//   - error: An error if the operating system is unsupported or if initialization fails.
func NewClient(opts *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	if initializer, exists := clientInitializers[runtime.GOOS]; exists {
		return initializer(opts)
	}
	return nil, fmt.Errorf("%w: %s", ErrUnsupportedOS, runtime.GOOS)
}
