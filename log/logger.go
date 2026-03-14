package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "debug"
	case INFO:
		return "info"
	case WARN:
		return "warn"
	case ERROR:
		return "error"
	default:
		return "unknown"
	}
}

// Logger defines an interface for logging.
type Logger interface {
	Print(v any)
	Printf(format string, v ...any)
}

// --- Stderr Logger ---

type stderrLogger struct {
	out io.Writer
}

// NewStderrLogger creates a logger that writes to stderr.
func NewStderrLogger() Logger {
	return &stderrLogger{out: os.Stderr}
}

func (l *stderrLogger) Print(v any) {
	if l.out == nil {
		return
	}
	fmt.Fprintln(l.out, v)
}

func (l *stderrLogger) Printf(format string, v ...any) {
	if l.out == nil {
		return
	}
	fmt.Fprintf(l.out, format+"\n", v...)
}

// --- Silent Logger ---

type silentLogger struct{}

func (s *silentLogger) Print(v any)                    {}
func (s *silentLogger) Printf(format string, v ...any) {}

// NewSilentLogger creates a logger that discards all output.
func NewSilentLogger() Logger {
	return &silentLogger{}
}

// --- Structured Logger ---

// FieldKey represents keys for structured log fields.
type FieldKey string

// Common field keys
const (
	FieldMessage     FieldKey = "message"
	FieldLevel       FieldKey = "level"
	FieldTimestamp   FieldKey = "timestamp"
	FieldMethod      FieldKey = "method"
	FieldRequestID   FieldKey = "request_id"
	FieldDocumentURI FieldKey = "document_uri"
	FieldPosition    FieldKey = "position"
	FieldError       FieldKey = "error"
)

// Fields represents a map of log fields.
type Fields map[string]any

// StructuredLogger defines an interface for structured logging.
type StructuredLogger interface {
	Log(level Level, msg string, fields Fields)
	Debug(msg string, fields Fields)
	Info(msg string, fields Fields)
	Warn(msg string, fields Fields)
	Error(msg string, fields Fields)
}

// BaseStructuredLogger provides common functionality for structured loggers.
type BaseStructuredLogger struct {
	minLevel Level
}

func (l *BaseStructuredLogger) shouldLog(level Level) bool {
	return level >= l.minLevel
}

func (l *BaseStructuredLogger) Log(level Level, msg string, fields Fields) {
	// This method should be overridden by implementations
}

func (l *BaseStructuredLogger) Debug(msg string, fields Fields) {
	l.Log(DEBUG, msg, fields)
}

func (l *BaseStructuredLogger) Info(msg string, fields Fields) {
	l.Log(INFO, msg, fields)
}

func (l *BaseStructuredLogger) Warn(msg string, fields Fields) {
	l.Log(WARN, msg, fields)
}

func (l *BaseStructuredLogger) Error(msg string, fields Fields) {
	l.Log(ERROR, msg, fields)
}

// --- JSON Logger ---

// JSONLogger outputs structured logs in JSON format.
type JSONLogger struct {
	BaseStructuredLogger
	out io.Writer
}

// NewJSONLogger creates a logger that outputs JSON to the given writer.
func NewJSONLogger(out io.Writer, minLevel Level) StructuredLogger {
	return &JSONLogger{
		BaseStructuredLogger: BaseStructuredLogger{minLevel: minLevel},
		out:                  out,
	}
}

func (l *JSONLogger) Log(level Level, msg string, fields Fields) {
	if !l.shouldLog(level) {
		return
	}

	if fields == nil {
		fields = make(Fields)
	}

	fields[string(FieldTimestamp)] = time.Now().UTC().Format(time.RFC3339Nano)
	fields[string(FieldLevel)] = level.String()
	fields[string(FieldMessage)] = msg

	fmt.Fprintln(l.out, mustMarshalJSON(fields))
}

func (l *JSONLogger) Debug(msg string, fields Fields) {
	l.Log(DEBUG, msg, fields)
}

func (l *JSONLogger) Info(msg string, fields Fields) {
	l.Log(INFO, msg, fields)
}

func (l *JSONLogger) Warn(msg string, fields Fields) {
	l.Log(WARN, msg, fields)
}

func (l *JSONLogger) Error(msg string, fields Fields) {
	l.Log(ERROR, msg, fields)
}

// --- Console Logger ---

// ConsoleLogger outputs structured logs in a human-readable format.
type ConsoleLogger struct {
	BaseStructuredLogger
	out io.Writer
}

// NewConsoleLogger creates a logger that outputs to the console.
func NewConsoleLogger(out io.Writer, minLevel Level) StructuredLogger {
	return &ConsoleLogger{
		BaseStructuredLogger: BaseStructuredLogger{minLevel: minLevel},
		out:                  out,
	}
}

func (l *ConsoleLogger) Log(level Level, msg string, fields Fields) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	levelStr := level.String()

	// Build the message
	output := fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, msg)

	// Add fields (excluding timestamp, level, message which are already included)
	for k, v := range fields {
		if k == string(FieldTimestamp) || k == string(FieldLevel) || k == string(FieldMessage) {
			continue
		}
		output += fmt.Sprintf(" %s=%v", k, v)
	}

	fmt.Fprintln(l.out, output)
}

func (l *ConsoleLogger) Debug(msg string, fields Fields) {
	l.Log(DEBUG, msg, fields)
}

func (l *ConsoleLogger) Info(msg string, fields Fields) {
	l.Log(INFO, msg, fields)
}

func (l *ConsoleLogger) Warn(msg string, fields Fields) {
	l.Log(WARN, msg, fields)
}

func (l *ConsoleLogger) Error(msg string, fields Fields) {
	l.Log(ERROR, msg, fields)
}

// --- Multi Logger ---

// MultiLogger writes to multiple loggers.
type MultiLogger struct {
	loggers []StructuredLogger
}

// NewMultiLogger creates a logger that writes to multiple outputs.
func NewMultiLogger(loggers ...StructuredLogger) StructuredLogger {
	return &MultiLogger{loggers: loggers}
}

func (l *MultiLogger) Log(level Level, msg string, fields Fields) {
	for _, logger := range l.loggers {
		logger.Log(level, msg, fields)
	}
}

func (l *MultiLogger) Debug(msg string, fields Fields) {
	for _, logger := range l.loggers {
		logger.Debug(msg, fields)
	}
}

func (l *MultiLogger) Info(msg string, fields Fields) {
	for _, logger := range l.loggers {
		logger.Info(msg, fields)
	}
}

func (l *MultiLogger) Warn(msg string, fields Fields) {
	for _, logger := range l.loggers {
		logger.Warn(msg, fields)
	}
}

func (l *MultiLogger) Error(msg string, fields Fields) {
	for _, logger := range l.loggers {
		logger.Error(msg, fields)
	}
}

// --- Compatibility Wrapper ---

// StructuredToBasic wraps a StructuredLogger to also satisfy the basic Logger interface.
type StructuredToBasic struct {
	StructuredLogger
}

// Print satisfies the Logger interface.
func (s *StructuredToBasic) Print(v any) {
	s.Info(fmt.Sprintf("%v", v), nil)
}

// Printf satisfies the Logger interface.
func (s *StructuredToBasic) Printf(format string, v ...any) {
	s.Info(fmt.Sprintf(format, v...), nil)
}

// NewStructuredToBasic wraps a StructuredLogger to satisfy Logger.
func NewStructuredToBasic(logger StructuredLogger) Logger {
	return &StructuredToBasic{StructuredLogger: logger}
}

// mustMarshalJSON marshals fields to JSON. Panics on error.
func mustMarshalJSON(v any) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
