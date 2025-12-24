package cmd

import (
	"flag"
	"fmt"
	"grammar/ast"
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
		if err := diffAST(path1, path2, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "Error diffing ASTs: %v\n", err)
			return 1
		}
	} else {
		// Default behavior
		fmt.Fprintf(stdout, "Diffing %s and %s\n", path1, path2)
	}

	return 0
}

func diffAST(path1, path2 string, stdout, stderr io.Writer) error {
	// Parse file 1
	content1, err := os.ReadFile(path1)
	if err != nil {
		return err
	}
	srcRunes1 := []rune(string(content1))
	tokenizer1 := token.NewTokenizer(srcRunes1)
	tokens1 := tokenizer1.Scan()
	parser1 := ast.NewParser(tokens1, srcRunes1)
	ast1, err := parser1.ParseFile()
	if err != nil {
		if errs, ok := err.(ast.ErrorList); ok {
			for _, e := range errs {
				line, col := token.FindLineAndCol(int(e.Pos), srcRunes1)
				fmt.Fprintf(stderr, "%s:%d:%d: %s\n", path1, line, col, e.Message)
			}
		} else {
			fmt.Fprintf(stderr, "Error parsing file %s: %s\n", path1, err)
		}
		return fmt.Errorf("parsing error in %s", path1)
	}

	// Parse file 2
	content2, err := os.ReadFile(path2)
	if err != nil {
		return err
	}
	srcRunes2 := []rune(string(content2))
	tokenizer2 := token.NewTokenizer(srcRunes2)
	tokens2 := tokenizer2.Scan()
	parser2 := ast.NewParser(tokens2, srcRunes2)
	ast2, err := parser2.ParseFile()
	if err != nil {
		if errs, ok := err.(ast.ErrorList); ok {
			for _, e := range errs {
				line, col := token.FindLineAndCol(int(e.Pos), srcRunes2)
				fmt.Fprintf(stderr, "%s:%d:%d: %s\n", path2, line, col, e.Message)
			}
		} else {
			fmt.Fprintf(stderr, "Error parsing file %s: %s\n", path2, err)
		}
		return fmt.Errorf("parsing error in %s", path2)
	}

	// Simple comparison of declarations
	maxLen := len(ast1.Decls)
	if len(ast2.Decls) > maxLen {
		maxLen = len(ast2.Decls)
	}

	printer1 := ast.NewPrinter(stdout, srcRunes1)
	printer2 := ast.NewPrinter(stdout, srcRunes2)

	for i := 0; i < maxLen; i++ {
		var decl1, decl2 ast.Decl

		if i < len(ast1.Decls) {
			decl1 = ast1.Decls[i]
		}
		if i < len(ast2.Decls) {
			decl2 = ast2.Decls[i]
		}

		if decl1 == nil {
			fmt.Fprint(stdout, "+ ")
			printer2.PrintDecl(decl2)
			continue
		}
		if decl2 == nil {
			fmt.Fprint(stdout, "- ")
			printer1.PrintDecl(decl1)
			continue
		}

		if !areDeclsEqual(decl1, decl2, srcRunes1, srcRunes2) {
			fmt.Fprint(stdout, "- ")
			printer1.PrintDecl(decl1)
			fmt.Fprint(stdout, "+ ")
			printer2.PrintDecl(decl2)
		} else {
			fmt.Fprint(stdout, "  ")
			printer1.PrintDecl(decl1)
		}
	}

	return nil
}

// areDeclsEqual structurally compares two ast.Decl nodes.
func areDeclsEqual(d1, d2 ast.Decl, src1, src2 []rune) bool {
	if fmt.Sprintf("%T", d1) != fmt.Sprintf("%T", d2) {
		return false
	}

	switch n1 := d1.(type) {
	case *ast.RuleDecl:
		n2 := d2.(*ast.RuleDecl)
		if n1.Name.Name != n2.Name.Name {
			return false
		}
		if !areExprListsEqual(n1.Body, n2.Body, src1, src2) {
			return false
		}
		return true
	case *ast.BindingDecl:
		n2 := d2.(*ast.BindingDecl)
		if n1.Name.Name != n2.Name.Name {
			return false
		}
		if n1.Path.Value != n2.Path.Value {
			return false
		}
		return true
	}
	return false
}

// areExprListsEqual compares two slices of ast.Expr recursively.
func areExprListsEqual(list1, list2 []ast.Expr, src1, src2 []rune) bool {
	if len(list1) != len(list2) {
		return false
	}
	for i := range list1 {
		if !areExprsEqual(list1[i], list2[i], src1, src2) {
			return false
		}
	}
	return true
}

