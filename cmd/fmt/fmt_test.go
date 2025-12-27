package cmd

import (
	"bytes"
	"grammar/command"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	unformattedComplex = `a = "a";
longname = "b";

c = "c" | "d";
`
	formattedComplex = `a        = "a";
longname = "b";

c = "c" | "d";
`
	invalidContent = `a = ;`
)

func TestFmtCommand(t *testing.T) {

	t.Run("HelpFlag", func(t *testing.T) {
		stdin := &bytes.Buffer{}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{"--help"}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
		if !strings.Contains(stdout.String(), command.FmtUsage) {
			t.Errorf("stdout should contain usage info")
		}
	})

	t.Run("NoArgs", func(t *testing.T) {
		stdin := &bytes.Buffer{}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
		if !strings.Contains(stdout.String(), command.FmtUsage) {
			t.Errorf("stdout should contain usage info")
		}
	})

	t.Run("StdinFormatting", func(t *testing.T) {
		stdin := strings.NewReader(unformattedComplex)
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{"--stdin"}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
		}

		if formattedComplex != stdout.String() {
			t.Errorf("FmtCommand() mismatch want:\n%q\ngot:\n%q", formattedComplex, stdout.String())
		}
	})

	t.Run("StdinFormattingError", func(t *testing.T) {
		stdin := strings.NewReader(invalidContent)
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{"--stdin"}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
		if !strings.Contains(stderr.String(), "parsing error from stdin") {
			t.Errorf("stderr should contain parsing error, got: %s", stderr.String())
		}
	})

	t.Run("FileFormattingInPlace", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "test.grammar")
		os.WriteFile(filePath, []byte(unformattedComplex), 0644)

		stdin := &bytes.Buffer{}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{filePath}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
		}

		content, _ := os.ReadFile(filePath)

		if formattedComplex != string(content) {
			t.Errorf("File content mismatch want:\n%q\ngot:\n%q", formattedComplex, string(content))
		}

		if !strings.Contains(stdout.String(), "Formatted "+filePath) {
			t.Errorf("stdout should contain formatted message")
		}
	})

	t.Run("FileFormattingToStdout", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "test.grammar")
		os.WriteFile(filePath, []byte(unformattedComplex), 0644)

		stdin := &bytes.Buffer{}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{"--stdout", filePath}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
		}

		if formattedComplex != stdout.String() {
			t.Errorf("FmtCommand() mismatch want:\n%q\ngot:\n%q", formattedComplex, stdout.String())
		}

		content, _ := os.ReadFile(filePath)
		if string(content) != unformattedComplex {
			t.Errorf("original file should not be modified, got:\n%q", string(content))
		}
	})

	t.Run("FileFormattingError", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "test.grammar")
		os.WriteFile(filePath, []byte(invalidContent), 0644)

		stdin := &bytes.Buffer{}
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		args := []string{filePath}

		exitCode := FmtCommand(args, stdin, stdout, stderr)

		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}

		if !strings.Contains(stderr.String(), "parsing error") {
			t.Errorf("stderr should contain parsing error")
		}
	})
}
