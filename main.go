package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
)

//go:embed command/check.txt
var checkTxt string

//go:embed command/diff.txt
var diffTxt string

//go:embed command/fmt.txt
var fmtTxt string

//go:embed command/grammar.txt
var grammarTxt string

//go:embed command/print.txt
var printTxt string

func main() {
	// The `run` function now accepts `args` as a parameter. We pass `os.Args[1:]`
	// to exclude the program name itself.
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, grammarTxt)
		return 1
	}

	switch args[0] {
	case "fmt":
		fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
		var help bool
		fmtCmd.SetOutput(stderr)
		fmtCmd.BoolVar(&help, "h", false, "Print this message.")
		fmtCmd.BoolVar(&help, "help", false, "Print this message.")
		fmtCmd.Parse(args[1:])
		if help || fmtCmd.NArg() == 0 {
			fmt.Fprint(stdout, fmtTxt)
			return 0
		}
		for _, path := range fmtCmd.Args() {
			fmt.Fprintf(stdout, "Formatting %s...\n", path)
		}
	case "check":
		checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
		var help bool
		checkCmd.SetOutput(stderr)
		checkCmd.BoolVar(&help, "h", false, "Print this message.")
		checkCmd.BoolVar(&help, "help", false, "Print this message.")
		checkCmd.Parse(args[1:])
		if help || checkCmd.NArg() == 0 {
			fmt.Fprint(stdout, checkTxt)
			return 0
		}
		for _, path := range checkCmd.Args() {
			fmt.Fprintf(stdout, "Validating %s...\n", path)
		}
	case "print":
		printCmd := flag.NewFlagSet("print", flag.ExitOnError)
		var token bool
		var ast bool
		printCmd.SetOutput(stderr)
		printCmd.BoolVar(&token, "t", false, "Print the tokens produce by the tokenizer.")
		printCmd.BoolVar(&token, "token", false, "Print the tokens produce by the tokenizer.")
		printCmd.BoolVar(&ast, "a", false, "Print the nodes produced by the parser.")
		printCmd.BoolVar(&ast, "ast", false, "Print the nodes produced by the parser.")
		printCmd.Parse(args[1:])
		if printCmd.NArg() == 0 {
			fmt.Fprint(stdout, printTxt)
			return 0
		}
		kind := "AST"
		if token {
			kind = "token"
		}
		for _, path := range printCmd.Args() {
			fmt.Fprintf(stdout, "Printing %s %s...\n", kind, path)
		}
	case "diff":
		diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
		var token bool
		var ast bool
		diffCmd.SetOutput(stderr)
		diffCmd.BoolVar(&token, "t", false, "Print differences in tokens.")
		diffCmd.BoolVar(&token, "tokens", false, "Print differences in tokens.")
		diffCmd.BoolVar(&ast, "a", false, "Print differences in ast nodes.")
		diffCmd.BoolVar(&ast, "ast", false, "Print differences in ast nodes.")
		diffCmd.Parse(args[1:])
		if diffCmd.NArg() != 2 {
			fmt.Fprint(stdout, diffTxt)
			return 0
		}
		fmt.Fprintf(stdout, "Diffing %s %s\n", diffCmd.Arg(0), diffCmd.Arg(1))
	case "lsp":
		fmt.Fprintln(stdout, "TBD")
	case "help":
		fmt.Fprint(stdout, grammarTxt)
	case "version":
		fmt.Fprintln(stdout, "0.1.0")
	default:
		fmt.Fprint(stdout, grammarTxt)
		return 1
	}
	return 0
}
