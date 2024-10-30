package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/leandrodaf/midi/sdk/contracts"
)

// StandardLogger is an optimized implementation of contracts.Logger
type StandardLogger struct {
	mu             sync.RWMutex
	logLevel       contracts.LogLevel
	dest           contracts.LogDestination
	file           *os.File
	externalLogURL string
	httpClient     *http.Client
}

// NewStandardLogger creates a new logger for console output
func NewStandardLogger() contracts.Logger {
	return &StandardLogger{
		logLevel:   contracts.InfoLevel,
		dest:       contracts.ConsoleLog,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// NewFileLogger creates a new logger for file output
func NewFileLogger(file *os.File) contracts.Logger {
	return &StandardLogger{
		logLevel:   contracts.InfoLevel,
		dest:       contracts.FileLog,
		file:       file,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Info logs a message at the INFO level
func (s *StandardLogger) Info(msg string, fields ...contracts.Field) {
	s.log(contracts.InfoLevel, "INFO", msg, fields...)
}

// Error logs a message at the ERROR level
func (s *StandardLogger) Error(msg string, fields ...contracts.Field) {
	s.log(contracts.ErrorLevel, "ERROR", msg, fields...)
}

// Debug logs a message at the DEBUG level
func (s *StandardLogger) Debug(msg string, fields ...contracts.Field) {
	s.log(contracts.DebugLevel, "DEBUG", msg, fields...)
}

// Warn logs a message at the WARN level
func (s *StandardLogger) Warn(msg string, fields ...contracts.Field) {
	s.log(contracts.WarnLevel, "WARN", msg, fields...)
}

// Fatal logs a message at the FATAL level and terminates the application
func (s *StandardLogger) Fatal(msg string, fields ...contracts.Field) {
	s.log(contracts.FatalLevel, "FATAL", msg, fields...)
	os.Exit(1)
}

// Field returns a new instance of Field
func (s *StandardLogger) Field() contracts.Field {
	return &simpleField{}
}

// SetLevel sets the logging level
func (s *StandardLogger) SetLevel(level contracts.LogLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logLevel = level
}

// SetDestination sets the logging destination
func (s *StandardLogger) SetDestination(dest contracts.LogDestination, filePath ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close the previous file if it exists
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}

	switch dest {
	case contracts.ConsoleLog:
		s.dest = dest
	case contracts.FileLog:
		if len(filePath) > 0 {
			file, err := os.OpenFile(filePath[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// Handle the error internally
				fmt.Fprintf(os.Stderr, "ERROR: Failed to open log file: %v\n", err)
			} else {
				s.file = file
				s.dest = dest
			}
		} else {
			fmt.Fprintln(os.Stderr, "ERROR: File path must be provided for FileLog")
		}
	default:
		fmt.Fprintln(os.Stderr, "ERROR: Unknown logging destination")
	}
}

// simpleField implements contracts.Field
type simpleField struct {
	key   string
	value interface{}
}

func (f *simpleField) Bool(key string, val bool) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Int(key string, val int) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Float64(key string, val float64) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) String(key string, val string) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Time(key string, val time.Time) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Int64(key string, val int64) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Error(key string, val error) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Uint64(key string, val uint64) contracts.Field {
	return &simpleField{key, val}
}

func (f *simpleField) Uint8(key string, val uint8) contracts.Field {
	return &simpleField{key, val}
}

// log is the internal function for logging messages
func (s *StandardLogger) log(level contracts.LogLevel, levelStr, msg string, fields ...contracts.Field) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.logLevel > level {
		return
	}

	// Capture the name of the file and the line number where the log was called
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	formattedFields := formatFields(fields...)
	logMessage := fmt.Sprintf("%s [%s] %s:%d: %s%s", timestamp, levelStr, file, line, msg, formattedFields)

	switch s.dest {
	case contracts.ConsoleLog:
		fmt.Println(logMessage)
	case contracts.FileLog:
		if s.file != nil {
			fmt.Fprintln(s.file, logMessage)
		} else {
			fmt.Fprintln(os.Stderr, "ERROR: Log file is not configured.")
		}
	}
}

// formatFields formats additional fields
func formatFields(fields ...contracts.Field) string {
	if len(fields) == 0 {
		return ""
	}

	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		if field != nil {
			if f, ok := field.(*simpleField); ok {
				fieldMap[f.key] = f.value
			}
		}
	}

	if len(fieldMap) == 0 {
		return ""
	}

	jsonBytes, err := json.Marshal(fieldMap)
	if err != nil {
		return fmt.Sprintf(" [failed to format fields: %v]", err)
	}

	return " " + string(jsonBytes)
}
