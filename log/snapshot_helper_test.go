package log

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"grammar/testutil"
)

func TestJSONLoggerWithSnapshotHelper(t *testing.T) {
	buf := &strings.Builder{}
	logger := NewJSONLogger(buf, DEBUG, false)

	logger.Info("test request", Fields{
		"method":       "textDocument/definition",
		"request_id":   1,
		"document_uri": "file:///test.grammar",
	})

	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	got["timestamp"] = "2024-01-01T00:00:00Z"

	testutil.AssertSnapshotJSON(t, "json_logger_output", got)
}

func TestConsoleLoggerWithSnapshotHelper(t *testing.T) {
	buf := &strings.Builder{}
	logger := NewConsoleLogger(buf, DEBUG)

	logger.Info("test request", Fields{
		"method": "textDocument/definition",
	})

	output := buf.String()

	output = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z`).ReplaceAllString(output, "2024-01-01T00:00:00Z")

	testutil.AssertSnapshotText(t, "console_logger_output", output)
}
