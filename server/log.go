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
	case RequestMessage: // Changed to pointer
		msg = fmt.Sprintf("(->) Request %d-%s", t.ID, t.Method)
	case ResponseMessage: // Changed to pointer
		if t.Error != nil {
			msg = fmt.Sprintf("(<-) Response %d (Error: %s)", t.ID, t.Error.Message)
		} else {
			msg = fmt.Sprintf("(<-) Response %d", t.ID)
		}
	case NotificationMessage: // Changed to pointer
		// Direction must be handled at the call site.
		msg = fmt.Sprintf("(--) Notification %s", t.Method)
	case InitializeParams: // Changed to pointer
		// Pretty-print capabilities as indented JSON
		caps, err := json.MarshalIndent(t.Capabilities, "", "  ")
		if err != nil {
			msg = fmt.Sprintf("Client capabilities: [error marshalling: %v]", err)
		} else {
			msg = fmt.Sprintf("Client capabilities:\n%s", string(caps))
		}
	case Diagnostic: // Changed to pointer
		msg = fmt.Sprintf("     diagnostic: [%d:%d] %s", t.Range.Start.Line, t.Range.Start.Character, t.Message)
	case string:
		msg = t
	case error:
		msg = t.Error()
	default:
		msg = fmt.Sprintf("unhandled log type: %T", v)
	}

	_, _ = fmt.Fprintln(l.out, msg)
}

func (l *lineLogger) Printf(format string, v ...any) {
	if l.out == nil {
		return
	}
	_, _ = fmt.Fprintf(l.out, format+"\n", v...)
}
