package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"grammar/command"
	"grammar/server" // Import the new server package
	"io"
	"os" // Needed for os.Stdin and os.Stdout
	"path/filepath"
)

func LspCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	lspCmd := flag.NewFlagSet("lsp", flag.ExitOnError)
	var help bool
	lspCmd.SetOutput(stderr)
	lspCmd.BoolVar(&help, "h", false, "Print this message.")
	lspCmd.BoolVar(&help, "help", false, "Print this message.")
	_ = lspCmd.Parse(args)
	if help {
		_, _ = fmt.Fprint(stdout, command.LspUsage)
		return 0
	}

	var srv *server.Server
	var bufin = bufio.NewReader(stdin)

	logPath := filepath.Join(os.Getenv("HOME"), ".cache", "grammar")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		_, err = fmt.Fprintf(stderr, "Failed to create log directory: %v", err)
		srv = server.NewServer(bufin, stdout, stderr)
	}

	logFilePath := filepath.Join(logPath, "lsp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(stderr, "Failed to open log file: %v", err)
		srv = server.NewServer(bufin, stdout, stderr)
	}

	// Create and start the LSP server
	srv = server.NewServer(bufin, stdout, logFile)
	srv.Start()

	return 0
}
