package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "debug"},
		{INFO, "info"},
		{WARN, "warn"},
		{ERROR, "error"},
		{Level(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("Level.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLevelOrder(t *testing.T) {
	if DEBUG >= INFO {
		t.Error("DEBUG should be less than INFO")
	}
	if INFO >= WARN {
		t.Error("INFO should be less than WARN")
	}
	if WARN >= ERROR {
		t.Error("WARN should be less than ERROR")
	}
}

func TestJSONLoggerOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, DEBUG)

	logger.Info("test message", Fields{
		"method": "testMethod",
	})

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Fatal("Expected JSON output, got empty string")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["message"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", parsed["message"])
	}

	if parsed["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", parsed["level"])
	}

	if parsed["timestamp"] == nil {
		t.Error("Expected timestamp field")
	}
}

func TestJSONLoggerLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, WARN)

	logger.Debug("debug message", nil)
	logger.Info("info message", nil)
	logger.Warn("warn message", nil)
	logger.Error("error message", nil)

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines (warn + error), got %d", len(lines))
	}
}

func TestJSONLoggerCustomFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, DEBUG)

	logger.Info("request", Fields{
		"method":       "textDocument/definition",
		"request_id":   1,
		"document_uri": "file:///test.grammar",
	})

	output := strings.TrimSpace(buf.String())
	var parsed map[string]any
	json.Unmarshal([]byte(output), &parsed)

	if parsed["method"] != "textDocument/definition" {
		t.Errorf("Expected method field, got %v", parsed["method"])
	}
	if parsed["request_id"] != float64(1) {
		t.Errorf("Expected request_id 1, got %v", parsed["request_id"])
	}
}

func TestConsoleLoggerOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, DEBUG)

	logger.Info("test message", Fields{
		"method": "testMethod",
	})

	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain message")
	}
	if !strings.Contains(output, "info") {
		t.Error("Expected output to contain level")
	}
}

func TestConsoleLoggerLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, ERROR)

	logger.Debug("debug", nil)
	logger.Info("info", nil)
	logger.Warn("warn", nil)
	logger.Error("error", nil)

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	if len(lines) != 1 {
		t.Errorf("Expected 1 log line (error only), got %d", len(lines))
	}
}

func TestMultiLogger(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	logger1 := NewJSONLogger(buf1, DEBUG)
	logger2 := NewConsoleLogger(buf2, DEBUG)

	multi := NewMultiLogger(logger1, logger2)
	multi.Info("test message", Fields{"key": "value"})

	if strings.TrimSpace(buf1.String()) == "" {
		t.Error("Expected output in first logger")
	}
	if strings.TrimSpace(buf2.String()) == "" {
		t.Error("Expected output in second logger")
	}
}

func TestMultiLoggerEmpty(t *testing.T) {
	multi := NewMultiLogger()
	multi.Info("test", nil)
}

func TestStructuredToBasic(t *testing.T) {
	buf := &bytes.Buffer{}
	structured := NewJSONLogger(buf, DEBUG)
	basic := NewStructuredToBasic(structured)

	basic.Print("test print")
	basic.Printf("test %s", "printf")

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
}
