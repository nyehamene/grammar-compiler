package ast

import (
	"fmt"
	"io"
	"strings"
	"grammar/token" // Import token package
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
func (p *Printer) PrintFile(file *File) {
	p.Printf("File\n")
	p.indent++
	for _, decl := range file.Decls {
		p.PrintDecl(decl)
	}
	p.indent--
}

// PrintDecl prints a declaration node.
func (p *Printer) PrintDecl(decl Decl) {
	line, col := token.FindLineAndCol(int(decl.Pos()), p.srcRunes)
	fmt.Fprintf(p.output, "%d:%-4d", line, col) // Print line/col with padding
	p.Printf("")  // This will print the indentation

	switch n := decl.(type) {
	case *RuleDecl:
		p.Printf("RuleDecl: %s\n", n.Name.Name)
		p.indent++
		for _, expr := range n.Body {
			p.PrintExpr(expr)
		}
		p.indent--
	case *BindingDecl:
		pathValue := strings.Trim(n.Path.Value, "\"")
		p.Printf("BindingDecl: %s = @import(\"%s\")\n", n.Name.Name, pathValue)
	default:
		p.Printf("Unknown Decl Type: %T\n", n)
	}
}

// PrintExpr prints an expression node.
func (p *Printer) PrintExpr(expr Expr) {
	line, col := token.FindLineAndCol(int(expr.Pos()), p.srcRunes)
	fmt.Fprintf(p.output, "%d:%-4d", line, col) // Print line/col with padding
	p.Printf("")  // This will print the indentation

	switch n := expr.(type) {
	case *Ident:
		p.Printf("Ident: %s\n", n.Name)
	case *StringLit:
		p.Printf("StringLit: %s\n", n.Value)
	case *RegexLit:
		p.Printf("RegexLit: %s\n", n.Value)
	case *AlternativeExpr:
		p.Printf("AlternativeExpr\n")
		p.indent++
		for _, altExpr := range n.Exprs {
			p.PrintExpr(altExpr)
		}
		p.indent--
	case *OptionalExpr:
		p.Printf("OptionalExpr\n")
		p.indent++
		p.PrintExpr(n.Expr)
		p.indent--
	case *RepetitionExpr:
		p.Printf("RepetitionExpr\n")
		p.indent++
		p.PrintExpr(n.Expr)
		p.indent--
	case *GroupExpr:
		p.Printf("GroupExpr\n")
		p.indent++
		p.PrintExpr(n.Expr)
		p.indent--
	case *TermExpr:
		p.Printf("TermExpr\n")
		p.indent++
		p.PrintExpr(n.X)
		p.indent--
	case *DirectiveExpr:
		p.Printf("DirectiveExpr: @%s\n", n.Name.Name)
		p.indent++
		for _, arg := range n.Args {
			p.PrintExpr(arg)
		}
		p.indent--
	case *MemberExpr:
		p.Printf("MemberExpr\n")
		p.indent++
		p.PrintExpr(n.Object)
		p.Printf("Member: %s\n", n.Member.Name)
		p.indent--
	default:
		p.Printf("Unknown Expr Type: %T\n", n)
	}
}

// Printf is a helper to print with current indentation.
func (p *Printer) Printf(format string, args ...interface{}) {
	for i := 0; i < p.indent; i++ {
		fmt.Fprint(p.output, p.indentStr)
	}
	fmt.Fprintf(p.output, format, args...)
}
