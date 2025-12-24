package token

import "fmt"

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

	// Punctuation
	Semicolon
	Colon
	LParen
	RParen
	LBrace
	RBrace
	Dot
	At
	Assign // New token kind for '='

	// Keywords
	Import
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
	case Dot:
		return "DOT"
	case At:
		return "AT"
	case Assign:
		return "ASSIGN"
	case Import:
		return "IMPORT"
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
