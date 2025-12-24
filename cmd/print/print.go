package printcmd

import (
	"flag"
	"fmt"
	"io"
)

const PrintUsage = `Usage: grammar print ARGUMENT PATH

Pretty print AST nodes or token information to the console.
Argument:
-t, --token   Print the tokens produce by the tokenizer.
-a, --ast     Print the nodes produced by the parser.
PATH          Path to a file or directory.

If PATH is a directory, print every file inside it. If an error
is found, print the tokens/ast and highlight the errors.
`

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
		fmt.Fprint(stdout, PrintUsage)
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
