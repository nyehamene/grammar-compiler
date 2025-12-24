package cmd

import (
	"flag"
	"fmt"
	"io"
)

const LspUsage = `Usage: grammar lsp

Starts a language server for the grammar language.

Argument:
-h, --help    Print this message.
`

func LspCommand(args []string, stdout, stderr io.Writer) int {
	lspCmd := flag.NewFlagSet("lsp", flag.ExitOnError)
	var help bool
	lspCmd.SetOutput(stderr)
	lspCmd.BoolVar(&help, "h", false, "Print this message.")
	lspCmd.BoolVar(&help, "help", false, "Print this message.")
	lspCmd.Parse(args)
	if help {
		fmt.Fprint(stdout, LspUsage)
		return 0
	}
	fmt.Fprintln(stdout, "TBD")
	return 0
}