package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger is the interface for the logging functions
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, err error, fields ...Field)
	WithContext(ctx context.Context) Logger
}

// Field is a log field
type Field struct {
	Key   string
	Value interface{}
}

// logger implements the Logger interface
type logger struct {
	logger zerolog.Logger
	ctx    context.Context
}

// Config contains the logger configuration
type Config struct {
	LogLevel string
	Pretty   bool
	WithTime bool
}

// NewLogger creates a new logger
func NewLogger(config Config) Logger {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Set up the log level
	level := zerolog.InfoLevel
	switch config.LogLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set up the writer
	var output io.Writer = os.Stdout
	if config.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create the logger
	loggerInstance := zerolog.New(output)
	if config.WithTime {
		loggerInstance = loggerInstance.With().Timestamp().Logger()
	}

	return &logger{
		logger: loggerInstance,
	}
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...Field) {
	event := l.getLoggerWithContext().Debug()
	event = addFields(event, fields)
	event.Msg(msg)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...Field) {
	event := l.getLoggerWithContext().Info()
	event = addFields(event, fields)
	event.Msg(msg)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...Field) {
	event := l.getLoggerWithContext().Warn()
	event = addFields(event, fields)
	event.Msg(msg)
}

// Error logs an error message
func (l *logger) Error(msg string, err error, fields ...Field) {
	event := l.getLoggerWithContext().Error()
	if err != nil {
		event = event.Err(err)
	}
	event = addFields(event, fields)
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string, err error, fields ...Field) {
	event := l.getLoggerWithContext().Fatal()
	if err != nil {
		event = event.Err(err)
	}
	event = addFields(event, fields)
	event.Msg(msg)
}

// WithContext adds a context to the logger
func (l *logger) WithContext(ctx context.Context) Logger {
	return &logger{
		logger: l.logger,
		ctx:    ctx,
	}
}

// getLoggerWithContext returns a logger with context values
func (l *logger) getLoggerWithContext() zerolog.Logger {
	if l.ctx == nil {
		return l.logger
	}

	// Extract request ID from context if available
	if requestID, ok := l.ctx.Value("requestID").(string); ok {
		return l.logger.With().Str("requestID", requestID).Logger()
	}

	// Extract user ID from context if available
	if userID, ok := l.ctx.Value("userID").(string); ok {
		return l.logger.With().Str("userID", userID).Logger()
	}

	return l.logger
}

// addFields adds fields to a log event
func addFields(event *zerolog.Event, fields []Field) *zerolog.Event {
	for _, field := range fields {
		switch v := field.Value.(type) {
		case string:
			event = event.Str(field.Key, v)
		case int:
			event = event.Int(field.Key, v)
		case bool:
			event = event.Bool(field.Key, v)
		case float64:
			event = event.Float64(field.Key, v)
		default:
			event = event.Interface(field.Key, v)
		}
	}
	return event
}
