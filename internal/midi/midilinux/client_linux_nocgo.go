//go:build linux && !cgo
// +build linux,!cgo

package midilinux

import (
	"context"
	"errors"
	"fmt"

	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

// ErrCGORequired is returned on Linux when CGo is disabled at build time.
var ErrCGORequired = errors.New("Linux MIDI requires CGo: rebuild with CGO_ENABLED=1")

type stubClient struct {
	logger contracts.Logger
}

func NewMIDIClient(options *contracts.ClientOptions) (contracts.ClientMIDI, error) {
	options.Logger.Warn("Linux MIDI stub: CGo was disabled at build time; MIDI is unavailable")
	return &stubClient{logger: options.Logger}, nil
}

func (s *stubClient) ListDevices() ([]contracts.DeviceInfo, error) {
	return nil, fmt.Errorf("%w", ErrCGORequired)
}

func (s *stubClient) SelectDevice(_ int) error {
	return fmt.Errorf("%w", ErrCGORequired)
}

func (s *stubClient) StartCapture(_ context.Context) (<-chan contracts.MIDI, error) {
	return nil, fmt.Errorf("%w", ErrCGORequired)
}

func (s *stubClient) Stop() error { return nil }

func (s *stubClient) WatchDevices(ctx context.Context) (<-chan contracts.DeviceEvent, error) {
	ch := make(chan contracts.DeviceEvent)
	go func() { <-ctx.Done(); close(ch) }()
	return ch, nil
}
