package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/log"
	"grammar/server"
	"strings"
	"testing"
)

func TestServerLoggingJSON(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.NewJSONLogger(&logBuf, log.DEBUG)
	basicLogger := log.NewStructuredToBasic(logger)

	srv := server.NewServer(strings.NewReader(""), &bytes.Buffer{}, basicLogger)
	_ = srv

	logger.Info("test request", log.Fields{
		"method":       "textDocument/definition",
		"request_id":   1,
		"document_uri": "file:///test.grammar",
	})

	output := strings.TrimSpace(logBuf.String())
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed["message"] != "test request" {
		t.Errorf("Expected message 'test request', got %v", parsed["message"])
	}

	if parsed["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", parsed["level"])
	}
}

func TestServerLoggingConsole(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.NewConsoleLogger(&logBuf, log.DEBUG)
	basicLogger := log.NewStructuredToBasic(logger)

	srv := server.NewServer(strings.NewReader(""), &bytes.Buffer{}, basicLogger)
	_ = srv

	logger.Info("test message", log.Fields{
		"method": "testMethod",
	})

	output := strings.TrimSpace(logBuf.String())
	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain message")
	}
	if !strings.Contains(output, "info") {
		t.Error("Expected output to contain level")
	}
}

func TestServerSilentLogger(t *testing.T) {
	logger := server.NewSilentLogger()
	logger.Print("test")
	logger.Printf("test %s", "format")
}
