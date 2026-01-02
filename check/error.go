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
	Line    int
	Col     int
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
func (p *ErrorList) add(path string, line, col int, msg string, isWarning bool) {
	*p = append(*p, Error{Path: path, Line: line, Col: col, Message: msg, Warning: isWarning})
}

// Add adds a new error to the list.
func (p *ErrorList) Add(path string, pos token.Pos, msg string) {
	// This will need to be updated by the caller to provide line and col
	// For now, convert pos to line/col
	// Assuming sources map is available (it's not directly in scope here)
	// This will be fixed in the compilation unit where source is available
	*p = append(*p, Error{Path: path, Line: 0, Col: 0, Message: msg, Warning: false}) // Placeholder
}

// AddWarning adds a new warning to the list.
func (p *ErrorList) AddWarning(path string, pos token.Pos, msg string) {
	// This will need to be updated by the caller to provide line and col
	*p = append(*p, Error{Path: path, Line: 0, Col: 0, Message: msg, Warning: true}) // Placeholder
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
	if p[i].Line != p[j].Line {
		return p[i].Line < p[j].Line
	}
	return p[i].Col < p[j].Col
}

// Error returns the first error message.
func (p ErrorList) Error() string {
	if len(p) == 0 {
		return ""
	}
	sort.Sort(p)
	return p[0].Message
}

// Format prints the error list to the writer in a standard format.
func (p ErrorList) Format(w io.Writer, sources map[string][]rune) {
	sort.Sort(p)
	for _, e := range p {
		// Sources map is no longer needed to find line/col, as they are stored in the Error struct
		fmt.Fprintf(w, "%s:%d:%d: %s\n", e.Path, e.Line, e.Col, e.Message)
	}
}
