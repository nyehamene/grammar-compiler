package cmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"io"
)

func FmtCommand(args []string, stdout, stderr io.Writer) int {
	fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	var help bool
	fmtCmd.SetOutput(stderr)
	fmtCmd.BoolVar(&help, "h", false, "Print this message.")
	fmtCmd.BoolVar(&help, "help", false, "Print this message.")
	fmtCmd.Parse(args)
	if help || fmtCmd.NArg() == 0 {
		fmt.Fprint(stdout, command.FmtUsage)
		return 0
	}
	for _, path := range fmtCmd.Args() {
		fmt.Fprintf(stdout, "Formatting %s...\n", path)
	}
	return 0
}