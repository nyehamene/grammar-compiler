package cmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"io"
)

func CheckCommand(args []string, stdout, stderr io.Writer) int {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	var help bool
	checkCmd.SetOutput(stderr)
	checkCmd.BoolVar(&help, "h", false, "Print this message.")
	checkCmd.BoolVar(&help, "help", false, "Print this message.")
	checkCmd.Parse(args)
	if help || checkCmd.NArg() == 0 {
		fmt.Fprint(stdout, command.CheckUsage)
		return 0
	}
	for _, path := range checkCmd.Args() {
		fmt.Fprintf(stdout, "Validating %s...\n", path)
	}
	return 0
}