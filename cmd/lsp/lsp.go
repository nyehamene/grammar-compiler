package cmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"io"
)

func LspCommand(args []string, stdout, stderr io.Writer) int {
	lspCmd := flag.NewFlagSet("lsp", flag.ExitOnError)
	var help bool
	lspCmd.SetOutput(stderr)
	lspCmd.BoolVar(&help, "h", false, "Print this message.")
	lspCmd.BoolVar(&help, "help", false, "Print this message.")
	lspCmd.Parse(args)
	if help {
		fmt.Fprint(stdout, command.LspUsage)
		return 0
	}
	fmt.Fprintln(stdout, "TBD")
	return 0
}
