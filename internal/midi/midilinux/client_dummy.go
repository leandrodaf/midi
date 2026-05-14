//go:build !linux
// +build !linux

package midilinux

import (
	"context"
	"fmt"

	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

type dummyMIDIClient struct {
	logger contracts.Logger
}

func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Info("Using dummy MIDI client for non-Linux system")
	return &dummyMIDIClient{logger: options.Logger}, nil
}

func (m *dummyMIDIClient) ListDevices() ([]contracts.DeviceInfo, error) {
	m.logger.Warn("ListDevices called on dummy MIDI client")
	return nil, fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *dummyMIDIClient) SelectDevice(_ int) error {
	m.logger.Warn("SelectDevice called on dummy MIDI client")
	return fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *dummyMIDIClient) StartCapture(_ context.Context) (<-chan contracts.MIDI, error) {
	m.logger.Warn("StartCapture called on dummy MIDI client")
	return nil, fmt.Errorf("MIDI functionality is not available on this platform")
}

func (m *dummyMIDIClient) Stop() error {
	m.logger.Warn("Stop called on dummy MIDI client")
	return nil
}
