package check

import (
	"fmt"
	"grammar/ast"
	"grammar/token"
	"io/ioutil"
	"path/filepath"
)

// CompilationUnit manages a cache of loaded namespaces and handles file loading.
type CompilationUnit struct {
	Namespaces map[string]*Namespace // Cache of loaded namespaces, mapping file path to Namespace.
	Errors     ast.ErrorList
	loading    map[string]bool // Used to detect import cycles.
}

// NewCompilationUnit creates a new compilation unit.
func NewCompilationUnit() *CompilationUnit {
	return &CompilationUnit{
		Namespaces: make(map[string]*Namespace),
		loading:    make(map[string]bool),
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

// LoadFile loads a grammar file from disk.
func (cu *CompilationUnit) LoadFile(path string) (*Namespace, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path for %s: %w", path, err)
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", absPath, err)
	}

	return cu.LoadSource(content, absPath)
}

// LoadSource parses grammar source content and returns its namespace.
// It handles caching based on the provided path.
func (cu *CompilationUnit) LoadSource(content []byte, path string) (*Namespace, error) {
	if cu.loading[path] {
		cu.AddError(token.NoPos, fmt.Sprintf("import cycle detected involving %s", path))
		// Return the partially loaded namespace to avoid infinite recursion
		// and allow other errors to be found.
		return cu.Namespaces[path], nil
	}
	cu.loading[path] = true
	defer func() { cu.loading[path] = false }()

	// 1. Check cache
	if ns, found := cu.Namespaces[path]; found {
		return ns, nil
	}

	// 2. Parse the source
	srcRunes := []rune(string(content))
	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()
	parser := ast.NewParser(tokens, srcRunes)
	file, err := parser.ParseFile()
	if err != nil {
		if errs, ok := err.(ast.ErrorList); ok {
			cu.Errors = append(cu.Errors, errs...)
		}
		return nil, err
	}

	// 3. Create and cache namespace early for cycle detection
	ns := NewNamespace(path)
	cu.Namespaces[path] = ns

	// 4. Populate namespace by processing declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.BindingDecl:
			if d.Path == nil {
				cu.AddError(d.Name.Pos(), "missing import path")
				continue
			}
			importPathLiteral := d.Path.Value
			// remove quotes
			importPath := importPathLiteral[1 : len(importPathLiteral)-1]
			
			// Resolve path relative to the current file's directory
			importDir := filepath.Dir(path)
			importedFilePath := filepath.Join(importDir, importPath)

			// Recursively load imported namespace
			importedNs, _ := cu.LoadFile(importedFilePath) // Use LoadFile for dependent files
			if importedNs == nil {
				cu.AddError(d.Path.Pos(), fmt.Sprintf("could not load imported namespace '%s'", importPath))
				continue
			}
			ns.Members[d.Name.Name] = d
			ns.Types[d.Name.Name] = &NamespaceType{Name: importedNs.Name}

		case *ast.RuleDecl:
			ns.Members[d.Name.Name] = d
			ns.Types[d.Name.Name] = Production
		}
	}

	return ns, nil
}
