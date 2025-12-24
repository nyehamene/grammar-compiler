package ast

import (
	"fmt"
	"grammar/token"
)

// Error represents a parser error.
type Error struct {
	Pos     token.Pos
	Message string
}

// ErrorList is a slice of *Error.
type ErrorList []*Error

// Add adds an Error to an ErrorList.
func (p *ErrorList) Add(pos token.Pos, msg string) {
	*p = append(*p, &Error{Pos: pos, Message: msg})
}

// Error returns a string representation of an ErrorList.
func (p ErrorList) Error() string {
	switch len(p) {
	case 0:
		return "no errors"
	case 1:
		return p[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", p[0], len(p)-1)
}

// Error returns a string representation of an Error.
func (e *Error) Error() string {
	return e.Message
}
