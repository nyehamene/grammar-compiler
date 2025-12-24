package cmd

import (
	"fmt"
	"grammar/command"
	checkcmd "grammar/cmd/check"
	diffcmd "grammar/cmd/diff"
	fmtcmd "grammar/cmd/fmt"
	lspcmd "grammar/cmd/lsp"
	printcmd "grammar/cmd/print"
	"io"
)

// Run executes the main command-line interface.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, command.GrammarUsage)
		return 1
	}

	subcommand := args[0]
	subcommandArgs := args[1:]

	switch subcommand {
	case "fmt":
		return fmtcmd.FmtCommand(subcommandArgs, stdout, stderr)
	case "check":
		return checkcmd.CheckCommand(subcommandArgs, stdout, stderr)
	case "print":
		return printcmd.PrintCommand(subcommandArgs, stdout, stderr)
	case "diff":
		return diffcmd.DiffCommand(subcommandArgs, stdout, stderr)
	case "lsp":
		return lspcmd.LspCommand(subcommandArgs, stdout, stderr)
	case "help":
		fmt.Fprint(stdout, command.GrammarUsage)
		return 0
	case "version":
		fmt.Fprintln(stdout, "0.1.0")
		return 0
	default:
		fmt.Fprint(stdout, command.GrammarUsage)
		return 1
	}
}
