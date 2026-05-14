//go:build !linux
// +build !linux

package midilinux_test

import (
	"context"
	"io"
	"testing"

	"github.com/leandrodaf/midi/v2/internal/logger"
	"github.com/leandrodaf/midi/v2/internal/midi/midilinux"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

func newDummyLinuxClient(t *testing.T) contracts.ClientMIDI {
	t.Helper()
	client, err := midilinux.NewMIDIClient(&contracts.ClientOptions{
		Logger: logger.NewLoggerWithWriter(io.Discard),
	})
	if err != nil {
		t.Fatalf("NewMIDIClient returned unexpected error: %v", err)
	}
	return client
}

func TestDummyLinuxClient_ListDevices_ReturnsError(t *testing.T) {
	client := newDummyLinuxClient(t)
	devices, err := client.ListDevices()
	if err == nil {
		t.Fatal("expected ListDevices to return an error on non-Linux")
	}
	if devices != nil {
		t.Errorf("expected nil devices, got %#v", devices)
	}
}

func TestDummyLinuxClient_SelectDevice_ReturnsError(t *testing.T) {
	client := newDummyLinuxClient(t)
	if err := client.SelectDevice(0); err == nil {
		t.Fatal("expected SelectDevice to return an error on non-Linux")
	}
}

func TestDummyLinuxClient_StartCapture_ReturnsError(t *testing.T) {
	client := newDummyLinuxClient(t)
	ch, err := client.StartCapture(context.Background())
	if err == nil {
		t.Fatal("expected StartCapture to return an error on non-Linux")
	}
	if ch != nil {
		t.Errorf("expected nil channel, got %v", ch)
	}
}

func TestDummyLinuxClient_Stop_ReturnsNil(t *testing.T) {
	client := newDummyLinuxClient(t)
	if err := client.Stop(); err != nil {
		t.Errorf("expected Stop to return nil, got %v", err)
	}
}
