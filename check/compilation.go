package check

import (
	"grammar/ast"
	"grammar/token"
)

// CompilationUnit manages a cache of loaded namespaces and handles file loading.
type CompilationUnit struct {
	Namespaces map[string]*Namespace // Cache of loaded namespaces, mapping file path to Namespace.
	Errors     ast.ErrorList
}

// NewCompilationUnit creates a new compilation unit.
func NewCompilationUnit() *CompilationUnit {
	return &CompilationUnit{
		Namespaces: make(map[string]*Namespace),
	}
}

// AddError adds a new error to the compilation unit.
func (cu *CompilationUnit) AddError(pos token.Pos, message string) {
	cu.Errors.Add(pos, message)
}

// Err returns the collected errors, or nil if there are none.
func (cu *CompilationUnit) Err() error {
	if len(cu.Errors) == 0 {
		return nil
	}
	return cu.Errors
}
