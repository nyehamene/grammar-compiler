package log

import (
	"fmt"
	"io"
	"os"
)

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

func (s *silentLogger) Print(v any) {}
func (s *silentLogger) Printf(format string, v ...any) {}

func NewSilentLogger() Logger {
	return &silentLogger{}
}
