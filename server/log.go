package server

import (
	"encoding/json"
	"fmt"
	"grammar/log"
	"io"
)

// --- Silent Logger ---

type silentLogger struct{}

func (s *silentLogger) Print(v any)                    {}
func (s *silentLogger) Printf(format string, v ...any) {}

// NewSilentLogger creates a logger that discards all output.
func NewSilentLogger() log.Logger {
	return &silentLogger{}
}

// --- Writer Logger ---

// writerLogger wraps an io.Writer to satisfy the Logger interface.
type writerLogger struct {
	out io.Writer
}

// NewWriterLogger creates a logger that writes to the given writer.
func NewWriterLogger(out io.Writer) log.Logger {
	if out == nil {
		out = io.Discard
	}
	return &writerLogger{out: out}
}

func (l *writerLogger) Print(v any) {
	if l.out == nil {
		return
	}
	fmt.Fprintln(l.out, v)
}

func (l *writerLogger) Printf(format string, v ...any) {
	if l.out == nil {
		return
	}
	_, _ = fmt.Fprintf(l.out, format+"\n", v...)
}

// --- Test Logger (exported for use in tests) ---

// TestLogger wraps an io.Writer to satisfy the Logger interface for testing.
type TestLogger struct {
	out io.Writer
}

// NewTestLogger creates a logger that writes to the given writer for testing.
func NewTestLogger(out io.Writer) log.Logger {
	return NewWriterLogger(out)
}

// --- Line Logger ---

type lineLogger struct {
	out io.Writer
}

// NewLineLogger creates a logger that writes to the given writer.
func NewLineLogger(out io.Writer) log.Logger {
	if out == nil {
		out = io.Discard
	}
	return &lineLogger{out: out}
}

func (l *lineLogger) Print(v any) {
	if l.out == nil {
		return
	}

	var msg string
	switch t := v.(type) {
	case map[string]any:
		if idVal, ok := t["id"]; ok {
			// It's an incoming request
			if id, ok := idVal.(float64); ok { // JSON numbers are float64
				if method, ok := t["method"].(string); ok {
					msg = fmt.Sprintf("(->) Request %d-%s", int(id), method)
				}
			}
		} else if method, ok := t["method"].(string); ok {
			// It's an incoming notification
			msg = fmt.Sprintf("(->) Notification %s", method)
		} else {
			// Fallback for unexpected map structures, try to pretty print raw JSON
			jsonBytes, err := json.MarshalIndent(t, "", "  ")
			if err != nil {
				msg = fmt.Sprintf("unhandled raw message (json marshal error): %v", t)
			} else {
				msg = fmt.Sprintf("unhandled raw message:\n%s", string(jsonBytes))
			}
		}
	case ResponseMessage:
		if t.Error != nil {
			msg = fmt.Sprintf("(<-) Response %d (Error: %s)", *t.ID, t.Error.Message)
		} else {
			msg = fmt.Sprintf("(<-) Response %d", *t.ID)
		}
	case NotificationMessage:
		msg = fmt.Sprintf("(<-) Notification %s", t.Method)
	case *InitializeParams:
		// Pretty-print capabilities as indented JSON
		caps, err := json.MarshalIndent(t.Capabilities, "", "  ")
		if err != nil {
			msg = fmt.Sprintf("Client capabilities: [error marshalling: %v]", err)
		} else {
			msg = fmt.Sprintf("Client capabilities:\n%s", string(caps))
		}
	case *Diagnostic:
		msg = fmt.Sprintf("     diagnostic: [%d:%d] %s", t.Range.Start.Line, t.Range.Start.Character, t.Message)
	case string:
		msg = t
	case error:
		msg = t.Error()
	default:
		// For any other unhandled type, try to marshal to indented JSON.
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			msg = fmt.Sprintf("unhandled log type: %T (json marshal error): %v", v, v)
		} else {
			msg = fmt.Sprintf("unhandled log type: %T:\n%s", v, string(jsonBytes))
		}
	}

	_, _ = fmt.Fprintln(l.out, msg)
}

func (l *lineLogger) Printf(format string, v ...any) {
	if l.out == nil {
		return
	}
	_, _ = fmt.Fprintf(l.out, format+"\n", v...)
}
