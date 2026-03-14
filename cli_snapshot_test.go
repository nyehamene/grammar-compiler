package main

import (
	"bytes"
	"grammar/cmd"
	"grammar/command"
	"grammar/testutil"
	"os"
	"testing"
)

func TestCLISnapshot(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "help_fmt",
			args: []string{"fmt", "-h"},
		},
		{
			name: "help_check",
			args: []string{"check", "-h"},
		},
		{
			name: "help_print",
			args: []string{"print", "--help"},
		},
		{
			name: "help_diff",
			args: []string{"diff", "-h"},
		},
		{
			name: "help_lsp",
			args: []string{"lsp", "--help"},
		},
		{
			name: "help_help",
			args: []string{"help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			var stdin = os.Stdin
			cmd.Run(tt.args, stdin, &stdout, &stderr)

			if stdout.Len() > 0 {
				testutil.AssertSnapshotText(t, tt.name, stdout.String())
			} else if stderr.Len() > 0 {
				testutil.AssertSnapshotText(t, tt.name, stderr.String())
			}
		})
	}
}

func TestCLIErrorSnapshot(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "error_no_command",
			args: []string{},
		},
		{
			name: "error_unknown_command",
			args: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			var stdin = os.Stdin
			cmd.Run(tt.args, stdin, &stdout, &stderr)

			output := stderr.String()
			if output == "" {
				output = stdout.String()
			}
			if output == "" {
				t.Fatal("no output captured")
			}
			testutil.AssertSnapshotText(t, tt.name, output)
		})
	}
}

var (
	_ = command.GrammarUsage
	_ = command.CheckUsage
	_ = command.DiffUsage
	_ = command.FmtUsage
	_ = command.PrintUsage
	_ = command.LspUsage
)
