package ast

import (
	"bytes"
	"grammar/token"
	"strings"
)

type Formatter struct {
	file   *FormatFile
	buffer bytes.Buffer
}

type FormatterParser struct {
	tokens   []token.Token
	srcRunes []rune
	parser   *Parser
}

type FormatGroup struct {
	Decls []Decl
}

type FormatFile struct {
	Decls       []FormatGroup
	LineOffsets []int
}

func NewFormatter(file *FormatFile) *Formatter {
	return &Formatter{
		file: file,
	}
}

func NewFormatterParser(tokens []token.Token, srcRunes []rune) *FormatterParser {
	return &FormatterParser{
		tokens:   tokens,
		srcRunes: srcRunes,
		parser:   NewParser(tokens, srcRunes),
	}
}

func (fp *FormatterParser) Parse() (*FormatFile, error) {
	var groups []FormatGroup
	var currentGroup []Decl

	// Compute line offsets
	lineOffsets := []int{0}
	for i, r := range fp.srcRunes {
		if r == '\n' {
			lineOffsets = append(lineOffsets, i+1)
		}
	}

	for fp.parser.peek().Kind != token.EOF {

		decl, hasBlankLine := fp.parser.parseDecl()
		if decl == nil {
			if len(fp.parser.errors) > 0 {
				return nil, fp.parser.errors
			}
			break
		}

		if hasBlankLine && len(currentGroup) > 0 {
			groups = append(groups, FormatGroup{Decls: currentGroup})
			currentGroup = []Decl{}
		}
		currentGroup = append(currentGroup, decl)
	}

	if len(currentGroup) > 0 {
		groups = append(groups, FormatGroup{Decls: currentGroup})
	}

	if len(fp.parser.errors) > 0 {
		return nil, fp.parser.errors
	}

	return &FormatFile{Decls: groups, LineOffsets: lineOffsets}, nil
}

func (f *Formatter) Format() string {
	for i, group := range f.file.Decls {
		f.formatDeclGroup(group.Decls)
		if i < len(f.file.Decls)-1 {
			f.buffer.WriteString("\n")
		}
	}
	return f.buffer.String()
}

func (f *Formatter) formatDeclGroup(group []Decl) {
	maxNameLen := 0
	for _, decl := range group {
		switch n := decl.(type) {
		case *RuleDecl:
			if len(n.Name.Name) > maxNameLen {
				maxNameLen = len(n.Name.Name)
			}
		case *BindingDecl:
			if len(n.Name.Name) > maxNameLen {
				maxNameLen = len(n.Name.Name)
			}
		}
	}

	for _, decl := range group {
		cmt := false
		switch n := decl.(type) {
		case *RuleDecl:
			ruleLine := f.findLine(decl.Pos())
			f.buffer.WriteString(n.Name.Name)
			f.buffer.WriteString(strings.Repeat(" ", maxNameLen-len(n.Name.Name)))
			f.buffer.WriteString(" = ")
			for i, expr := range n.Body {
				if i > 0 && f.findLine(expr.Pos()) == ruleLine {
					f.buffer.WriteString(" ")
				}
				f.formatExpr(expr, ruleLine, maxNameLen)
			}
			// Handle trailing semicolon alignment
			lastExpr := n.Body[len(n.Body)-1]
			lastExprLine := f.findLine(lastExpr.End() - 1) // Get line of the last character
			if lastExprLine > ruleLine {                   // If alternatives were vertically aligned
				f.buffer.WriteString("\n")
				f.buffer.WriteString(strings.Repeat(" ", maxNameLen+1))
			}
			f.buffer.WriteString(";")
		case *BindingDecl:
			f.buffer.WriteString(n.Name.Name)
			f.buffer.WriteString(strings.Repeat(" ", maxNameLen-len(n.Name.Name)))
			f.buffer.WriteString(" = ")
			f.buffer.WriteString("@import(")
			f.buffer.WriteString(n.Path.Value)
			f.buffer.WriteString(");")
		case *CommentGroup:
			cmt = true
			for _, comment := range n.List {
				f.buffer.WriteString(comment.Text)
			}
		}
		if !cmt {
			f.buffer.WriteString("\n")
		}
	}
}

