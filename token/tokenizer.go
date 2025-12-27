package token

import (
	"unicode"
)

// Tokenizer scans an input source and produces a slice of Tokens.
type Tokenizer struct {
	src          []rune
	offset       int  // current offset in src
	readOffset   int  // next char to read
	ch           rune // current char
	SkipComments bool // Add SkipComments field
	SkipNewlines bool // Add SkipNewlines field
}

// NewTokenizer creates a new Tokenizer for the given input source (rune slice).
func NewTokenizer(src []rune, skipComments bool, skipNewlines bool) *Tokenizer {
	t := &Tokenizer{src: src, SkipComments: skipComments, SkipNewlines: skipNewlines}
	t.nextChar()
	return t
}

// Scan scans the input source and returns a slice of Tokens.
func (t *Tokenizer) Scan() []Token {
	var tokens []Token
loop:
	for t.ch != 0 {
		t.skipWhitespace()
		start := t.offset

		var token Token
		switch {
		case t.ch == 0:
			break loop
		case t.ch == '\n':
			if t.SkipNewlines {
				t.nextChar()
				continue
			}
			token = t.newToken(Newline, start, start+1)
			t.nextChar()
		case isLetter(t.ch):
			token = t.scanIdent()
		case isDigit(t.ch):
			token = t.scanNumber()
		case t.ch == '"':
			token = t.scanString()
		case t.ch == '@':
			token = t.scanDirective()
		case t.ch == '/':
			// Check if it's a regex or a comment
			if t.peek() == '/' { // Comment
				if t.SkipComments {
					t.skipComment()
					continue
				}
				token = t.scanComment()
			} else {
				token = t.scanRegex()
			}
		default:
			switch t.ch {
			case ';':
				token = t.newToken(Semicolon, start, t.offset+1)
				t.nextChar()
			case ':':
				token = t.newToken(Colon, start, t.offset+1)
				t.nextChar()
			case '(':
				token = t.newToken(LParen, start, t.offset+1)
				t.nextChar()
			case ')':
				token = t.newToken(RParen, start, t.offset+1)
				t.nextChar()
			case '{':
				token = t.newToken(LBrace, start, t.offset+1)
				t.nextChar()
			case '}':
				token = t.newToken(RBrace, start, t.offset+1)
				t.nextChar()
			case '[':
				token = t.newToken(LBrack, start, t.offset+1)
				t.nextChar()
			case ']':
				token = t.newToken(RBrack, start, t.offset+1)
				t.nextChar()
			case '|':
				token = t.newToken(Pipe, start, t.offset+1)
				t.nextChar()
			case '.':
				token = t.newToken(Dot, start, t.offset+1)
				t.nextChar()
			case '=': // Handle assignment token
				token = t.newToken(Assign, start, t.offset+1)
				t.nextChar()
			default:
				token = t.newToken(Illegal, start, t.offset+1)
				t.nextChar()
			}
		}

		tokens = append(tokens, token)
	}
	tokens = append(tokens, t.newToken(EOF, t.offset, t.offset))
	return tokens
}

func (t *Tokenizer) newToken(kind Kind, start, end int) Token {
	return Token{Kind: kind, State: Valid, Start: start, End: end}
}

func (t *Tokenizer) nextChar() {
	if t.readOffset < len(t.src) {
		t.ch = t.src[t.readOffset]
	} else {
		t.ch = 0 // EOF
	}
	t.offset = t.readOffset
	t.readOffset++
}

func (t *Tokenizer) peek() rune {
	if t.readOffset < len(t.src) {
		return t.src[t.readOffset]
	}
	return 0
}

func (t *Tokenizer) skipWhitespace() {
	for t.ch == ' ' || t.ch == '\t' || t.ch == '\r' {
		t.nextChar()
	}
}

func (t *Tokenizer) skipComment() {
	for t.ch != '\n' && t.ch != 0 {
		t.nextChar()
	}
	if t.ch == '\n' {
		t.nextChar() // Consume the newline
	}
}

func (t *Tokenizer) scanComment() Token {
	start := t.offset
	t.skipComment()
	return t.newToken(Comment, start, t.offset)
}

func (t *Tokenizer) scanIdent() Token {
	start := t.offset
	// Identifiers can contain letters, digits, and underscores
	for isLetter(t.ch) || isDigit(t.ch) {
		t.nextChar()
	}
	return t.newToken(Ident, start, t.offset)
}

func (t *Tokenizer) scanDirective() Token {
	start := t.offset
	t.nextChar() // Consume '@'

	for isLetter(t.ch) || isDigit(t.ch) {
		t.nextChar()
	}

	return t.newToken(AtDirective, start, t.offset)
}

func (t *Tokenizer) scanNumber() Token {
	start := t.offset
	// For simplicity, just consume digits for now.
	// A more robust implementation would handle decimals, signs, etc.
	for isDigit(t.ch) {
		t.nextChar()
	}
	return t.newToken(Number, start, t.offset)
}

func (t *Tokenizer) scanString() Token {
	start := t.offset
	t.nextChar() // Consume the opening '"'
	for t.ch != '"' && t.ch != 0 {
		t.nextChar()
	}
	if t.ch == 0 { // Unterminated string
		return Token{Kind: String, State: Invalid, Start: start, End: t.offset}
	}
	t.nextChar() // Consume the closing '"'
	return t.newToken(String, start, t.offset)
}

func (t *Tokenizer) scanRegex() Token {
	start := t.offset
	t.nextChar() // Consume the opening '/'

	for {
		if t.ch == '/' {
			break
		}
		switch t.ch {
		case '\\':
			t.nextChar() // Consume the backslash
			switch t.ch {
			case '/', 'n', 't', 'r', 'd', 's', 'c', '\\', '.', '(', ')', '{', '}', '[', ']':
				t.nextChar() // Consume the valid escaped character
			default:
				// Invalid escape sequence
				// Consume until the closing slash or EOF to create the token
				for t.ch != '/' && t.ch != 0 {
					t.nextChar()
				}
				if t.ch == '/' {
					t.nextChar()
				}
				return Token{Kind: Regex, State: Invalid, Start: start, End: t.offset}
			}
		case 0:
			// Unterminated regex
			return Token{Kind: Regex, State: Invalid, Start: start, End: t.offset}
		default:
			t.nextChar()
		}
	}
	t.nextChar() // Consume the closing '/'
	return t.newToken(Regex, start, t.offset)
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
