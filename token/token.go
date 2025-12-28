package token

import (
	"fmt"
	"strconv"
)

// Pos specifies the zero-based character offset of a token or node in a source file.
type Pos int

const NoPos Pos = 0

// TokenKind is the type of a token.
type Kind int

const (
	// Special tokens
	Illegal Kind = iota
	EOF

	// Literals
	Ident
	String
	Number // New token kind for numbers
	Regex
	Comment
	Newline

	// Punctuation
	Semicolon
	Colon
	LParen
	RParen
	LBrace
	RBrace
	LBrack
	RBrack
	Pipe
	Dot
	Assign // New token kind for '='
	AtDirective
	External
)

// String returns the string representation of a TokenKind.
func (k Kind) String() string {
	switch k {
	case Illegal:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case Ident:
		return "IDENT"
	case String:
		return "STRING"
	case Number:
		return "NUMBER"
	case Regex:
		return "REGEX"
	case Comment:
		return "COMMENT"
	case Newline:
		return "NEWLINE"
	case Semicolon:
		return "SEMICOLON"
	case Colon:
		return "COLON"
	case LParen:
		return "LPAREN"
	case RParen:
		return "RPAREN"
	case LBrace:
		return "LBRACE"
	case RBrace:
		return "RBRACE"
	case LBrack:
		return "LBRACK"
	case RBrack:
		return "RBRACK"
	case Pipe:
		return "PIPE"
	case Dot:
		return "DOT"
	case Assign:
		return "ASSIGN"
	case AtDirective:
		return "AT_Directive"
	case External:
		return "EXTERNAL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", k)
	}
}

// State indicates if a token is valid or invalid.
type State int

const (
	Valid State = iota
	Invalid
)

// Token represents a lexical token.
type Token struct {
	Kind  Kind
	State State
	Start int // Start offset in the input source
	End   int // End offset in the input source
}

// Literal returns the string value of a token.
func Literal(tok Token, src []rune) string {
	if tok.Start < 0 || tok.End > len(src) || tok.Start > tok.End {
		return "" // Invalid token range
	}
	return string(src[tok.Start:tok.End])
}

// FindLineAndCol finds the line and column number for a given offset in the source.
func FindLineAndCol(offset int, srcRunes []rune) (int, int) {
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

// EscapeLexeme escapes a string for display, handling special characters.
func EscapeLexeme(s string) string {
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
