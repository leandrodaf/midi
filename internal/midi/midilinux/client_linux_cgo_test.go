//go:build linux && cgo
// +build linux,cgo

package midilinux_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/leandrodaf/midi/v2/internal/logger"
	"github.com/leandrodaf/midi/v2/internal/midi/midilinux"
	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

func newLinuxClient(t *testing.T) contracts.ClientMIDI {
	t.Helper()
	client, err := midilinux.NewMIDIClient(&contracts.ClientOptions{
		Logger:            logger.NewLoggerWithWriter(io.Discard),
		ChannelBufferSize: 16,
	})
	if err != nil {
		t.Fatalf("NewMIDIClient returned unexpected error: %v", err)
	}
	return client
}

func TestLinuxClient_WatchDevices_ClosesOnCancel(t *testing.T) {
	client := newLinuxClient(t)
	ctx, cancel := context.WithCancel(context.Background())

	ch, err := client.WatchDevices(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Errorf("expected channel to be closed after context cancel")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out: channel not closed after context cancel")
	}
}

func TestLinuxClient_WatchDevices_ReturnsChannel(t *testing.T) {
	client := newLinuxClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := client.WatchDevices(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch == nil {
		t.Fatal("expected non-nil channel")
	}
}

func TestDiffDevices_AddsNewDevice(t *testing.T) {
	prev := []contracts.DeviceInfo{}
	curr := []contracts.DeviceInfo{{Name: "Piano", Manufacturer: "Yamaha"}}
	evCh := make(chan contracts.DeviceEvent, 8)

	midilinux.DiffDevicesExported(prev, curr, evCh)

	select {
	case ev := <-evCh:
		if ev.Type != contracts.DeviceAdded {
			t.Errorf("expected DeviceAdded, got %v", ev.Type)
		}
		if ev.Device.Name != "Piano" {
			t.Errorf("unexpected device name: %s", ev.Device.Name)
		}
	default:
		t.Fatal("expected a DeviceAdded event")
	}
}

func TestDiffDevices_RemovesGoneDevice(t *testing.T) {
	prev := []contracts.DeviceInfo{{Name: "Piano", Manufacturer: "Yamaha"}}
	curr := []contracts.DeviceInfo{}
	evCh := make(chan contracts.DeviceEvent, 8)

	midilinux.DiffDevicesExported(prev, curr, evCh)

	select {
	case ev := <-evCh:
		if ev.Type != contracts.DeviceRemoved {
			t.Errorf("expected DeviceRemoved, got %v", ev.Type)
		}
	default:
		t.Fatal("expected a DeviceRemoved event")
	}
}

func TestDiffDevices_NoChangeNoEvents(t *testing.T) {
	dev := contracts.DeviceInfo{Name: "Piano", Manufacturer: "Yamaha"}
	prev := []contracts.DeviceInfo{dev}
	curr := []contracts.DeviceInfo{dev}
	evCh := make(chan contracts.DeviceEvent, 8)

	midilinux.DiffDevicesExported(prev, curr, evCh)

	select {
	case ev := <-evCh:
		t.Errorf("expected no events, got %+v", ev)
	default:
		// good
	}
}
