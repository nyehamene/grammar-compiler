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

func FmtCommand(args []string, stdout, stderr io.Writer) int {
	fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	var help bool
	var stdoutFlag bool
	fmtCmd.SetOutput(stderr)
	fmtCmd.BoolVar(&help, "h", false, "Print this message.")
	fmtCmd.BoolVar(&help, "help", false, "Print this message.")
	fmtCmd.BoolVar(&stdoutFlag, "stdout", false, "Print formatted output to stdout. Do not modify input file.")
	fmtCmd.Parse(args)
	if help || fmtCmd.NArg() == 0 {
		fmt.Fprint(stdout, command.FmtUsage)
		return 0
	}

	for _, path := range fmtCmd.Args() {
		var err error
		if stdoutFlag {
			err = formatFile(path, stdout, stderr, stdout)
		} else {
			err = formatFile(path, stdout, stderr, nil)
		}

		if err != nil {
			fmt.Fprintf(stderr, "Error formatting %s: %v\n", path, err)
			return 1
		}
		if !stdoutFlag {
			fmt.Fprintf(stdout, "Formatted %s\n", path)
		}
	}
	return 0
}

func formatFile(path string, stdout, stderr io.Writer, outputWriter io.Writer) error {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	srcRunes := []rune(string(fileContent))

	tokenizer := token.NewTokenizer(srcRunes, false, false) // Do not skip comments or newlines for formatting
	tokens := tokenizer.Scan()

	formatterParser := ast.NewFormatterParser(tokens, srcRunes)
	formatFile, err := formatterParser.Parse()
	if err != nil {
		if errs, ok := err.(ast.ErrorList); ok {
			for _, e := range errs {
				line, col := token.FindLineAndCol(int(e.Pos), srcRunes)
				fmt.Fprintf(stderr, "%s:%d:%d: %s\n", path, line, col, e.Message)
			}
		} else {
			fmt.Fprintf(stderr, "Error parsing file %s: %s\n", path, err)
		}
		return fmt.Errorf("parsing error in %s", path)
	}

	// Create a new formatter and format the AST
	formatter := ast.NewFormatter(formatFile)
	formattedContent := formatter.Format()

	if outputWriter != nil {
		_, err := outputWriter.Write([]byte(formattedContent))
		return err
	}
	// Write the formatted content back to the file
	return os.WriteFile(path, []byte(formattedContent), 0644)
}