package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leandrodaf/midi/internal/logger"
	"github.com/leandrodaf/midi/sdk/contracts"
)

func TestNewLogger_NotNil(t *testing.T) {
	if logger.NewLogger() == nil {
		t.Errorf("expected NewLogger to return a non-nil logger")
	}
}

func TestLogger_DebugSuppressedAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	log.Debug("debug hidden")

	if buf.Len() != 0 {
		t.Errorf("expected debug output to be suppressed at info level, got %q", buf.String())
	}
}

func TestLogger_InfoShownAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	log.Info("info visible")

	if !strings.Contains(buf.String(), "info visible") {
		t.Errorf("expected info message to be written, got %q", buf.String())
	}
}

func TestLogger_WarnShownAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	log.Warn("warn visible")

	if !strings.Contains(buf.String(), "warn visible") {
		t.Errorf("expected warn message to be written, got %q", buf.String())
	}
}

func TestLogger_ErrorShownAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	log.Error("error visible")

	if !strings.Contains(buf.String(), "error visible") {
		t.Errorf("expected error message to be written, got %q", buf.String())
	}
}

func TestLogger_DebugShownAtDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	log.SetLevel(contracts.DebugLevel)

	log.Debug("debug visible")

	if !strings.Contains(buf.String(), "debug visible") {
		t.Errorf("expected debug message to be written at debug level, got %q", buf.String())
	}
}

func TestLogger_InfoSuppressedAtErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)
	log.SetLevel(contracts.ErrorLevel)

	log.Info("info hidden")

	if buf.Len() != 0 {
		t.Errorf("expected info output to be suppressed at error level, got %q", buf.String())
	}
}

func TestLogger_FieldsAppearedInOutput(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	log.Info("with fields", contracts.StringField("key", "value"))

	if !strings.Contains(buf.String(), "\"key\"") {
		t.Errorf("expected output to contain field key, got %q", buf.String())
	}
}

func TestLogger_EmptyFieldsNoSuffix(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLoggerWithWriter(&buf)

	log.Info("plain message")

	output := buf.String()
	if !strings.Contains(output, "plain message") {
		t.Fatalf("expected output to contain message, got %q", output)
	}
	if strings.Contains(output, "{") {
		t.Errorf("expected output without fields to omit JSON suffix, got %q", output)
	}
}
