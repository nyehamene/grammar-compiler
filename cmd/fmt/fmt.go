package cmd

import (
	"flag"
	"fmt"
	"io"
)

const FmtUsage = `Usage: grammar fmt ARGUMENT

Formats input files inplace. Any syntax error is printed to
the console and the input file is left unchanged.

Argument:
PATH         Path to a file or directory.
-h, --help   Print this message.

If PATH is a directory, every grammar file inside it is formatted
inplace. If an error is found the file containing the error is left unchanged.
`

func FmtCommand(args []string, stdout, stderr io.Writer) int {
	fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	var help bool
	fmtCmd.SetOutput(stderr)
	fmtCmd.BoolVar(&help, "h", false, "Print this message.")
	fmtCmd.BoolVar(&help, "help", false, "Print this message.")
	fmtCmd.Parse(args)
	if help || fmtCmd.NArg() == 0 {
		fmt.Fprint(stdout, FmtUsage)
		return 0
	}
	for _, path := range fmtCmd.Args() {
		fmt.Fprintf(stdout, "Formatting %s...\n", path)
	}
	return 0
}
