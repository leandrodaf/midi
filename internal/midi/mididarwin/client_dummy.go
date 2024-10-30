//go:build !darwin
// +build !darwin

package mididarwin

import (
	"fmt"

	"github.com/leandrodaf/midi/sdk/contracts"
)

type DummyMIDIClient struct {
	logger contracts.Logger
}

func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Info("Using dummy MIDI client for non-macOS system")
	return &DummyMIDIClient{
		logger: options.Logger,
	}, nil
}

func (m *DummyMIDIClient) ListDevices() ([]contracts.DeviceInfo, error) {
	m.logger.Warn("ListDevices called on dummy MIDI client")
	return nil, fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *DummyMIDIClient) SelectDevice(deviceID int) error {
	m.logger.Warn("SelectDevice called on dummy MIDI client")
	return fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *DummyMIDIClient) StartCapture(eventChannel chan contracts.MIDI) {
	m.logger.Warn("StartCapture called on dummy MIDI client")
}

func (m *DummyMIDIClient) Stop() error {
	m.logger.Warn("Stop called on dummy MIDI client")
	return nil
}
