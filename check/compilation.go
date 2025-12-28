package check

import (
	"fmt"
	"grammar/ast"
	"grammar/token"
	"log"
	"net/url"
	"path/filepath"
)

// CompilationUnit manages a cache of loaded namespaces and handles file loading.
type CompilationUnit struct {
	loader     FileLoader
	Namespaces map[string]*Namespace
	Errors     map[string]ErrorList
	Sources    map[string][]rune
	loading    map[string]bool
	log        *log.Logger
}

// NewCompilationUnit creates a new compilation unit with a file loader.
func NewCompilationUnit(loader FileLoader, logger *log.Logger) *CompilationUnit {
	l := log.New(logger.Writer(), "(cu)", log.Flags())
	return &CompilationUnit{
		loader:     loader,
		log:        l,
		Namespaces: make(map[string]*Namespace),
		Sources:    make(map[string][]rune),
		loading:    make(map[string]bool),
		Errors:     map[string]ErrorList{},
	}
}

// AddError adds a new error to the compilation unit.
func (cu *CompilationUnit) AddError(path string, pos token.Pos, message string) {
	errlist, ok := cu.Errors[path]
	if !ok {
		errlist = make(ErrorList, 0, 10)
	}
	errlist.Add(path, pos, message)
	cu.Errors[path] = errlist
}

// Err returns the collected errors, or nil if there are none.
func (cu *CompilationUnit) Err(path string) error {
	if len(cu.Errors) == 0 {
		return nil
	}
	return cu.Errors[path]
}

// LoadFile loads a grammar file using the compilation unit's loader.
func (cu *CompilationUnit) LoadFile(path string) (*Namespace, error) {
	content, err := cu.loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	return cu.LoadSource(content, path)
}

func (cu *CompilationUnit) RemoveNamespace(path string) {
	delete(cu.Namespaces, path)
	delete(cu.Errors, path)
}

func resolveImport(base, imp string) (string, error) {
	baseURI, err := url.Parse(base)
	// If it's not a valid URL or has no scheme, treat as a file path.
	if err != nil || baseURI.Scheme == "" {
		return filepath.Join(filepath.Dir(base), imp), nil
	}

	// It's a URL, resolve reference.
	impURI, err := url.Parse(imp)
	if err != nil {
		return "", fmt.Errorf("invalid import path: %w", err)
	}
	return baseURI.ResolveReference(impURI).String(), nil
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
	cu.Sources[path] = srcRunes

	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()
	parser := ast.NewParser(tokens, srcRunes)
	file, parseErr := parser.ParseFile()
	if parseErr != nil {
		if errs, ok := parseErr.(ast.ErrorList); ok {
			for _, e := range errs {
				cu.AddError(path, e.Pos, e.Message)
			}
		}
		// Continue with partial file for completion/hover
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

			importedPath, resolveErr := resolveImport(path, importPath)
			if resolveErr != nil {
				cu.AddError(path, d.Path.Pos(), resolveErr.Error())
				continue
			}

			importedNs, loadErr := cu.LoadFile(importedPath)
			if loadErr != nil {
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

	return ns, parseErr
}
