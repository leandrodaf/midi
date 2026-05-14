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
	Commands []MIDICommand
}

// CoreMIDIConfig holds configuration for CoreMIDI.
type CoreMIDIConfig struct {
	ClientName string
}
