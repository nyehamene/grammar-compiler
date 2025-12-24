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
	p.printf("File\n")
	p.indent++
	for _, decl := range file.Decls {
		p.printDecl(decl)
	}
	p.indent--
}


func (p *Printer) printDecl(decl Decl) {
	line, col := token.FindLineAndCol(int(decl.Pos()), p.srcRunes)
	fmt.Fprintf(p.output, "%d:%-4d", line, col) // Print line/col with padding
	p.printf("")  // This will print the indentation

	switch n := decl.(type) {
	case *RuleDecl:
		p.printf("RuleDecl: %s\n", n.Name.Name)
		p.indent++
		for _, expr := range n.Body {
			p.printExpr(expr)
		}
		p.indent--
	case *BindingDecl:
		pathValue := strings.Trim(n.Path.Value, "\"")
		p.printf("BindingDecl: %s = @import(\"%s\")\n", n.Name.Name, pathValue)
	default:
		p.printf("Unknown Decl Type: %T\n", n)
	}
}

func (p *Printer) printExpr(expr Expr) {
	line, col := token.FindLineAndCol(int(expr.Pos()), p.srcRunes)
	fmt.Fprintf(p.output, "%d:%-4d", line, col) // Print line/col with padding
	p.printf("")  // This will print the indentation

	switch n := expr.(type) {
	case *Ident:
		p.printf("Ident: %s\n", n.Name)
	case *StringLit:
		p.printf("StringLit: %s\n", n.Value)
	case *RegexLit:
		p.printf("RegexLit: %s\n", n.Value)
	case *AlternativeExpr:
		p.printf("AlternativeExpr\n")
		p.indent++
		for _, altExpr := range n.Exprs {
			p.printExpr(altExpr)
		}
		p.indent--
	case *OptionalExpr:
		p.printf("OptionalExpr\n")
		p.indent++
		p.printExpr(n.Expr)
		p.indent--
	case *RepetitionExpr:
		p.printf("RepetitionExpr\n")
		p.indent++
		p.printExpr(n.Expr)
		p.indent--
	case *GroupExpr:
		p.printf("GroupExpr\n")
		p.indent++
		p.printExpr(n.Expr)
		p.indent--
	case *TermExpr:
		p.printf("TermExpr\n")
		p.indent++
		p.printExpr(n.X)
		p.indent--
	case *DirectiveExpr:
		p.printf("DirectiveExpr: @%s\n", n.Name.Name)
		p.indent++
		for _, arg := range n.Args {
			p.printExpr(arg)
		}
		p.indent--
	case *MemberExpr:
		p.printf("MemberExpr\n")
		p.indent++
		p.printExpr(n.Object)
		p.printf("Member: %s\n", n.Member.Name)
		p.indent--
	default:
		p.printf("Unknown Expr Type: %T\n", n)
	}
}

// printf is a helper to print with current indentation.
func (p *Printer) printf(format string, args ...interface{}) {
	for i := 0; i < p.indent; i++ {
		fmt.Fprint(p.output, p.indentStr)
	}
	fmt.Fprintf(p.output, format, args...)
}
