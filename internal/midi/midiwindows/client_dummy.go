//go:build !windows
// +build !windows

package midiwindows

import (
	"context"
	"fmt"

	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

// dummyMIDIClient is the no-op ClientMIDI used on non-Windows systems when the
// midiwindows package is selected by the build system.
type dummyMIDIClient struct {
	logger contracts.Logger
}

func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Info("Using dummy MIDI client for non-Windows system")
	return &dummyMIDIClient{logger: options.Logger}, nil
}

func (m *dummyMIDIClient) ListDevices() ([]contracts.DeviceInfo, error) {
	m.logger.Warn("ListDevices called on dummy MIDI client")
	return nil, fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *dummyMIDIClient) SelectDevice(deviceID int) error {
	m.logger.Warn("SelectDevice called on dummy MIDI client")
	return fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *dummyMIDIClient) StartCapture(ctx context.Context) (<-chan contracts.MIDI, error) {
	m.logger.Warn("StartCapture called on dummy MIDI client")
	return nil, fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *dummyMIDIClient) Stop() error {
	m.logger.Warn("Stop called on dummy MIDI client")
	return nil
}

// WatchDevices returns a channel that is closed when ctx is cancelled.
// No device events are ever emitted — this is a no-op stub.
func (m *dummyMIDIClient) WatchDevices(ctx context.Context) (<-chan contracts.DeviceEvent, error) {
	ch := make(chan contracts.DeviceEvent)
	go func() { <-ctx.Done(); close(ch) }()
	return ch, nil
}
