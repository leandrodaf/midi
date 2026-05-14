package contracts

import "time"

// LogLevel represents the severity level for logging.
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// LogDestination specifies where log messages are directed.
type LogDestination string

const (
	ConsoleLog LogDestination = "console"
	FileLog    LogDestination = "file"
)

// Field is a structured log field holding a key-value pair.
type Field struct {
	Key   string
	Value any
}

// Standalone Field constructors — use these instead of logger.Field().XXX().
func BoolField(key string, val bool) Field       { return Field{key, val} }
func IntField(key string, val int) Field         { return Field{key, val} }
func Float64Field(key string, val float64) Field { return Field{key, val} }
func StringField(key string, val string) Field   { return Field{key, val} }
func TimeField(key string, val time.Time) Field  { return Field{key, val} }
func Int64Field(key string, val int64) Field     { return Field{key, val} }
func ErrField(key string, val error) Field       { return Field{key, val} }
func Uint64Field(key string, val uint64) Field   { return Field{key, val} }
func Uint8Field(key string, val uint8) Field     { return Field{key, val} }

// Logger provides methods for structured logging.
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	SetLevel(level LogLevel)
	SetDestination(dest LogDestination, filePath ...string)
}
