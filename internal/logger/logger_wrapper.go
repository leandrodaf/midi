package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/leandrodaf/midi/sdk/contracts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger é uma implementação do contrato de Logger que usa o logger do Uber.
type ZapLogger struct {
	logger *zap.Logger
	level  contracts.LogLevel // Nível de log
}

// NewZapLogger cria um novo logger do Uber.
func NewZapLogger() contracts.Logger {
	logger, _ := zap.NewProduction() // Ou zap.NewDevelopment() para desenvolvimento
	return &ZapLogger{logger: logger, level: contracts.InfoLevel}
}

// Info logs a message at the INFO level
func (z *ZapLogger) Info(msg string, fields ...contracts.Field) {
	z.log(zapcore.InfoLevel, msg, fields...)
}

// Error logs a message at the ERROR level
func (z *ZapLogger) Error(msg string, fields ...contracts.Field) {
	z.log(zapcore.ErrorLevel, msg, fields...)
}

// Debug logs a message at the DEBUG level
func (z *ZapLogger) Debug(msg string, fields ...contracts.Field) {
	z.log(zapcore.DebugLevel, msg, fields...)
}

// Warn logs a message at the WARN level
func (z *ZapLogger) Warn(msg string, fields ...contracts.Field) {
	z.log(zapcore.WarnLevel, msg, fields...)
}

// Fatal logs a message at the FATAL level and terminates the application
func (z *ZapLogger) Fatal(msg string, fields ...contracts.Field) {
	z.log(zapcore.FatalLevel, msg, fields...)
	os.Exit(1)
}

// Field returns a new instance of Field
func (z *ZapLogger) Field() contracts.Field {
	return &zapField{}
}

// SetLevel sets the logging level
func (z *ZapLogger) SetLevel(level contracts.LogLevel) {
	z.level = level
}

// SetDestination sets the logging destination (não aplicável para ZapLogger).
func (z *ZapLogger) SetDestination(dest contracts.LogDestination, filePath ...string) {
	// O ZapLogger não tem suporte a filePath, então não implementamos essa funcionalidade.
}

// log é a função interna para registrar mensagens
func (z *ZapLogger) log(level zapcore.Level, msg string, fields ...contracts.Field) {
	if z.level > contracts.LogLevel(level) {
		return
	}

	// Captura o nome do arquivo e a linha onde o log foi chamado
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		file = filepath.Base(file)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	formattedFields := formatFields(fields...)
	logMessage := fmt.Sprintf("%s [%s] %s:%d: %s%s", timestamp, level.String(), file, line, msg, formattedFields)

	// Usar o logger do Uber
	switch level {
	case zapcore.InfoLevel:
		z.logger.Info(logMessage)
	case zapcore.ErrorLevel:
		z.logger.Error(logMessage)
	case zapcore.DebugLevel:
		z.logger.Debug(logMessage)
	case zapcore.WarnLevel:
		z.logger.Warn(logMessage)
	case zapcore.FatalLevel:
		z.logger.Fatal(logMessage)
	}
}

// formatFields formats additional fields
func formatFields(fields ...contracts.Field) string {
	if len(fields) == 0 {
		return ""
	}

	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		if f, ok := field.(*zapField); ok {
			fieldMap[f.key] = f.value
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

// zapField implements contracts.Field
type zapField struct {
	key   string
	value interface{}
}

func (f *zapField) Bool(key string, val bool) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Int(key string, val int) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Float64(key string, val float64) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) String(key string, val string) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Time(key string, val time.Time) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Int64(key string, val int64) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Error(key string, val error) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Uint64(key string, val uint64) contracts.Field {
	return &zapField{key, val}
}

func (f *zapField) Uint8(key string, val uint8) contracts.Field {
	return &zapField{key, val}
}
