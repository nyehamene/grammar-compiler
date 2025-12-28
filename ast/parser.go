package ast

import (
	"fmt"
	"grammar/token"
)

// Parser holds the parser's state.
type Parser struct {
	tokens       []token.Token
	srcRunes     []rune
	pos          int // current position in the token slice
	errors       ErrorList
	hadBlankLine bool // Add hadBlankLine field
}

// NewParser creates a new Parser.
func NewParser(tokens []token.Token, srcRunes []rune) *Parser {
	return &Parser{tokens: tokens, srcRunes: srcRunes}
}

// ParseFile parses the input tokens and returns a File node.
func (p *Parser) ParseFile() (*File, error) {
	var decls []Decl
	for p.peek().Kind != token.EOF {
		if decl, _ := p.parseDecl(); decl != nil {
			decls = append(decls, decl)
		}
	}
	endPos := p.peek().End

	if len(p.errors) > 0 {
		return nil, p.errors
	}

	return &File{Decls: decls, EndPos: token.Pos(endPos)}, nil
}

func (p *Parser) parseDecl() (Decl, bool) {
	hadBlankLine := p.hadBlankLine
	p.hadBlankLine = false

	if p.peek().Kind == token.Comment {
		return p.parseCommentGroup(), hadBlankLine
	}

	if p.peek().Kind != token.Ident {
		tok := p.next()
		p.errorf(token.Pos(tok.Start), "expected IDENT, got %s", tok.Kind)
		return nil, hadBlankLine
	}

	ident := p.expect(token.Ident)
	p.expect(token.Assign)
	if p.peek().Kind == token.AtDirective {
		return p.parseBinding(ident), hadBlankLine
	}
	return p.parseRule(ident), hadBlankLine
}

func (p *Parser) parseCommentGroup() *CommentGroup {
	var list []*Comment
	for p.peek().Kind == token.Comment {
		tok := p.next()
		list = append(list, &Comment{Slash: token.Pos(tok.Start), Text: token.Literal(tok, p.srcRunes)})
	}
	return &CommentGroup{List: list}
}

func (p *Parser) parseBinding(name token.Token) Decl {
	directive := p.parseImportDirective()
	semicolon := p.expect(token.Semicolon)

	var path *StringLit
	if len(directive.Args) > 0 {
		if strLit, ok := directive.Args[0].(*StringLit); ok {
			path = strLit
		}
	}

	return &BindingDecl{
		Name:   &Ident{NamePos: token.Pos(name.Start), Name: token.Literal(name, p.srcRunes)},
		Path:   path,
		EndPos: token.Pos(semicolon.End),
	}
}

func (p *Parser) parseRule(name token.Token) Decl {
	var body []Expr
	for p.peek().Kind != token.Semicolon && p.peek().Kind != token.EOF {
		body = append(body, p.parseProduction())
	}

	if len(body) == 0 {
		// Report error at the position where the body was expected (after the '=')
		p.errorf(token.Pos(name.Start), "rule declaration must have a body")
	}

	semicolon := p.expect(token.Semicolon)

	return &RuleDecl{
		Name:   &Ident{NamePos: token.Pos(name.Start), Name: token.Literal(name, p.srcRunes)},
		Body:   body,
		EndPos: token.Pos(semicolon.End),
	}
}

func (p *Parser) parseProduction() Expr {
	expr := p.parseTerm()
	if p.peek().Kind == token.Pipe {
		alts := []Expr{expr}
		for p.peek().Kind == token.Pipe {
			p.next()
			alts = append(alts, p.parseTerm())
		}
		return &AlternativeExpr{Exprs: alts}
	}
	return expr
}

func (p *Parser) parseTerm() Expr {
	terms := []Expr{p.parseBasic()}
	for p.peek().Kind != token.Pipe && p.peek().Kind != token.Semicolon && p.peek().Kind != token.EOF && p.peek().Kind != token.RParen && p.peek().Kind != token.RBrack && p.peek().Kind != token.RBrace {
		terms = append(terms, p.parseBasic())
	}

	if len(terms) == 1 {
		return terms[0]
	}
	return &SequenceExpr{Exprs: terms}
}

