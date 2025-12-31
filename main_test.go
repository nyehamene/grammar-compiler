package main

import (
	"bytes"
	"grammar/cmd"
	"grammar/command"
	"os"
	"strings"
	"testing"
)

// These variables now refer to the exported Usage constants in the command package.
var (
	grammarTxt = command.GrammarUsage
	checkTxt   = command.CheckUsage
	diffTxt    = command.DiffUsage
	fmtTxt     = command.FmtUsage
	printTxt   = command.PrintUsage
	lspTxt     = command.LspUsage
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdout   string
		stderr   string
		exitCode int
	}{
		{
			name:   "fmt help",
			args:   []string{"fmt", "-h"},
			stdout: fmtTxt,
		},
		{
			name:   "check help",
			args:   []string{"check", "-h"},
			stdout: checkTxt,
		},
		{
			name:   "print help",
			args:   []string{"print", "--help"},
			stdout: printTxt,
		},
		{
			name:   "diff help",
			args:   []string{"diff", "-h"},
			stdout: diffTxt,
		},
		{
			name:   "lsp help",
			args:   []string{"lsp", "--help"},
			stdout: lspTxt,
		},
		{
			name:     "no command",
			args:     []string{},
			stdout:   grammarTxt,
			exitCode: 1,
		},

		{
			name:   "help command",
			args:   []string{"help"},
			stdout: grammarTxt,
		},
		{
			name:   "version command",
			args:   []string{"version"},
			stdout: "0.1.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			var stdin = os.Stdin
			gotExitCode := cmd.Run(tt.args, stdin, &stdout, &stderr) // Call cmd.Run

			if gotExitCode != tt.exitCode {
				t.Errorf("expected exit code %d, got %d", tt.exitCode, gotExitCode)
			}

			if !strings.Contains(stdout.String(), tt.stdout) {
				t.Errorf("expected stdout to contain %q, got %q", tt.stdout, stdout.String())
			}

			if !strings.Contains(stderr.String(), tt.stderr) {
				t.Errorf("expected stderr to contain %q, got %q", tt.stderr, stderr.String())
			}
		})
	}
}
