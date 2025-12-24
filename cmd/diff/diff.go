package cmd

import (
	"flag"
	"fmt"
	"io"
)

const DiffUsage = `Usage: grammar diff ARGUMENT PATH1 PATH2

Compare input files and print their differences.
Both paths most be files (not directories).

Arguemnt:
-t, --tokens   Print differences in tokens.
-a, --ast      Print differences in ast nodes.
`

func DiffCommand(args []string, stdout, stderr io.Writer) int {
	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	var token bool
	var ast bool
	diffCmd.SetOutput(stderr)
	diffCmd.BoolVar(&token, "t", false, "Print differences in tokens.")
	diffCmd.BoolVar(&token, "tokens", false, "Print differences in tokens.")
	diffCmd.BoolVar(&ast, "a", false, "Print differences in ast nodes.")
	diffCmd.BoolVar(&ast, "ast", false, "Print differences in ast nodes.")
	var help bool
	diffCmd.BoolVar(&help, "h", false, "Print this message.")
	diffCmd.BoolVar(&help, "help", false, "Print this message.")
	diffCmd.Parse(args)
	if help || diffCmd.NArg() != 2 {
		fmt.Fprint(stdout, DiffUsage)
		return 0
	}
	fmt.Fprintf(stdout, "Diffing %s %s\n", diffCmd.Arg(0), diffCmd.Arg(1))
	return 0
}