package cmd

import (
	"flag"
	"fmt"
	"grammar/command"
	"grammar/token"
	"io"
	"os"
	"strings"
)

func DiffCommand(args []string, stdout, stderr io.Writer) int {
	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	var tokenFlag bool
	var astFlag bool
	diffCmd.SetOutput(stderr)
	diffCmd.BoolVar(&tokenFlag, "t", false, "Print differences in tokens.")
	diffCmd.BoolVar(&tokenFlag, "tokens", false, "Print differences in tokens.")
	diffCmd.BoolVar(&astFlag, "a", false, "Print differences in ast nodes.")
	diffCmd.BoolVar(&astFlag, "ast", false, "Print differences in ast nodes.")
	var help bool
	diffCmd.BoolVar(&help, "h", false, "Print this message.")
	diffCmd.BoolVar(&help, "help", false, "Print this message.")
	diffCmd.Parse(args)
	if help || diffCmd.NArg() != 2 {
		fmt.Fprint(stdout, command.DiffUsage)
		return 0
	}

	path1 := diffCmd.Arg(0)
	path2 := diffCmd.Arg(1)

	if tokenFlag {
		if err := diffTokens(path1, path2, stdout); err != nil {
			fmt.Fprintf(stderr, "Error diffing tokens: %v\n", err)
			return 1
		}
	} else if astFlag {
		fmt.Fprintf(stdout, "Diffing AST for %s and %s\n", path1, path2)
	} else {
		// Default behavior
		fmt.Fprintf(stdout, "Diffing %s and %s\n", path1, path2)
	}

	return 0
}

func diffTokens(path1, path2 string, w io.Writer) error {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return err
	}
	content2, err := os.ReadFile(path2)
	if err != nil {
		return err
	}

	srcRunes1 := []rune(string(content1))
	tokenizer1 := token.NewTokenizer(srcRunes1)
	tokens1 := tokenizer1.Scan()

	srcRunes2 := []rune(string(content2))
	tokenizer2 := token.NewTokenizer(srcRunes2)
	tokens2 := tokenizer2.Scan()

	// LCS-based diff
	lcs := computeLCS(tokens1, tokens2, srcRunes1, srcRunes2)

	// Print header
	const (
		lineColWidth = 10
		kindWidth    = 15
	)
	fmt.Fprintf(w, "  %-*s %-*s %s\n", lineColWidth, "Line:Col", kindWidth, "KIND", "LEXEME")
	fmt.Fprintln(w, "  "+strings.Repeat("-", lineColWidth+kindWidth+20)) // Adjust separator

	i, j := 0, 0
	for i < len(tokens1) && j < len(tokens2) {
		if tokensEqual(tokens1[i], tokens2[j], srcRunes1, srcRunes2) {
			printToken(w, ' ', tokens1[i], srcRunes1)
			i++
			j++
		} else {
			// Find which token is not in LCS
			inLCS1 := false
			for _, t := range lcs {
				if tokensEqual(tokens1[i], t, srcRunes1, srcRunes1) { // A bit inefficient
					inLCS1 = true
					break
				}
			}
			inLCS2 := false
			for _, t := range lcs {
				if tokensEqual(tokens2[j], t, srcRunes2, srcRunes2) {
					inLCS2 = true
					break
				}
			}

			if !inLCS1 {
				printToken(w, '-', tokens1[i], srcRunes1)
				i++
			}
			if !inLCS2 {
				printToken(w, '+', tokens2[j], srcRunes2)
				j++
			}
		}
	}

	// Print remaining tokens
	for ; i < len(tokens1); i++ {
		printToken(w, '-', tokens1[i], srcRunes1)
	}
	for ; j < len(tokens2); j++ {
		printToken(w, '+', tokens2[j], srcRunes2)
	}

	return nil
}

func printToken(w io.Writer, prefix rune, tok token.Token, srcRunes []rune) {
	const (
		lineColWidth = 10
		kindWidth    = 15
	)
	line, col := token.FindLineAndCol(tok.Start, srcRunes)
	lineCol := fmt.Sprintf("%d:%d", line, col)
	lexeme := token.EscapeLexeme(token.Literal(tok, srcRunes))
	if tok.Kind == token.String {
		lexeme = token.Literal(tok, srcRunes)
	}
	fmt.Fprintf(w, "%c %-*s %-*s %s\n", prefix, lineColWidth, lineCol, kindWidth, tok.Kind, lexeme)
}

func computeLCS(tokens1, tokens2 []token.Token, src1, src2 []rune) []token.Token {
	m, n := len(tokens1), len(tokens2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if tokensEqual(tokens1[i-1], tokens2[j-1], src1, src2) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Reconstruct LCS from dp table
	var lcs []token.Token
	i, j := m, n
	for i > 0 && j > 0 {
		if tokensEqual(tokens1[i-1], tokens2[j-1], src1, src2) {
			lcs = append(lcs, tokens1[i-1])
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	// Reverse LCS to get correct order
	for k, l := 0, len(lcs)-1; k < l; k, l = k+1, l-1 {
		lcs[k], lcs[l] = lcs[l], lcs[k]
	}
	return lcs
}

func tokensEqual(t1, t2 token.Token, src1, src2 []rune) bool {
	return t1.Kind == t2.Kind && token.Literal(t1, src1) == token.Literal(t2, src2)
}

