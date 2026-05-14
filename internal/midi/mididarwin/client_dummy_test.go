//go:build !darwin
// +build !darwin

package mididarwin_test

import (
	"context"
	"io"
	"testing"

	"github.com/leandrodaf/midi/v2/internal/logger"
	"github.com/leandrodaf/midi/v2/internal/midi/mididarwin"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

func newDummyDarwinClient(t *testing.T) contracts.ClientMIDI {
	t.Helper()

	client, err := mididarwin.NewMIDIClient(&contracts.ClientOptions{Logger: logger.NewLoggerWithWriter(io.Discard)})
	if err != nil {
		t.Fatalf("NewMIDIClient returned error: %v", err)
	}

	return client
}

func TestDummyClient_ListDevices_ReturnsError(t *testing.T) {
	client := newDummyDarwinClient(t)

	devices, err := client.ListDevices()
	if err == nil {
		t.Fatalf("expected ListDevices to return an error")
	}
	if devices != nil {
		t.Errorf("expected ListDevices to return nil devices, got %#v", devices)
	}
}

func TestDummyClient_SelectDevice_ReturnsError(t *testing.T) {
	client := newDummyDarwinClient(t)

	err := client.SelectDevice(0)
	if err == nil {
		t.Fatalf("expected SelectDevice to return an error")
	}
}

func TestDummyClient_StartCapture_ReturnsError(t *testing.T) {
	client := newDummyDarwinClient(t)

	channel, err := client.StartCapture(context.Background())
	if err == nil {
		t.Fatalf("expected StartCapture to return an error")
	}
	if channel != nil {
		t.Errorf("expected StartCapture to return a nil channel")
	}
}

func TestDummyClient_Stop_ReturnsNil(t *testing.T) {
	client := newDummyDarwinClient(t)

	if err := client.Stop(); err != nil {
		t.Errorf("expected Stop to return nil, got %v", err)
	}
}
