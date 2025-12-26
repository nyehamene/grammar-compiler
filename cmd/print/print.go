package printcmd

import (
	"flag"
	"fmt"
	"grammar/ast" // Import the ast package
	"grammar/command"
	"grammar/token"
	"io"
	"os"
)

func PrintCommand(args []string, stdout, stderr io.Writer) int {
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	var tokenFlag bool
	var astFlag bool
	printCmd.SetOutput(stderr)
	printCmd.BoolVar(&tokenFlag, "t", false, "Print the tokens produce by the tokenizer.")
	printCmd.BoolVar(&tokenFlag, "token", false, "Print the tokens produce by the tokenizer.")
	printCmd.BoolVar(&astFlag, "a", false, "Print the nodes produced by the parser.")
	printCmd.BoolVar(&astFlag, "ast", false, "Print the nodes produced by the parser.")
	var help bool
	printCmd.BoolVar(&help, "h", false, "Print this message.")
	printCmd.BoolVar(&help, "help", false, "Print this message.")
	printCmd.Parse(args)

	if help || printCmd.NArg() == 0 {
		fmt.Fprint(stdout, command.PrintUsage)
		return 0
	}

	for _, path := range printCmd.Args() {
		if tokenFlag {
			if err := printTokens(path, stdout); err != nil {
				fmt.Fprintf(stderr, "Error printing tokens for %s: %v\n", path, err)
				return 1
			}
		} else if astFlag {
			if err := printAST(path, stdout); err != nil {
				fmt.Fprintf(stderr, "Error printing AST for %s: %v\n", path, err)
				return 1
			}
		} else {
			// Default behavior if neither --tokens nor --ast is specified
			fmt.Fprintf(stdout, "Printing AST %s...\n", path)
		}
	}

	return 0
}

func printTokens(path string, stdout io.Writer) error {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	srcRunes := []rune(string(fileContent))

	tokenizer := token.NewTokenizer(srcRunes, false, false) // Do not skip comments or newlines
	tokens := tokenizer.Scan()

	// Define column widths
	const (
		lineColWidth = 10
		kindWidth    = 15
	)

	// Print header
	fmt.Fprintf(stdout, "%-*s %-*s %s\n", lineColWidth, "Line:Col", kindWidth, "KIND", "LEXEME")
	fmt.Fprintln(stdout, "--------------------------------------------------") // Adjust separator to match total width

	for _, tok := range tokens {
		line, col := token.FindLineAndCol(tok.Start, srcRunes)
		lineCol := fmt.Sprintf("%d:%d", line, col)

		lexeme := token.Literal(tok, srcRunes)
		var formattedLexeme string
		if tok.Kind == token.String {
			formattedLexeme = lexeme
		} else {
			formattedLexeme = token.EscapeLexeme(lexeme)
		}

		fmt.Fprintf(stdout, "%-*s %-*s %s\n", lineColWidth, lineCol, kindWidth, tok.Kind, formattedLexeme)
	}

	return nil
}

func printAST(path string, stdout io.Writer) error {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	srcRunes := []rune(string(fileContent))

	tokenizer := token.NewTokenizer(srcRunes, false, false) // Do not skip comments or newlines
	tokens := tokenizer.Scan()

	parser := ast.NewParser(tokens, srcRunes)
	astFile, err := parser.ParseFile()
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}

	printer := ast.NewPrinter(stdout, srcRunes)
	printer.PrintFile(astFile)

	return nil
}
