package contracts_test

import (
	"testing"

	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

func TestIsCommandAllowed_NilFilter(t *testing.T) {
	if !contracts.IsCommandAllowed(byte(contracts.NoteOn), nil) {
		t.Errorf("expected nil filter to allow all commands")
	}
}

func TestIsCommandAllowed_EmptyCommands(t *testing.T) {
	filter := &contracts.MIDIEventFilter{Commands: []contracts.MIDICommand{}}

	if contracts.IsCommandAllowed(byte(contracts.NoteOn), filter) {
		t.Errorf("expected empty command list to block all commands")
	}
}

func TestIsCommandAllowed_CommandPresent(t *testing.T) {
	filter := &contracts.MIDIEventFilter{Commands: []contracts.MIDICommand{contracts.NoteOn}}

	if !contracts.IsCommandAllowed(byte(contracts.NoteOn), filter) {
		t.Errorf("expected listed command to be allowed")
	}
}

func TestIsCommandAllowed_CommandAbsent(t *testing.T) {
	filter := &contracts.MIDIEventFilter{Commands: []contracts.MIDICommand{contracts.NoteOff}}

	if contracts.IsCommandAllowed(byte(contracts.NoteOn), filter) {
		t.Errorf("expected unlisted command to be blocked")
	}
}

func TestIsCommandAllowed_MultipleCommands(t *testing.T) {
	filter := &contracts.MIDIEventFilter{Commands: []contracts.MIDICommand{contracts.NoteOff, contracts.NoteOn}}

	if !contracts.IsCommandAllowed(byte(contracts.NoteOn), filter) {
		t.Errorf("expected command to match one of multiple allowed commands")
	}
}
