package main

import (
	"bytes"
	"strings"
	"testing"
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
			args:   []string{"print"},
			stdout: printTxt,
		},
		{
			name:   "diff help",
			args:   []string{"diff"},
			stdout: diffTxt,
		},
		{
			name:   "no command",
			args:   []string{},
			stdout: grammarTxt,
			exitCode: 1,
		},
		{
			name: "fmt path",
			args: []string{"fmt", "testdata/file.grammar"},
			stdout: "Formatting testdata/file.grammar...\n",
		},
		{
			name: "check path",
			args: []string{"check", "testdata/file.grammar"},
			stdout: "Validating testdata/file.grammar...\n",
		},
		{
			name: "print token path",
			args: []string{"print", "-t", "testdata/file.grammar"},
			stdout: "Printing token testdata/file.grammar...\n",
		},
		{
			name: "print ast path",
			args: []string{"print", "-a", "testdata/file.grammar"},
			stdout: "Printing AST testdata/file.grammar...\n",
		},
		{
			name: "diff paths",
			args: []string{"diff", "testdata/file1.grammar", "testdata/file2.grammar"},
			stdout: "Diffing testdata/file1.grammar testdata/file2.grammar\n",
		},
		{
			name: "lsp command",
			args: []string{"lsp"},
			stdout: "TBD\n",
		},
		{
			name: "help command",
			args: []string{"help"},
			stdout: grammarTxt,
		},
		{
			name: "version command",
			args: []string{"version"},
			stdout: "TBD\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			gotExitCode := run(tt.args, &stdout, &stderr)

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
