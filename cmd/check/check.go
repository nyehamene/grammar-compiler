package cmd

import (
	"flag"
	"fmt"
	"grammar/ast"
	"grammar/command"
	"grammar/token"
	"io"
	"os"
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
		fileContent, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "Error opening file %s: %s\n", path, err)
			return 1
		}
		srcRunes := []rune(string(fileContent))

		tokenizer := token.NewTokenizer(srcRunes)
		tokens := tokenizer.Scan()

		parser := ast.NewParser(tokens, srcRunes)
		_, err = parser.ParseFile()
		if err != nil {
			if errs, ok := err.(ast.ErrorList); ok {
				for _, e := range errs {
					line, col := token.FindLineAndCol(int(e.Pos), srcRunes)
					fmt.Fprintf(stderr, "%s:%d:%d: %s\n", path, line, col, e.Message)
				}
			} else {
				fmt.Fprintf(stderr, "Error parsing file %s: %s\n", path, err)
			}
			return 1
		}

		fmt.Fprintf(stdout, "Validating %s...\n", path)
	}
	return 0
}
