package cmd

import (
	"flag"
	"fmt"
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

		hasError := false
		for _, t := range tokens {
			if t.State == token.Invalid {
				// Get the token value from the tokenizer using its start and end offsets
				tokenValue := tokenizer.Literal(t, srcRunes)
				fmt.Fprintf(stderr, "Error: Invalid token '%s' in file %s\n", tokenValue, path)
				hasError = true
				break
			}
		}

		if hasError {
			return 1
		}

		fmt.Fprintf(stdout, "Validating %s...\n", path)
	}
	return 0
}
