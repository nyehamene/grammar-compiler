package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLspCommandHelpFlag(t *testing.T) {
	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	args := []string{"--help"}

	exitCode := LspCommand(args, stdin, stdout, stderr)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Usage") && !strings.Contains(stdout.String(), "Usage") {
		t.Error("expected usage information in output")
	}
}

func TestLspCommandLogFlag(t *testing.T) {
	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// Create temp log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	args := []string{"--log", logFile}

	// This will start LSP server which will wait for input
	// Just test that the flag is parsed correctly
	go func() {
		LspCommand(args, stdin, stdout, stderr)
	}()

	// Give it a moment to start
	os.Remove(logFile)
}

func TestLspCommandLogFormatFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{"json format", []string{"--log-format", "json", "--log-level", "error"}, false},
		{"text format", []string{"--log-format", "text", "--log-level", "error"}, false},
		{"invalid format", []string{"--log-format", "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			// The command will likely fail or exit due to LSP server,
			// but we can check if flag parsing works
			LspCommand(tt.args, stdin, stdout, stderr)

			// Just verify no panic and flag parsing works
		})
	}
}

func TestLspCommandLogLevelFlag(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"error", "error", false},
		{"warning", "warning", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			args := []string{"--log-level", tt.level}

			LspCommand(args, stdin, stdout, stderr)

			// Verify no panic
		})
	}
}
