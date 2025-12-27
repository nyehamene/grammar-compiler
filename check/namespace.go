package check

import "grammar/ast"

// Namespace represents the semantic elements of a grammar file.
// It maps identifiers to their declarations and types.
type Namespace struct {
	Name    string               // Namespace name, typically the file path.
	Members map[string]ast.Decl // Map of member names to their declarations.
	Types   map[string]Type      // Map of member names to their types.
}

// NewNamespace creates a new, empty namespace.
func NewNamespace(name string) *Namespace {
	return &Namespace{
		Name:    name,
		Members: make(map[string]ast.Decl),
		Types:   make(map[string]Type),
	}
}