func (f *Formatter) formatExpr(expr Expr, ruleLine int, maxLen int) {
	switch n := expr.(type) {
	case *Ident:
		f.buffer.WriteString(n.Name)
	case *StringLit:
		f.buffer.WriteString(n.Value)
	case *RegexLit:
		f.buffer.WriteString(n.Value)
	case *AlternativeExpr:
		f.formatAlternativeExpr(n, ruleLine, maxLen)
	case *OptionalExpr:
		f.formatOptionalExpr(n, ruleLine, maxLen)
	case *RepetitionExpr:
		f.formatRepetitionExpr(n, ruleLine, maxLen)
	case *GroupExpr:
		f.formatGroupExpr(n, ruleLine, maxLen)
	case *SequenceExpr:
		f.formatSequenceExpr(n, ruleLine, maxLen)
	case *TermExpr:
		f.formatTermExpr(n, ruleLine, maxLen)
	case *DirectiveExpr:
		f.formatDirectiveExpr(n, ruleLine, maxLen)
	case *MemberExpr:
		f.formatMemberExpr(n, ruleLine, maxLen)
	}
}

// findLine takes a token.Pos and returns the line number (1-based).
func (f *Formatter) findLine(pos token.Pos) int {
	// Binary search to find the line number
	low, high := 0, len(f.file.LineOffsets)-1
	lineNum := 0
	for low <= high {
		mid := low + (high-low)/2
		if f.file.LineOffsets[mid] <= int(pos) {
			lineNum = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return lineNum + 1
}

func (f *Formatter) formatAlternativeExpr(n *AlternativeExpr, ruleLine int, maxLen int) {
	for i, expr := range n.Exprs {
		if i > 0 {
			exprLine := f.findLine(expr.Pos())
			if exprLine == ruleLine { // If on the same line, just add " | "
				f.buffer.WriteString(" | ")
			} else { // If on a new line, it's already handled by formatExpr
				f.buffer.WriteString("\n")
				f.buffer.WriteString(strings.Repeat(" ", maxLen))
				f.buffer.WriteString(" | ")
			}
		}
		f.formatExpr(expr, ruleLine, maxLen)
	}
}

func (f *Formatter) formatOptionalExpr(n *OptionalExpr, ruleLine int, maxLen int) {
	f.buffer.WriteString("[ ")
	f.formatExpr(n.Expr, ruleLine, maxLen)
	f.buffer.WriteString(" ]")
}

func (f *Formatter) formatRepetitionExpr(n *RepetitionExpr, ruleLine int, maxLen int) {
	f.buffer.WriteString("{ ")
	f.formatExpr(n.Expr, ruleLine, maxLen)
	f.buffer.WriteString(" }")
}

func (f *Formatter) formatGroupExpr(n *GroupExpr, ruleLine int, maxLen int) {
	f.buffer.WriteString("( ")
	f.formatExpr(n.Expr, ruleLine, maxLen)
	f.buffer.WriteString(" )")
}

func (f *Formatter) formatTermExpr(n *TermExpr, ruleLine int, maxLen int) {
	f.formatExpr(n.X, ruleLine, maxLen)
}

func (f *Formatter) formatSequenceExpr(n *SequenceExpr, ruleLine int, maxLen int) {
	for i, expr := range n.Exprs {
		if i > 0 {
			f.buffer.WriteString(" ")
		}
		f.formatExpr(expr, ruleLine, maxLen)
	}
}

func (f *Formatter) formatDirectiveExpr(n *DirectiveExpr, ruleLine int, maxLen int) {
	f.buffer.WriteString("@")
	f.buffer.WriteString(n.Name.Name)
	f.buffer.WriteString("(")
	for i, arg := range n.Args {
		if i > 0 {
			f.buffer.WriteString(", ")
		}
		f.formatExpr(arg, ruleLine, maxLen)
	}
	f.buffer.WriteString(")")
}

func (f *Formatter) formatMemberExpr(n *MemberExpr, ruleLine int, maxLen int) {
	f.formatExpr(n.Object, ruleLine, maxLen)
	f.buffer.WriteString(".")
	f.buffer.WriteString(n.Member.Name)
}
