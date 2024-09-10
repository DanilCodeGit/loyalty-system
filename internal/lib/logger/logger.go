package logger

import (
	"log/slog"
	"os"
)

// Logger Структура для логгера
type Logger struct {
	logger *slog.Logger
}

// NewLogger returns a new Logger.
func NewLogger() *Logger {

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return &Logger{
		logger: log,
	}
}

// Info logs a message at the Info level.
func (log *Logger) Info(msg string, fields ...any) {
	log.logger.Info(msg, fields...)
}

// Error logs a message at the Error level.
func (log *Logger) Error(msg string, fields ...any) {
	log.logger.Error(msg, fields...)
}

// With returns a new Logger with the specified fields.
func (log *Logger) With(fields ...any) *Logger {
	newLog := &Logger{logger: log.logger.With(fields...)}
	return newLog
}

// ErrorField returns an error with the specified fields.
func (log *Logger) ErrorField(err error) interface{} {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

// StringField returns a string field with the specified key and
func (log *Logger) StringField(key, value string) interface{} {
	return slog.String(key, value)
}

// AnyField returns a any field with the specified key and
func (log *Logger) AnyField(key string, value any) interface{} {
	return slog.Any(key, value)
}
