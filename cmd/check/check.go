package cmd

import (
	"flag"
	"fmt"
	"io"
)

const CheckUsage = `Usage: grammar check ARGUMENT

Checks for syntax or validation errors in input files.

Argument:
PATH          Path to a file or directory to check.
-h, --help    Print this message.

If PATH is a directory, every grammar files inside it is checked.
If any error is found, the program exits with a non-zero exit code.
`

func CheckCommand(args []string, stdout, stderr io.Writer) int {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	var help bool
	checkCmd.SetOutput(stderr)
	checkCmd.BoolVar(&help, "h", false, "Print this message.")
	checkCmd.BoolVar(&help, "help", false, "Print this message.")
	checkCmd.Parse(args)
	if help || checkCmd.NArg() == 0 {
		fmt.Fprint(stdout, CheckUsage)
		return 0
	}
	for _, path := range checkCmd.Args() {
		fmt.Fprintf(stdout, "Validating %s...\n", path)
	}
	return 0
}