// areExprsEqual compares two ast.Expr nodes recursively.
func areExprsEqual(e1, e2 ast.Expr, src1, src2 []rune) bool {
	if fmt.Sprintf("%T", e1) != fmt.Sprintf("%T", e2) {
		return false
	}

	switch n1 := e1.(type) {
	case *ast.Ident:
		n2 := e2.(*ast.Ident)
		return n1.Name == n2.Name
	case *ast.StringLit:
		n2 := e2.(*ast.StringLit)
		return n1.Value == n2.Value
	case *ast.RegexLit:
		n2 := e2.(*ast.RegexLit)
		return n1.Value == n2.Value
	case *ast.AlternativeExpr:
		n2 := e2.(*ast.AlternativeExpr)
		return areExprListsEqual(n1.Exprs, n2.Exprs, src1, src2)
	case *ast.OptionalExpr:
		n2 := e2.(*ast.OptionalExpr)
		return areExprsEqual(n1.Expr, n2.Expr, src1, src2)
	case *ast.RepetitionExpr:
		n2 := e2.(*ast.RepetitionExpr)
		return areExprsEqual(n1.Expr, n2.Expr, src1, src2)
	case *ast.GroupExpr:
		n2 := e2.(*ast.GroupExpr)
		return areExprsEqual(n1.Expr, n2.Expr, src1, src2)
	case *ast.TermExpr:
		n2 := e2.(*ast.TermExpr)
		return areExprsEqual(n1.X, n2.X, src1, src2)
	case *ast.DirectiveExpr:
		n2 := e2.(*ast.DirectiveExpr)
		if n1.Name.Name != n2.Name.Name {
			return false
		}
		return areExprListsEqual(n1.Args, n2.Args, src1, src2)
	case *ast.MemberExpr:
		n2 := e2.(*ast.MemberExpr)
		return areExprsEqual(n1.Object, n2.Object, src1, src2) && areExprsEqual(n1.Member, n2.Member, src1, src2)
	}
	return false
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


// areDeclsEqual structurally compares two ast.Decl nodes.
func areDeclsEqual(d1, d2 ast.Decl, src1, src2 []rune) bool {
	if fmt.Sprintf("%T", d1) != fmt.Sprintf("%T", d2) {
		return false
	}

	switch n1 := d1.(type) {
	case *ast.RuleDecl:
		n2 := d2.(*ast.RuleDecl)
		if n1.Name.Name != n2.Name.Name {
			return false
		}
		if !areExprListsEqual(n1.Body, n2.Body, src1, src2) {
			return false
		}
		return true
	case *ast.BindingDecl:
		n2 := d2.(*ast.BindingDecl)
		if n1.Name.Name != n2.Name.Name {
			return false
		}
		if n1.Path.Value != n2.Path.Value {
			return false
		}
		return true
	}
	return false
}

// areExprListsEqual compares two slices of ast.Expr recursively.
func areExprListsEqual(list1, list2 []ast.Expr, src1, src2 []rune) bool {
	if len(list1) != len(list2) {
		return false
	}
	for i := range list1 {
		if !areExprsEqual(list1[i], list2[i], src1, src2) {
			return false
		}
	}
	return true
}

// areExprsEqual compares two ast.Expr nodes recursively.
func areExprsEqual(e1, e2 ast.Expr, src1, src2 []rune) bool {
	if fmt.Sprintf("%T", e1) != fmt.Sprintf("%T", e2) {
		return false
	}

	switch n1 := e1.(type) {
	case *ast.Ident:
		n2 := e2.(*ast.Ident)
		return n1.Name == n2.Name
	case *ast.StringLit:
		n2 := e2.(*ast.StringLit)
		return n1.Value == n2.Value
	case *ast.RegexLit:
		n2 := e2.(*ast.RegexLit)
		return n1.Value == n2.Value
	case *ast.AlternativeExpr:
		n2 := e2.(*ast.AlternativeExpr)
		return areExprListsEqual(n1.Exprs, n2.Exprs, src1, src2)
	case *ast.OptionalExpr:
		n2 := e2.(*ast.OptionalExpr)
		return areExprsEqual(n1.Expr, n2.Expr, src1, src2)
	case *ast.RepetitionExpr:
		n2 := e2.(*ast.RepetitionExpr)
		return areExprsEqual(n1.Expr, n2.Expr, src1, src2)
	case *ast.GroupExpr:
		n2 := e2.(*ast.GroupExpr)
		return areExprsEqual(n1.Expr, n2.Expr, src1, src2)
	case *ast.TermExpr:
		n2 := e2.(*ast.TermExpr)
		return areExprsEqual(n1.X, n2.X, src1, src2)
	case *ast.DirectiveExpr:
		n2 := e2.(*ast.DirectiveExpr)
		if n1.Name.Name != n2.Name.Name {
			return false
		}
		return areExprListsEqual(n1.Args, n2.Args, src1, src2)
	case *ast.MemberExpr:
		n2 := e2.(*ast.MemberExpr)
		return areExprsEqual(n1.Object, n2.Object, src1, src2) && areExprsEqual(n1.Member, n2.Member, src1, src2)
	}
	return false
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