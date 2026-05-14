// Package logger exposes the default structured logger for consumers of the
// MIDI library. Inject it via contracts.WithLogger, or omit it to have the
// client create one automatically.
package logger

import (
	internal "github.com/leandrodaf/midi/internal/logger"
	"github.com/leandrodaf/midi/sdk/contracts"
)

// NewLogger returns a structured logger that writes to stderr at InfoLevel.
// Fields are appended as a JSON object at the end of each line.
func NewLogger() contracts.Logger {
	return internal.NewLogger()
}
