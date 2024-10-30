//go:build !windows
// +build !windows

package midiwindows

import (
	"fmt"

	"github.com/leandrodaf/midi/sdk/contracts"
)

type dummyMIDIClient struct {
	logger contracts.Logger
}

// NewMIDIClient initializes a dummy MIDI client for non-Windows systems.
func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Info("Using dummy MIDI client for non-Windows system")
	return &dummyMIDIClient{
		logger: options.Logger,
	}, nil
}

// ListDevices logs a warning and returns an error indicating that MIDI functionality is unavailable on this platform.
func (m *dummyMIDIClient) ListDevices() ([]contracts.DeviceInfo, error) {
	m.logger.Warn("ListDevices called on dummy MIDI client")
	return nil, fmt.Errorf("MIDI functionality is not available on this platform")
}

// SelectDevice logs a warning and returns an error indicating that MIDI functionality is unavailable on this platform.
func (m *dummyMIDIClient) SelectDevice(deviceID int) error {
	m.logger.Warn("SelectDevice called on dummy MIDI client")
	return fmt.Errorf("MIDI functionality is not available on this platform")
}

// StartCapture logs a warning indicating that StartCapture was called on the dummy MIDI client.
func (m *dummyMIDIClient) StartCapture(eventChannel chan contracts.MIDI) {
	m.logger.Warn("StartCapture called on dummy MIDI client")
}

// Stop logs a warning indicating that Stop was called on the dummy MIDI client.
func (m *dummyMIDIClient) Stop() error {
	m.logger.Warn("Stop called on dummy MIDI client")
	return nil
}
