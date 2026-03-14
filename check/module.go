package check

import "grammar/ast"

// Module represents a module (a single .grammar file).
// It maps identifiers to their declarations and types.
type Module struct {
	Name        string              // Module name, typically the file path.
	File        *ast.File           // The parsed AST of the file.
	Members     map[string]ast.Decl // Map of member names to their declarations.
	Types       map[string]Type     // Map of member names to their types.
	Package     *Package            // The package this module belongs to (nil if not part of a package)
	PackageName string              // The declared package name (from @package directive)
}

// Package represents a directory containing .grammar files.
type Package struct {
	Name    string             // Package name (from @package directive or directory name)
	Path    string             // Package directory path
	Modules map[string]*Module // Modules in the package, keyed by module name (filename without .grammar)
}

// NewModule creates a new, empty module.
func NewModule(name string) *Module {
	return &Module{
		Name:    name,
		Members: make(map[string]ast.Decl),
		Types:   make(map[string]Type),
	}
}

// NewPackage creates a new, empty package.
func NewPackage(name, path string) *Package {
	return &Package{
		Name:    name,
		Path:    path,
		Modules: make(map[string]*Module),
	}
}
