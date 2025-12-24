package printcmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"io"
)

func PrintCommand(args []string, stdout, stderr io.Writer) int {
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	var token bool
	var ast bool
	printCmd.SetOutput(stderr)
	printCmd.BoolVar(&token, "t", false, "Print the tokens produce by the tokenizer.")
	printCmd.BoolVar(&token, "token", false, "Print the tokens produce by the tokenizer.")
	printCmd.BoolVar(&ast, "a", false, "Print the nodes produced by the parser.")
	printCmd.BoolVar(&ast, "ast", false, "Print the nodes produced by the parser.")
	var help bool
	printCmd.BoolVar(&help, "h", false, "Print this message.")
	printCmd.BoolVar(&help, "help", false, "Print this message.")
	printCmd.Parse(args)
	if help || printCmd.NArg() == 0 {
		fmt.Fprint(stdout, command.PrintUsage)
		return 0
	}
	kind := "AST"
	if token {
		kind = "token"
	}
	for _, path := range printCmd.Args() {
		fmt.Fprintf(stdout, "Printing %s %s...\n", kind, path)
	}
	return 0
}