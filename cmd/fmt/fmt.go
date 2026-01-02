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

func FmtCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fmtCmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	var help bool
	var stdoutFlag bool
	var stdinFlag bool
	fmtCmd.SetOutput(stderr)
	fmtCmd.BoolVar(&help, "h", false, "Print this message.")
	fmtCmd.BoolVar(&help, "help", false, "Print this message.")
	fmtCmd.BoolVar(&stdoutFlag, "stdout", false, "Print formatted output to stdout. Do not modify input file.")
	fmtCmd.BoolVar(&stdinFlag, "stdin", false, "Format code from stdin, output to stdout.")
	_ = fmtCmd.Parse(args)

	if help {
		if _, err := fmt.Fprint(stdout, command.FmtUsage); err != nil {
			return 1
		}
		return 0
	}

	if stdinFlag {
		err := formatReader(stdin, stdout, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "Error formatting from stdin: %v\n", err)
			return 1
		}
		return 0
	}

	if fmtCmd.NArg() == 0 {
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

func formatSource(srcRunes []rune, sourceName string, stderr io.Writer) (string, error) {
	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()

	formatterParser := ast.NewFormatterParser(tokens, srcRunes)
	formatFile, err := formatterParser.Parse()
	if err != nil {
		if errs, ok := err.(ast.ErrorList); ok {
			for _, e := range errs {
				line, col := token.FindLineAndCol(e.Pos, srcRunes)
				fmt.Fprintf(stderr, "%s:%d:%d: %s\n", sourceName, line, col, e.Message)
			}
		} else {
			fmt.Fprintf(stderr, "Error parsing from %s: %s\n", sourceName, err)
		}
		return "", fmt.Errorf("parsing error from %s", sourceName)
	}

	formatter := ast.NewFormatter(formatFile)
	formattedContent := formatter.Format()
	return formattedContent, nil
}

func formatReader(in io.Reader, out, stderr io.Writer) error {
	fileContent, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	// if the content is empty, do nothing
	if len(fileContent) == 0 {
		return nil
	}
	srcRunes := []rune(string(fileContent))

	formattedContent, err := formatSource(srcRunes, "stdin", stderr)
	if err != nil {
		return err
	}

	_, err = out.Write([]byte(formattedContent))
	return err
}

func formatFile(path string, stdout, stderr io.Writer, outputWriter io.Writer) error {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	srcRunes := []rune(string(fileContent))

	formattedContent, err := formatSource(srcRunes, path, stderr)
	if err != nil {
		return err
	}

	if outputWriter != nil {
		_, err := outputWriter.Write([]byte(formattedContent))
		return err
	}
	// Write the formatted content back to the file
	return os.WriteFile(path, []byte(formattedContent), 0644)
}
