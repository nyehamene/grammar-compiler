package cmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"grammar/server" // Import the new server package
	"io"
	"os" // Needed for os.Stdin and os.Stdout
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

	// Create and start the LSP server
	s := server.NewServer(os.Stdin, os.Stdout)
	s.Start()

	return 0
}
