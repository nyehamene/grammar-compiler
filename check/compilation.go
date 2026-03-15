package check

import (
	"fmt"
	"grammar/ast"
	"grammar/log"
	"grammar/token"
	"net/url"
	"path/filepath"
)

// CompilationUnit manages a cache of loaded modules/packages and handles file loading.
type CompilationUnit struct {
	loader     FileLoader
	Modules    map[string]*Module  // Loaded modules (files)
	Namespaces map[string]*Module  // Deprecated: kept for backward compatibility
	Packages   map[string]*Package // Loaded packages (directories)
	Errors     map[string]ErrorList
	Sources    map[string][]rune
	loading    map[string]bool
	log        log.StructuredLogger
}

// NewCompilationUnit creates a new compilation unit with a file loader.
func NewCompilationUnit(loader FileLoader, logger log.StructuredLogger) *CompilationUnit {
	return &CompilationUnit{
		loader:     loader,
		log:        logger,
		Modules:    make(map[string]*Module),
		Namespaces: make(map[string]*Module),
		Packages:   make(map[string]*Package),
		Sources:    make(map[string][]rune),
		loading:    make(map[string]bool),
		Errors:     map[string]ErrorList{},
	}
}

// AddError adds a new error to the compilation unit.
func (cu *CompilationUnit) AddError(path string, line, col int, message string) {
	errlist, ok := cu.Errors[path]
	if !ok {
		errlist = make(ErrorList, 0, 10)
	}
	errlist.add(path, line, col, message, false)
	cu.Errors[path] = errlist
}

// AddWarning adds a new warning to the compilation unit.
func (cu *CompilationUnit) AddWarning(path string, line, col int, message string) {
	errlist, ok := cu.Errors[path]
	if !ok {
		errlist = make(ErrorList, 0, 10)
	}
	errlist.add(path, line, col, message, true)
	cu.Errors[path] = errlist
}

