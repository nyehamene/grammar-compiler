package ast

import (
	"fmt"
	"grammar/token" // Import token package
	"io"
	"strings"
)

// Printer holds the state for printing the AST.
type Printer struct {
	output    io.Writer
	indent    int
	indentStr string
	srcRunes  []rune // Add srcRunes to Printer
}

// NewPrinter creates a new AST printer.
func NewPrinter(output io.Writer, srcRunes []rune) *Printer {
	return &Printer{
		output:    output,
		indentStr: "  ", // 2 spaces per indent level
		srcRunes:  srcRunes,
	}
}

// PrintFile prints the entire File AST.
func (p *Printer) PrintFile(file *File) error {
	if err := p.Printf("File\n"); err != nil {
		return err
	}
	p.indent++
	for _, decl := range file.Decls {
		if err := p.PrintDecl(decl); err != nil {
			return err
		}
	}
	p.indent--
	return nil
}

// PrintDecl prints a declaration node.
func (p *Printer) PrintDecl(decl Decl) error {
	line, col := token.FindLineAndCol(int(decl.Pos()), p.srcRunes)
	if _, err := fmt.Fprintf(p.output, "%d:%-4d", line, col); err != nil { // Print line/col with padding
		return err
	}
	if err := p.Printf(""); err != nil { // This will print the indentation
		return err
	}

	switch n := decl.(type) {
	case *RuleDecl:
		if err := p.Printf("RuleDecl: %s\n", n.Name.Name); err != nil {
			return err
		}
		p.indent++
		for _, expr := range n.Body {
			if err := p.PrintExpr(expr); err != nil {
				return err
			}
		}
		p.indent--
	case *BindingDecl:
		pathValue := strings.Trim(n.Path.Value, "\"")
		if err := p.Printf("BindingDecl: %s = @import(\"%s\")\n", n.Name.Name, pathValue); err != nil {
			return err
		}
	case *CommentGroup:
		for _, comment := range n.List {
			if _, err := fmt.Fprint(p.output, comment.Text); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(p.output, "\n"); err != nil {
			return err
		}
	default:
		if err := p.Printf("Unknown Decl Type: %T\n", n); err != nil {
			return err
		}
	}
	return nil
}

// PrintExpr prints an expression node.
func (p *Printer) PrintExpr(expr Expr) error {
	line, col := token.FindLineAndCol(int(expr.Pos()), p.srcRunes)
	if _, err := fmt.Fprintf(p.output, "%d:%-4d", line, col); err != nil { // Print line/col with padding
		return err
	}
	if err := p.Printf(""); err != nil { // This will print the indentation
		return err
	}

	switch n := expr.(type) {
	case *Ident:
		if err := p.Printf("Ident: %s\n", n.Name); err != nil {
			return err
		}
	case *StringLit:
		if err := p.Printf("StringLit: %s\n", n.Value); err != nil {
			return err
		}
	case *RegexLit:
		if err := p.Printf("RegexLit: %s\n", n.Value); err != nil {
			return err
		}
	case *ExternalValue: // Added for completeness, after previous steps
		if err := p.Printf("ExternalValue: $%s\n", n.Name); err != nil {
			return err
		}
	case *AlternativeExpr:
		if err := p.Printf("AlternativeExpr\n"); err != nil {
			return err
		}
		p.indent++
		for _, altExpr := range n.Exprs {
			if err := p.PrintExpr(altExpr); err != nil {
				return err
			}
		}
		p.indent--
	case *OptionalExpr:
		if err := p.Printf("OptionalExpr\n"); err != nil {
			return err
		}
		p.indent++
		if err := p.PrintExpr(n.Expr); err != nil {
			return err
		}
		p.indent--
	case *RepetitionExpr:
		if err := p.Printf("RepetitionExpr\n"); err != nil {
			return err
		}
		p.indent++
		if err := p.PrintExpr(n.Expr); err != nil {
			return err
		}
		p.indent--
	case *GroupExpr:
		if err := p.Printf("GroupExpr\n"); err != nil {
			return err
		}
		p.indent++
		if err := p.PrintExpr(n.Expr); err != nil {
			return err
		}
		p.indent--
	case *TermExpr:
		if err := p.Printf("TermExpr\n"); err != nil {
			return err
		}
		p.indent++
		if err := p.PrintExpr(n.X); err != nil {
			return err
		}
		p.indent--
	case *DirectiveExpr:
		if err := p.Printf("DirectiveExpr: @%s\n", n.Name.Name); err != nil {
			return err
		}
		p.indent++
		for _, arg := range n.Args {
			if err := p.PrintExpr(arg); err != nil {
				return err
			}
		}
		p.indent--
	case *MemberExpr:
		if err := p.Printf("MemberExpr\n"); err != nil {
			return err
		}
		p.indent++
		if err := p.PrintExpr(n.Object); err != nil {
			return err
		}
		if err := p.Printf("Member: %s\n", n.Member.Name); err != nil {
			return err
		}
		p.indent--
	default:
		if err := p.Printf("Unknown Expr Type: %T\n", n); err != nil {
			return err
		}
	}
	return nil
}

// Printf is a helper to print with current indentation.
func (p *Printer) Printf(format string, args ...interface{}) error {
	for i := 0; i < p.indent; i++ {
		if _, err := fmt.Fprint(p.output, p.indentStr); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(p.output, format, args...)
	return err
}
