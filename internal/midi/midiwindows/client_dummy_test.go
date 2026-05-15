//go:build !windows
// +build !windows

package midiwindows_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/leandrodaf/midi/v2/internal/logger"
	"github.com/leandrodaf/midi/v2/internal/midi/midiwindows"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

func newDummyWindowsClient(t *testing.T) contracts.ClientMIDI {
	t.Helper()

	client, err := midiwindows.NewMIDIClient(&contracts.ClientOptions{Logger: logger.NewLoggerWithWriter(io.Discard)})
	if err != nil {
		t.Fatalf("NewMIDIClient returned error: %v", err)
	}

	return client
}

func TestDummyClient_ListDevices_ReturnsError(t *testing.T) {
	client := newDummyWindowsClient(t)

	devices, err := client.ListDevices()
	if err == nil {
		t.Fatalf("expected ListDevices to return an error")
	}
	if devices != nil {
		t.Errorf("expected ListDevices to return nil devices, got %#v", devices)
	}
}

func TestDummyClient_SelectDevice_ReturnsError(t *testing.T) {
	client := newDummyWindowsClient(t)

	err := client.SelectDevice(0)
	if err == nil {
		t.Fatalf("expected SelectDevice to return an error")
	}
}

func TestDummyClient_StartCapture_ReturnsError(t *testing.T) {
	client := newDummyWindowsClient(t)

	channel, err := client.StartCapture(context.Background())
	if err == nil {
		t.Fatalf("expected StartCapture to return an error")
	}
	if channel != nil {
		t.Errorf("expected StartCapture to return a nil channel")
	}
}

func TestDummyClient_Stop_ReturnsNil(t *testing.T) {
	client := newDummyWindowsClient(t)

	if err := client.Stop(); err != nil {
		t.Errorf("expected Stop to return nil, got %v", err)
	}
}

func TestDummyWindowsClient_WatchDevices_ClosesOnCancel(t *testing.T) {
	client := newDummyWindowsClient(t)
	ctx, cancel := context.WithCancel(context.Background())

	ch, err := client.WatchDevices(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Errorf("expected channel to be closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out: channel not closed after context cancel")
	}
}

func TestDummyWindowsClient_WatchDevices_NoEventsEmitted(t *testing.T) {
	client := newDummyWindowsClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := client.WatchDevices(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case ev, ok := <-ch:
		if ok {
			t.Errorf("expected no events from stub, got %+v", ev)
		}
	case <-time.After(20 * time.Millisecond):
		// good — no events emitted
	}
}
