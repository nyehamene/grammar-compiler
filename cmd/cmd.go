package cmd

import (
	"fmt"
	"io"

	checkcmd "grammar/cmd/check"
	diffcmd "grammar/cmd/diff"
	fmtcmd "grammar/cmd/fmt"
	lspcmd "grammar/cmd/lsp"
	printcmd "grammar/cmd/print" // Using alias to avoid conflict with built-in print
)

const GrammarUsage = `Usage: grammar COMMAND

A tool for describing the syntax of programming languages.

Command:
fmt      Format grammar files inplace.
check    Validate grammar files and report diagnostic errors.
print    Print tokens or Ast nodes.
diff     Compare 2 input files.
lsp      Start a language server.
help     Print this message.
version  Print version infomation.

An grammar file is any file with the file extension .grammar.
`

// Run executes the main command-line interface.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, GrammarUsage)
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
		fmt.Fprint(stdout, GrammarUsage)
		return 0
	case "version":
		fmt.Fprintln(stdout, "0.1.0")
		return 0
	default:
		fmt.Fprint(stdout, GrammarUsage)
		return 1
	}
}