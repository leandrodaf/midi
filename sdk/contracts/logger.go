package contracts

import "time"

// LogLevel represents the severity level for logging.
type LogLevel int

const (
	// InfoLevel indicates informational messages that highlight the progress of the application.
	InfoLevel LogLevel = iota
	// DebugLevel indicates debug messages that are useful for developers to troubleshoot issues.
	DebugLevel
	// ErrorLevel indicates error messages that represent serious issues that need attention.
	ErrorLevel
	// WarnLevel indicates potentially harmful situations that should be monitored.
	WarnLevel
	// FatalLevel indicates very severe error events that will presumably lead the application to abort.
	FatalLevel
)

// LogDestination specifies where the log messages should be directed.
type LogDestination string

const (
	// ConsoleLog directs log messages to the console output.
	ConsoleLog LogDestination = "console"
	// FileLog directs log messages to a file.
	FileLog LogDestination = "file"
)

// Field representa um campo de log com vários tipos de dados.
type Field interface {
	Bool(key string, val bool) Field
	Int(key string, val int) Field
	Float64(key string, val float64) Field
	String(key string, val string) Field
	Time(key string, val time.Time) Field
	Int64(key string, val int64) Field
	Error(key string, val error) Field
	Uint64(key string, val uint64) Field
	Uint8(key string, val uint8) Field
}

// Logger fornece métodos para registrar mensagens em diferentes níveis.
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	Field() Field

	SetLevel(level LogLevel)
	SetDestination(dest LogDestination, filePath ...string)
}
