package log

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONLoggerSnapshot(t *testing.T) {
	buf := &strings.Builder{}
	logger := NewJSONLogger(buf, DEBUG, false)

	logger.Info("test request", Fields{
		"method":       "textDocument/definition",
		"request_id":   1,
		"document_uri": "file:///test.grammar",
	})

	output := buf.String()

	t.Logf("Logger output: %s", output)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if got["message"] != "test request" {
		t.Errorf("Expected message 'test request', got %v", got["message"])
	}

	if got["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", got["level"])
	}

	if got["method"] != "textDocument/definition" {
		t.Errorf("Expected method 'textDocument/definition', got %v", got["method"])
	}
}

func TestConsoleLoggerSnapshot(t *testing.T) {
	buf := &strings.Builder{}
	logger := NewConsoleLogger(buf, DEBUG)

	logger.Info("test request", Fields{
		"method": "textDocument/definition",
	})

	output := buf.String()

	t.Logf("Console output: %s", output)

	if output == "" {
		t.Error("Expected non-empty output")
	}

	if !strings.Contains(output, "test request") {
		t.Error("Expected output to contain 'test request'")
	}

	if !strings.Contains(output, "info") {
		t.Error("Expected output to contain 'info'")
	}
}