// Err returns the collected errors (excluding warnings), or nil if there are none.
func (cu *CompilationUnit) Err(path string) error {
	if len(cu.Errors) == 0 {
		return nil
	}

	// Filter out warnings
	errlist := cu.Errors[path]
	var errors ErrorList
	for _, err := range errlist {
		if !err.Warning {
			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

// LoadFile loads a grammar file using the compilation unit's loader.
func (cu *CompilationUnit) LoadFile(path string) (*Module, error) {
	content, err := cu.loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	return cu.LoadSource(content, path), nil
}

// LoadPackage loads all .grammar files in a directory as a package.
func (cu *CompilationUnit) LoadPackage(dirPath string) (*Package, error) {
	// Normalize and check if it's a directory
	normalizedPath, err := cu.loader.NormalizePath(dirPath)
	if err != nil {
		return nil, fmt.Errorf("invalid package path: %w", err)
	}

	isDir, err := cu.loader.IsDir(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("could not check path: %w", err)
	}
	if !isDir {
		return nil, fmt.Errorf("package path is not a directory: %s", normalizedPath)
	}

	// Check if already loaded
	if pkg, found := cu.Packages[normalizedPath]; found {
		return pkg, nil
	}

	// Load all .grammar files in the directory
	files, err := cu.loader.LoadDir(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("package directory has no .grammar files: %s", normalizedPath)
	}

	// Collect package names from all modules first
	var packageNames []string
	var moduleNameToPath = make(map[string]string)

	for _, filePath := range files {
		// Skip files that are already being loaded (to avoid cycles)
		if cu.loading[filePath] {
			continue
		}

		moduleName := filepath.Base(filePath[:len(filePath)-len(".grammar")])
		moduleNameToPath[moduleName] = filePath

		// Load the module to collect @package declarations
		// Note: This might cause recursion if other files in the package also load this package
		mod, loadErr := cu.LoadFile(filePath)
		if loadErr != nil {
			// Ignore errors here - the file might have other issues
			// We'll report those when checking the file directly
			continue
		}

		// Collect package name from @package directive
		for _, decl := range mod.File.Decls {
			if dir, ok := decl.(*ast.DirectiveExpr); ok {
				if dir.Name != nil && dir.Name.Name == "package" {
					if len(dir.Args) > 0 {
						if strLit, ok := dir.Args[0].(*ast.StringLit); ok {
							packageName := strLit.Value[1 : len(strLit.Value)-1] // remove quotes
							packageNames = append(packageNames, packageName)
						}
					}
				}
			}
		}
	}

	// Determine package name
	var packageName string
	if len(packageNames) > 0 {
		// All names must be consistent
		packageName = packageNames[0]
		for _, name := range packageNames {
			if name != packageName {
				// Package name mismatch - report on all files
				for _, filePath := range files {
					cu.AddError(filePath, 0, 0, fmt.Sprintf("package name mismatch: '%s' vs '%s'", packageName, name))
				}
				packageName = packageNames[0] // Use first for now
			}
		}
	} else {
		// Infer from directory name
		packageName = filepath.Base(normalizedPath)
	}

	// Create the package
	pkg := NewPackage(packageName, normalizedPath)
	cu.Packages[normalizedPath] = pkg

	// Populate package modules
	for moduleName, filePath := range moduleNameToPath {
		mod := cu.Modules[filePath]
		if mod != nil {
			mod.Package = pkg
			mod.PackageName = packageName
			pkg.Modules[moduleName] = mod
		}
	}

	return pkg, nil
}

// GetPackageForFile returns the package that contains the given file path.
// If the file is not part of any package (e.g., a single file outside any directory with other .grammar files),
// it returns nil.
func (cu *CompilationUnit) GetPackageForFile(filePath string) *Package {
	dir := filepath.Dir(filePath)
	for {
		if pkg, found := cu.Packages[dir]; found {
			return pkg
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil
}

func (cu *CompilationUnit) RemoveModule(path string) {
	delete(cu.Modules, path)
	delete(cu.Errors, path)
}

// RemoveNamespace is deprecated, use RemoveModule instead
func (cu *CompilationUnit) RemoveNamespace(path string) {
	cu.RemoveModule(path)
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

// LoadSource parses grammar source content and returns its module.
func (cu *CompilationUnit) LoadSource(content []byte, path string) *Module {
	srcRunes := []rune(string(content))
	cu.Sources[path] = srcRunes

	if cu.loading[path] {
		line, col := token.FindLineAndCol(token.NoPos, srcRunes)
		cu.AddError(path, line, col, fmt.Sprintf("import cycle detected involving %s", path))
		return cu.Modules[path]
	}
	cu.loading[path] = true
	defer func() { cu.loading[path] = false }()

	if mod, found := cu.Modules[path]; found {
		return mod
	}

	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()
	parser := ast.NewParser(tokens, srcRunes)
	file, parseErr := parser.ParseFile()
	if parseErr != nil {
		if errs, ok := parseErr.(ast.ErrorList); ok {
			for _, e := range errs {
				line, col := token.FindLineAndCol(e.Pos, srcRunes)
				cu.AddError(path, line, col, e.Message)
			}
		} else {
			cu.log.Info("CheckSource: unexpected parser error type", log.Fields{
				"type":  fmt.Sprintf("%T", parseErr),
				"error": parseErr,
			})
		}
	}

	mod := NewModule(path)
	mod.File = file
	cu.Modules[path] = mod

	// Also populate Namespaces for backward compatibility
	cu.Namespaces[path] = &Module{
		Name:    path,
		File:    file,
		Members: mod.Members,
		Types:   mod.Types,
	}

	// Note: We don't automatically load packages here to avoid import cycles.
	// Packages are loaded when explicitly imported via @import directive.
	// The @package directive only declares the package name and returns a package reference.

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.BindingDecl:
			if _, found := mod.Members[d.Name.Name]; found {
				line, col := token.FindLineAndCol(d.Pos(), srcRunes)
				cu.AddError(path, line, col, fmt.Sprintf("identifier '%s' redeclared in this namespace", d.Name.Name))
				continue
			}
			if d.Path == nil {
				line, col := token.FindLineAndCol(d.Pos(), srcRunes)
				cu.AddError(path, line, col, "missing import path")
				continue
			}
			importPathLiteral := d.Path.Value
			importPath := importPathLiteral[1 : len(importPathLiteral)-1]

			importedPath, resolveErr := resolveImport(path, importPath)
			if resolveErr != nil {
				line, col := token.FindLineAndCol(d.Path.Pos(), srcRunes)
				cu.AddError(path, line, col, resolveErr.Error())
				continue
			}

			// Handle @package directive specially - it uses the current file's directory as the package
			if d.Kind == ast.ImportPackage {
				// For @package, use the current file's directory as the package
				dirPath := filepath.Dir(path)
				pkg, pkgErr := cu.LoadPackage(dirPath)
				if pkgErr != nil {
					line, col := token.FindLineAndCol(d.Path.Pos(), srcRunes)
					cu.AddError(path, line, col, fmt.Sprintf("could not load package: %s", pkgErr))
					continue
				}
				// Verify the package name matches
				// This check is incorrect: importPath is the directory name, pkg.Name is the declared package name.
				// If a user imports "@import("some_dir")" but the package inside is named "@package("actual_package_name")",
				// they expect to use the "actual_package_name" not "some_dir".
				// The actual name mismatch should be handled by LoadPackage when it collects names from multiple modules.
				// Removed: if pkg.Name != importPath { ... }
				mod.Members[d.Name.Name] = d
				mod.Types[d.Name.Name] = &PackageType{Name: pkg.Name, Path: pkg.Path}
				continue
			}

			// Check if importing a directory (package) or a file
			isDir, err := cu.loader.IsDir(importedPath)
			if err == nil && isDir {
				// Package-based import
				d.Kind = ast.ImportPackage
				pkg, pkgErr := cu.LoadPackage(importedPath)
				if pkgErr != nil {
					line, col := token.FindLineAndCol(d.Path.Pos(), srcRunes)
					cu.AddError(path, line, col, fmt.Sprintf("could not load package '%s': %s", importPath, pkgErr))
					continue
				}
				mod.Members[d.Name.Name] = d
				mod.Types[d.Name.Name] = &PackageType{Name: pkg.Name, Path: pkg.Path}
			} else {
				// File-based import (deprecated)
				d.Kind = ast.ImportFile
				importedMod, loadErr := cu.LoadFile(importedPath)
				if loadErr != nil {
					line, col := token.FindLineAndCol(d.Path.Pos(), srcRunes)
					cu.AddError(path, line, col, fmt.Sprintf("could not load imported namespace '%s'", importPath))
					continue
				}
				cu.AddWarning(path, 0, 0, fmt.Sprintf("file-based import '%s' is deprecated, use package import instead", importPath))
				mod.Members[d.Name.Name] = d
				mod.Types[d.Name.Name] = &ModuleType{Name: importedMod.Name}
			}

		case *ast.RuleDecl:
			if _, found := mod.Members[d.Name.Name]; found {
				line, col := token.FindLineAndCol(d.Pos(), srcRunes)
				cu.AddError(path, line, col, fmt.Sprintf("identifier '%s' redeclared in this namespace", d.Name.Name))
				continue
			}
			mod.Members[d.Name.Name] = d
			mod.Types[d.Name.Name] = Production

		case *ast.CommentGroup:
			// No name to check for comments.

		case *ast.DirectiveExpr:
			// Handle @package directive - it's already processed above
			// No member needed for @package as it returns a package reference
		}
	}

	return mod
}