func (p *Parser) parseBasic() Expr {
	switch p.peek().Kind {
	case token.Ident:
		return p.parseNonTerminal()
	case token.String, token.Regex:
		return p.parseTerminal()
	case token.AtDirective:
		return p.parseImportDirective()
	case token.LBrack:
		return p.parseOptional()
	case token.LBrace:
		return p.parseRepetition()
	case token.LParen:
		return p.parseGroup()
	default:
		tok := p.next()
		p.errorf(token.Pos(tok.Start), "unexpected token %s", tok.Kind)
		return &Ident{}
	}
}

func (p *Parser) parseImportDirective() *DirectiveExpr {
	at := p.expect(token.AtDirective)
	p.expect(token.LParen)
	arg := p.expect(token.String)
	rparen := p.expect(token.RParen)

	return &DirectiveExpr{
		AtPos:  token.Pos(at.Start),
		Name:   &Ident{NamePos: token.Pos(at.Start) + 1, Name: "import"},
		Args:   []Expr{&StringLit{ValuePos: token.Pos(arg.Start), Value: token.Literal(arg, p.srcRunes)}},
		EndPos: token.Pos(rparen.End),
	}
}

func (p *Parser) parseTerminal() Expr {
	tok := p.next()
	if tok.State == token.Invalid {
		switch tok.Kind {
		case token.String:
			p.errorf(token.Pos(tok.Start), "invalid string literal")
		case token.Regex:
			p.errorf(token.Pos(tok.Start), "invalid regex literal")
		}
	}
	if tok.Kind == token.String {
		return &StringLit{ValuePos: token.Pos(tok.Start), Value: token.Literal(tok, p.srcRunes)}
	}
	return &RegexLit{ValuePos: token.Pos(tok.Start), Value: token.Literal(tok, p.srcRunes)}
}

func (p *Parser) parseNonTerminal() Expr {
	identToken := p.expect(token.Ident)
	var expr Expr = &Ident{NamePos: token.Pos(identToken.Start), Name: token.Literal(identToken, p.srcRunes)}

	for p.peek().Kind == token.Dot {
		p.next()
		selToken := p.expect(token.Ident)
		expr = &MemberExpr{
			Object: expr,
			Member: &Ident{NamePos: token.Pos(selToken.Start), Name: token.Literal(selToken, p.srcRunes)},
		}
	}

	return expr
}

func (p *Parser) parseOptional() Expr {
	lbrack := p.expect(token.LBrack)
	expr := p.parseProduction()
	rbrack := p.expect(token.RBrack)
	return &OptionalExpr{Lbrack: token.Pos(lbrack.Start), Expr: expr, Rbrack: token.Pos(rbrack.Start)}
}

func (p *Parser) parseRepetition() Expr {
	lbrace := p.expect(token.LBrace)
	expr := p.parseProduction()
	rbrace := p.expect(token.RBrace)
	return &RepetitionExpr{Lbrace: token.Pos(lbrace.Start), Expr: expr, Rbrace: token.Pos(rbrace.Start)}
}

func (p *Parser) parseGroup() Expr {
	lparen := p.expect(token.LParen)
	expr := p.parseProduction()
	rparen := p.expect(token.RParen)
	return &GroupExpr{Lparen: token.Pos(lparen.Start), Expr: expr, Rparen: token.Pos(rparen.Start)}
}

func (p *Parser) peek() token.Token {
	newlineCount := 0
	for p.pos < len(p.tokens) && p.tokens[p.pos].Kind == token.Newline {
		p.pos++
		newlineCount++
	}
	if newlineCount > 1 {
		p.hadBlankLine = true
	}

	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos]
}

func (p *Parser) next() token.Token {
	p.peek() // Skips newlines
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) expect(kind token.Kind) token.Token {
	tok := p.next()
	if tok.Kind != kind {
		p.errorf(token.Pos(tok.Start), "expected %s, got %s", kind, tok.Kind)
	}
	if tok.State == token.Invalid {
		// Report errors for invalid tokens consumed via expect()
		switch tok.Kind {
		case token.String:
			p.errorf(token.Pos(tok.Start), "invalid string literal")
		case token.Regex:
			p.errorf(token.Pos(tok.Start), "unterminated regex literal")
		}
	}
	return tok
}

func (p *Parser) errorf(pos token.Pos, format string, args ...interface{}) {
	p.errors.Add(pos, fmt.Sprintf(format, args...))
}
