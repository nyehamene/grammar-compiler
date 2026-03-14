package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"grammar/command"
	"grammar/log"
	"grammar/server"
	"io"
	"os"
)

func LspCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	lspCmd := flag.NewFlagSet("lsp", flag.ExitOnError)
	var help bool
	var logFile string
	var logFormat string
	var logLevel string
	var pretty bool

	lspCmd.SetOutput(stderr)
	lspCmd.BoolVar(&help, "h", false, "Print this message.")
	lspCmd.BoolVar(&help, "help", false, "Print this message.")
	lspCmd.StringVar(&logFile, "log", "", "Log file path. If not specified, no file logging.")
	lspCmd.StringVar(&logFormat, "log-format", "json", "Log format: json or text (default: json when --log is specified)")
	lspCmd.StringVar(&logLevel, "log-level", "debug", "Log level: debug, info, warn, error (default: debug)")
	lspCmd.BoolVar(&pretty, "pretty", false, "Pretty-print JSON output (default: false)")

	_ = lspCmd.Parse(args)

	if help {
		_, _ = fmt.Fprint(stdout, command.LspUsage)
		return 0
	}

	// Determine log level
	level := log.DEBUG
	switch logLevel {
	case "debug":
		level = log.DEBUG
	case "info":
		level = log.INFO
	case "warn", "warning":
		level = log.WARN
	case "error":
		level = log.ERROR
	default:
		level = log.DEBUG
	}

	// Determine log format based on whether log file is specified
	useJSON := logFile != ""
	if logFormat == "text" {
		useJSON = false
	}

	var logger log.StructuredLogger

	if logFile != "" {
		// Open log file
		logWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Failed to open log file: %v\n", err)
			// Fall back to stderr
			logger = log.NewConsoleLogger(stderr, log.INFO)
		} else {
			defer logWriter.Close()
			if useJSON {
				logger = log.NewJSONLogger(logWriter, level, pretty)
			} else {
				logger = log.NewConsoleLogger(logWriter, level)
			}
		}
	} else {
		// No log file - use stderr with text format and info level
		logger = log.NewConsoleLogger(stderr, log.INFO)
	}

	var bufin = bufio.NewReader(stdin)

	// Create and start the LSP server
	// Pass StructuredLogger directly - server will wrap it for Print/Printf
	srv := server.NewServer(bufin, stdout, logger)
	srv.Start()

	return 0
}
