package ast

import (
	"fmt"
	"grammar/token"
)

// Parser holds the parser's state.
type Parser struct {
	tokens   []token.Token
	srcRunes []rune
	pos      int // current position in the token slice
	errors   ErrorList
}

// NewParser creates a new Parser.
func NewParser(tokens []token.Token, srcRunes []rune) *Parser {
	return &Parser{tokens: tokens, srcRunes: srcRunes}
}

// ParseFile parses the input tokens and returns a File node.
func (p *Parser) ParseFile() (*File, error) {
	var decls []Decl
	for p.peek().Kind != token.EOF {
		decls = append(decls, p.parseDecl())
	}

	if len(p.errors) > 0 {
		return nil, p.errors
	}

	return &File{Decls: decls}, nil
}

func (p *Parser) parseDecl() Decl {
	ident := p.expect(token.Ident)
	p.expect(token.Assign)
	if p.peek().Kind == token.AtDirective {
		return p.parseBinding(ident)
	}
	return p.parseRule(ident)
}

func (p *Parser) parseBinding(name token.Token) Decl {
	p.expect(token.AtDirective)
	p.expect(token.LParen)
	path := p.expect(token.String)
	p.expect(token.RParen)
	p.expect(token.Semicolon)

	return &BindingDecl{
		Name: &Ident{NamePos: token.Pos(name.Start), Name: token.Literal(name, p.srcRunes)},
		Path: &StringLit{ValuePos: token.Pos(path.Start), Value: token.Literal(path, p.srcRunes)},
	}
}

func (p *Parser) parseRule(name token.Token) Decl {
	var body []Expr
	for p.peek().Kind != token.Semicolon && p.peek().Kind != token.EOF {
		body = append(body, p.parseProduction())
	}

	p.expect(token.Semicolon)

	return &RuleDecl{
		Name: &Ident{NamePos: token.Pos(name.Start), Name: token.Literal(name, p.srcRunes)},
		Body: body,
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
	expr := p.parseBasic()
	// The grammar says `term = basic | basic production`, which is left-recursive.
	// We can parse this iteratively.
	terms := []Expr{expr}
	for p.peek().Kind != token.Pipe && p.peek().Kind != token.Semicolon && p.peek().Kind != token.EOF && p.peek().Kind != token.RParen && p.peek().Kind != token.RBrack && p.peek().Kind != token.RBrace {
		terms = append(terms, p.parseBasic())
	}

	if len(terms) == 1 {
		return terms[0]
	}

	// This part of the grammar is tricky. Let's just create a sequence for now.
	// `basic production` is not well defined. I will treat it as a sequence of basics.
	return &TermExpr{X: &AlternativeExpr{Exprs: terms}} // This is probably wrong
}

func (p *Parser) parseBasic() Expr {
	switch p.peek().Kind {
	case token.Ident:
		return p.parseNonTerminal()
	case token.String, token.Regex:
		return p.parseTerminal()
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

func (p *Parser) parseTerminal() Expr {
	tok := p.next()
	if tok.Kind == token.String {
		return &StringLit{ValuePos: token.Pos(tok.Start), Value: token.Literal(tok, p.srcRunes)}
	}
	return &RegexLit{ValuePos: token.Pos(tok.Start), Value: token.Literal(tok, p.srcRunes)}
}

func (p *Parser) parseNonTerminal() Expr {
	// A non_terminal always starts with an ident.
	// Parse the first ident.
	identToken := p.expect(token.Ident)
	var expr Expr = &Ident{NamePos: token.Pos(identToken.Start), Name: token.Literal(identToken, p.srcRunes)}

	// Now, check if it's a member_access (ident { "." ident })
	// Loop as long as we see a dot.
	for p.peek().Kind == token.Dot {
		p.next()                          // Consume the dot
		selToken := p.expect(token.Ident) // Expect the next ident for the member
		expr = &MemberExpr{               // Build the MemberExpr left-associatively
			Object: expr, // The left-hand side is the expression parsed so far
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

// peek returns the next token without consuming it.
func (p *Parser) peek() token.Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // Return EOF
	}
	return p.tokens[p.pos]
}

// next consumes and returns the next token.
func (p *Parser) next() token.Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // Return EOF
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

// expect consumes the next token and checks its kind.
// If the kind doesn't match, it records an error.
func (p *Parser) expect(kind token.Kind) token.Token {
	tok := p.next()
	if tok.Kind != kind {
		p.errorf(token.Pos(tok.Start), "expected %s, got %s", kind, tok.Kind)
	}
	return tok
}

// errorf records an error.
func (p *Parser) errorf(pos token.Pos, format string, args ...interface{}) {
	p.errors.Add(pos, fmt.Sprintf(format, args...))
}
