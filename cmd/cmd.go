package cmd

import (
	"fmt"
	checkcmd "grammar/cmd/check"
	diffcmd "grammar/cmd/diff"
	fmtcmd "grammar/cmd/fmt"
	lspcmd "grammar/cmd/lsp"
	printcmd "grammar/cmd/print"
	"grammar/command"
	"io"
)

// NOTE: update after each feature implementation
const VERSION = "0.4.0"

// Run executes the main command-line interface.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stdout, command.GrammarUsage)
		return 1
	}

	subcommand := args[0]
	subcommandArgs := args[1:]

	switch subcommand {
	case "fmt":
		return fmtcmd.FmtCommand(subcommandArgs, stdin, stdout, stderr)
	case "check":
		return checkcmd.CheckCommand(subcommandArgs, stdout, stderr)
	case "print":
		return printcmd.PrintCommand(subcommandArgs, stdout, stderr)
	case "diff":
		return diffcmd.DiffCommand(subcommandArgs, stdout, stderr)
	case "lsp":
		return lspcmd.LspCommand(subcommandArgs, stdin, stdout, stderr)
	case "help":
		_, _ = fmt.Fprint(stdout, command.GrammarUsage)
		return 0
	case "version":
		_, _ = fmt.Fprintln(stdout, VERSION)
		return 0
	default:
		_, _ = fmt.Fprint(stdout, command.GrammarUsage)
		return 1
	}
}
