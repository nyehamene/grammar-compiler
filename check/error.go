package check

import (
	"fmt"
	"grammar/token"
	"io"
	"sort"
)

// Error represents a single semantic error.
type Error struct {
	Path    string
	Pos     token.Pos
	Message string
	Warning bool
}

func (e Error) Error() string {
	// The format method on ErrorList should be used for user-facing output.
	return e.Message
}

// ErrorList is a slice of errors, sorted by position.
type ErrorList []Error

// add adds a new error to the list.
func (p *ErrorList) add(path string, pos token.Pos, msg string, isWarning bool) {
	*p = append(*p, Error{Path: path, Pos: pos, Message: msg, Warning: isWarning})
}

// Add adds a new error to the list.
func (p *ErrorList) Add(path string, pos token.Pos, msg string) {
	p.add(path, pos, msg, false)
}

// AddWarning adds a new warning to the list.
func (p *ErrorList) AddWarning(path string, pos token.Pos, msg string) {
	p.add(path, pos, msg, true)
}

// Len is the number of elements in the collection.
func (p ErrorList) Len() int { return len(p) }

// Swap swaps the elements with indexes i and j.
func (p ErrorList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Less reports whether the element with index i
// must be ordered before the element with index j.
func (p ErrorList) Less(i, j int) bool {
	if p[i].Path != p[j].Path {
		return p[i].Path < p[j].Path
	}
	return p[i].Pos < p[j].Pos
}

// Error returns the first error message.
func (p ErrorList) Error() string {
	if len(p) == 0 {
		return ""
	}
	sort.Sort(p)
	return p[0].Error()
}

// Format prints the error list to the writer in a standard format.
func (p ErrorList) Format(w io.Writer, sources map[string][]rune) {
	sort.Sort(p)
	for _, e := range p {
		src, ok := sources[e.Path]
		if !ok {
			fmt.Fprintf(w, "%s: %s\n", e.Path, e.Message)
			continue
		}
		line, col := token.FindLineAndCol(e.Pos, src)
		fmt.Fprintf(w, "%s:%d:%d: %s\n", e.Path, line, col, e.Message)
	}
}
