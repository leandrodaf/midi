package midi

import (
	"testing"

	"github.com/leandrodaf/midi/sdk/contracts"
)

type stubLogger struct{}

func (l *stubLogger) Info(msg string, fields ...contracts.Field)  {}
func (l *stubLogger) Error(msg string, fields ...contracts.Field) {}
func (l *stubLogger) Debug(msg string, fields ...contracts.Field) {}
func (l *stubLogger) Warn(msg string, fields ...contracts.Field)  {}
func (l *stubLogger) Fatal(msg string, fields ...contracts.Field) {}
func (l *stubLogger) SetLevel(level contracts.LogLevel)           {}
func (l *stubLogger) SetDestination(dest contracts.LogDestination, filePath ...string) {
}

func TestApplyDefaultOptions_DefaultLogger(t *testing.T) {
	options, err := applyDefaultOptions()
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.Logger == nil {
		t.Errorf("expected default logger to be set")
	}
}

func TestApplyDefaultOptions_DefaultLogLevel(t *testing.T) {
	options, err := applyDefaultOptions()
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.LogLevel != contracts.InfoLevel {
		t.Errorf("expected default log level %v, got %v", contracts.InfoLevel, options.LogLevel)
	}
}

func TestApplyDefaultOptions_DefaultChannelBufferSize(t *testing.T) {
	options, err := applyDefaultOptions()
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.ChannelBufferSize != 100 {
		t.Errorf("expected default channel buffer size 100, got %d", options.ChannelBufferSize)
	}
}

func TestApplyDefaultOptions_DefaultCoreMIDIConfig(t *testing.T) {
	options, err := applyDefaultOptions()
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.CoreMIDIConfig == nil {
		t.Errorf("expected default CoreMIDIConfig to be set")
	}
}

func TestApplyDefaultOptions_WithLogLevel(t *testing.T) {
	options, err := applyDefaultOptions(contracts.WithLogLevel(contracts.DebugLevel))
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.LogLevel != contracts.DebugLevel {
		t.Errorf("expected log level %v, got %v", contracts.DebugLevel, options.LogLevel)
	}
	if !options.LogLevelIsSet() {
		t.Errorf("expected LogLevelIsSet to be true when WithLogLevel is used")
	}
}

func TestApplyDefaultOptions_PreservesCustomLogger(t *testing.T) {
	customLogger := &stubLogger{}

	options, err := applyDefaultOptions(contracts.WithLogger(customLogger))
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.Logger != customLogger {
		t.Errorf("expected custom logger to be preserved")
	}
}

func TestApplyDefaultOptions_WithChannelBufferSize(t *testing.T) {
	options, err := applyDefaultOptions(contracts.WithChannelBufferSize(50))
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.ChannelBufferSize != 50 {
		t.Errorf("expected channel buffer size 50, got %d", options.ChannelBufferSize)
	}
}

func TestApplyDefaultOptions_LogLevelNotSetByDefault(t *testing.T) {
	options, err := applyDefaultOptions()
	if err != nil {
		t.Fatalf("applyDefaultOptions returned error: %v", err)
	}
	if options.LogLevelIsSet() {
		t.Errorf("expected LogLevelIsSet to be false by default")
	}
}
