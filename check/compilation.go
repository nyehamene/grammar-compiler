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
	Errors     ErrorList
	Sources    map[string][]rune // Cache of file contents.
	loading    map[string]bool   // Used to detect import cycles.
}

// NewCompilationUnit creates a new compilation unit.
func NewCompilationUnit() *CompilationUnit {
	return &CompilationUnit{
		Namespaces: make(map[string]*Namespace),
		Sources:    make(map[string][]rune),
		loading:    make(map[string]bool),
	}
}

// AddError adds a new error to the compilation unit.
func (cu *CompilationUnit) AddError(path string, pos token.Pos, message string) {
	cu.Errors.Add(path, pos, message)
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
func (cu *CompilationUnit) LoadSource(content []byte, path string) (*Namespace, error) {
	if cu.loading[path] {
		cu.AddError(path, token.NoPos, fmt.Sprintf("import cycle detected involving %s", path))
		return cu.Namespaces[path], nil
	}
	cu.loading[path] = true
	defer func() { cu.loading[path] = false }()

	if ns, found := cu.Namespaces[path]; found {
		return ns, nil
	}

	srcRunes := []rune(string(content))
	cu.Sources[path] = srcRunes // Store source content

	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()
	parser := ast.NewParser(tokens, srcRunes)
	file, err := parser.ParseFile()
	if err != nil {
		if errs, ok := err.(ast.ErrorList); ok {
			for _, e := range errs {
				cu.AddError(path, e.Pos, e.Message)
			}
		}
		return nil, err
	}

	ns := NewNamespace(path)
	ns.File = file
	cu.Namespaces[path] = ns

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.BindingDecl:
			if _, found := ns.Members[d.Name.Name]; found {
				cu.AddError(path, d.Pos(), fmt.Sprintf("identifier '%s' redeclared in this namespace", d.Name.Name))
				continue
			}
			if d.Path == nil {
				cu.AddError(path, d.Pos(), "missing import path")
				continue
			}
			importPathLiteral := d.Path.Value
			importPath := importPathLiteral[1 : len(importPathLiteral)-1]
			importDir := filepath.Dir(path)
			importedFilePath := filepath.Join(importDir, importPath)

			importedNs, err := cu.LoadFile(importedFilePath)
			if err != nil {
				cu.AddError(path, d.Path.Pos(), fmt.Sprintf("could not load imported namespace '%s'", importPath))
				continue
			}
			ns.Members[d.Name.Name] = d
			ns.Types[d.Name.Name] = &NamespaceType{Name: importedNs.Name}

		case *ast.RuleDecl:
			if _, found := ns.Members[d.Name.Name]; found {
				cu.AddError(path, d.Pos(), fmt.Sprintf("identifier '%s' redeclared in this namespace", d.Name.Name))
				continue
			}
			ns.Members[d.Name.Name] = d
			ns.Types[d.Name.Name] = Production

		case *ast.CommentGroup:
			// No name to check for comments.
		}
	}

	return ns, nil
}