package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/leandrodaf/midi/sdk/contracts"
)

type stdLogger struct {
	level  contracts.LogLevel
	output io.Writer
}

// NewLogger creates a new structured logger writing to stderr.
func NewLogger() contracts.Logger {
	return NewLoggerWithWriter(os.Stderr)
}

// NewLoggerWithWriter creates a new structured logger writing to the provided writer.
func NewLoggerWithWriter(w io.Writer) contracts.Logger {
	if w == nil {
		w = os.Stderr
	}
	return &stdLogger{
		level:  contracts.InfoLevel,
		output: w,
	}
}

// NewZapLogger is an alias for NewLogger kept for compatibility.
func NewZapLogger() contracts.Logger { return NewLogger() }

// NewStandardLogger is an alias for NewLogger kept for compatibility.
func NewStandardLogger() contracts.Logger { return NewLogger() }

func (l *stdLogger) Info(msg string, fields ...contracts.Field) {
	l.emit(contracts.InfoLevel, "INFO", msg, fields...)
}

func (l *stdLogger) Error(msg string, fields ...contracts.Field) {
	l.emit(contracts.ErrorLevel, "ERROR", msg, fields...)
}

func (l *stdLogger) Debug(msg string, fields ...contracts.Field) {
	l.emit(contracts.DebugLevel, "DEBUG", msg, fields...)
}

func (l *stdLogger) Warn(msg string, fields ...contracts.Field) {
	l.emit(contracts.WarnLevel, "WARN", msg, fields...)
}

func (l *stdLogger) Fatal(msg string, fields ...contracts.Field) {
	l.emit(contracts.FatalLevel, "FATAL", msg, fields...)
	os.Exit(1)
}

func (l *stdLogger) SetLevel(level contracts.LogLevel) {
	l.level = level
}

func (l *stdLogger) SetDestination(dest contracts.LogDestination, filePath ...string) {
	if dest == contracts.FileLog && len(filePath) > 0 {
		f, err := os.OpenFile(filePath[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			l.output = f
		}
	}
}

var levelOrder = map[contracts.LogLevel]int{
	contracts.DebugLevel: 0,
	contracts.InfoLevel:  1,
	contracts.WarnLevel:  2,
	contracts.ErrorLevel: 3,
	contracts.FatalLevel: 4,
}

func (l *stdLogger) shouldLog(level contracts.LogLevel) bool {
	return levelOrder[level] >= levelOrder[l.level]
}

func (l *stdLogger) emit(level contracts.LogLevel, levelStr, msg string, fields ...contracts.Field) {
	if !l.shouldLog(level) {
		return
	}
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file, line = "unknown", 0
	} else {
		file = filepath.Base(file)
	}
	timestamp := time.Now().UTC().Format(time.RFC3339)
	suffix := formatFields(fields...)
	fmt.Fprintf(l.output, "%s [%s] %s:%d: %s%s\n", timestamp, levelStr, file, line, msg, suffix)
}

func formatFields(fields ...contracts.Field) string {
	if len(fields) == 0 {
		return ""
	}
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		if f.Key != "" {
			m[f.Key] = f.Value
		}
	}
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf(" [failed to format fields: %v]", err)
	}
	return " " + string(b)
}
