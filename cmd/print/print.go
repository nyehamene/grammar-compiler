package printcmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"grammar/token"
	"io" // Added this import back
	"os"
	"strconv"
)

// escapeLexeme escapes a string for display, handling special characters.
func escapeLexeme(s string) string {
	buf := make([]rune, 0, len(s))
	for _, r := range s {
		switch r {
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\t':
			buf = append(buf, '\\', 't')
		case '\r':
			buf = append(buf, '\\', 'r')
		default:
			if strconv.IsPrint(r) {
				buf = append(buf, r)
			} else {
				buf = append(buf, []rune(fmt.Sprintf("\\u%04x", r))...)
			}
		}
	}
	return string(buf)
}

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
		} else {
			kind := "AST"
			if tokenFlag {
				kind = "token"
			}
			fmt.Fprintf(stdout, "Printing %s %s...\n", kind, path)
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

	tokenizer := token.NewTokenizer(srcRunes)
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
		line, col := findLineAndCol(tok.Start, srcRunes)
		lineCol := fmt.Sprintf("%d:%d", line, col)

		lexeme := tokenizer.Literal(tok, srcRunes)
		var formattedLexeme string
		if tok.Kind == token.String {
			formattedLexeme = lexeme
		} else {
			formattedLexeme = escapeLexeme(lexeme)
		}

		fmt.Fprintf(stdout, "%-*s %-*s %s\n", lineColWidth, lineCol, kindWidth, tok.Kind, formattedLexeme)
	}

	return nil
}

func findLineAndCol(offset int, srcRunes []rune) (int, int) {
	lineNum := 1
	lineStartOffset := 0
	for i, r := range srcRunes {
		if i == offset {
			break
		}
		if r == '\n' {
			lineNum++
			lineStartOffset = i + 1
		}
	}
	return lineNum, offset - lineStartOffset + 1
}