package check

import (
	"fmt"
	"grammar/ast"
	"grammar/log"
	"grammar/token" // Added import
	"unicode"
)

// Checker holds the state for the type-checking process.
type Checker struct {
	cu      *CompilationUnit
	log     log.StructuredLogger
	symbols map[string]*SymbolTable
}

// NewChecker creates a new Checker.
func NewChecker(cu *CompilationUnit, logger log.StructuredLogger) *Checker {
	return &Checker{
		cu:      cu,
		log:     logger,
		symbols: make(map[string]*SymbolTable),
	}
}

// CompilationUnit returmod the checker's compilation unit.
func (c *Checker) CompilationUnit() *CompilationUnit {
	return c.cu
}

// TypeOf returmod the type of an expression in a given module.
func (c *Checker) TypeOf(expr ast.Expr, mod *Module) Type {
	return c.typeOf(expr, mod)
}

// Sources returmod the source code of all files processed by the checker.
func (c *Checker) Sources() map[string][]rune {
	return c.cu.Sources
}

// Check initiates the checking process for a given path.
func (c *Checker) Check(path string) {
	isDir, err := c.cu.loader.IsDir(path)
	if err != nil {
		c.cu.AddError(path, 0, 0, fmt.Sprintf("invalid path: %v", err))
		return
	}

	if isDir {
		pkg, err := c.cu.LoadPackage(path)
		if err != nil {
			c.cu.AddError(path, 0, 0, fmt.Sprintf("could not load package '%s': %v", path, err))
			return
		}
		for _, mod := range pkg.Modules {
			c.symbols[mod.Name] = NewSymbolTable(mod.Name)
			c.collectSymbols(mod.File, c.symbols[mod.Name])
		}
		for _, mod := range pkg.Modules {
			c.checkNode(mod.File, mod)
		}
	} else {
		mod, _ := c.cu.LoadFile(path)
		if mod != nil {
			c.symbols[path] = NewSymbolTable(path)
			c.collectSymbols(mod.File, c.symbols[path])
			// Use the Namespaces map for backward compatibility
			if modLegacy, ok := c.cu.Namespaces[path]; ok {
				c.checkNode(mod.File, modLegacy)
			}
		}
	}
}

// CheckSource initiates the checking process for a given source content.
func (c *Checker) CheckSource(content []byte, path string) {
	mod := c.cu.LoadSource(content, path)
	if mod != nil {
		c.symbols[path] = NewSymbolTable(path)
		c.collectSymbols(mod.File, c.symbols[path])
		// Use the Namespaces map for backward compatibility
		if modLegacy, ok := c.cu.Namespaces[path]; ok {
			c.checkNode(mod.File, modLegacy)
		}
	}
}

func (c *Checker) collectSymbols(node ast.Node, st *SymbolTable) {
	if node == nil {
		return
	}
	ast.Walk(node, func(n, parent ast.Node) { // Changed signature here
		switch decl := n.(type) {
		case *ast.RuleDecl:
			if decl.Name != nil {
				isPublic := unicode.IsUpper(rune(decl.Name.Name[0]))
				symbol := &Symbol{
					Name:     decl.Name.Name,
					Kind:     RuleSymbol,
					IsPublic: isPublic,
					Pos:      decl.Name.Pos(),
					IsUsed:   false, // Private rules are not used by default
				}
				st.Add(symbol)
			}
		case *ast.BindingDecl:
			if decl.Name != nil {
				symbol := &Symbol{
					Name:     decl.Name.Name,
					Kind:     BindingSymbol,
					IsPublic: false,
					Pos:      decl.Name.Pos(),
					IsUsed:   false,
				}
				st.Add(symbol)
			}
		}
	})
}

func (c *Checker) checkNode(node ast.Node, mod *Module) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.File:
		st, ok := c.symbols[mod.Name]
		if !ok {
			c.log.Debug("UNREACHABLE", nil)
			return // Should not happen
		}

		for _, decl := range n.Decls {
			c.checkNode(decl, mod)
		}

		// After checking all nodes, analyze the symbol table for unused symbols.
		for _, symbol := range st.Symbols {
			if !symbol.IsUsed {
				line, col := token.FindLineAndCol(symbol.Pos, c.cu.Sources[mod.Name])
				c.cu.AddWarning(mod.Name, line, col, fmt.Sprintf("unused symbol: %s", symbol.Name))
			}
		}

	case *ast.RuleDecl:
		if st, ok := c.symbols[mod.Name]; ok {
			if symbol, found := st.Find(n.Name.Name); found && symbol.IsPublic {
				symbol.IsUsed = true
			}
		}
		if n.Body != nil {
			for _, expr := range n.Body {
				c.checkNode(expr, mod)
			}
		}
	case *ast.SequenceExpr:
		for _, expr := range n.Exprs {
			c.checkNode(expr, mod)
		}
	case *ast.AlternativeExpr:
		for _, expr := range n.Exprs {
			c.checkNode(expr, mod)
		}
	case *ast.OptionalExpr:
		c.checkNode(n.Expr, mod)
	case *ast.RepetitionExpr:
		c.checkNode(n.Expr, mod)
	case *ast.GroupExpr:
		c.checkNode(n.Expr, mod)
	case *ast.ExternalValue:
		// No children to check.
	case *ast.Ident:
		if _, found := mod.Members[n.Name]; !found {
			line, col := token.FindLineAndCol(n.Pos(), c.cu.Sources[mod.Name])
			c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined identifier: %s", n.Name))
		} else {
			// Mark symbol as used
			if st, ok := c.symbols[mod.Name]; ok {
				if symbol, found := st.Find(n.Name); found {
					symbol.IsUsed = true
				}
			}
		}
	case *ast.MemberExpr:
		// Mark the object of the member expression as used.
		if ident, isIdent := n.Object.(*ast.Ident); isIdent {
			if st, ok := c.symbols[mod.Name]; ok {
				if symbol, found := st.Find(ident.Name); found {
					symbol.IsUsed = true
				}
			}
		}

		receiverType := c.typeOf(n.Object, mod)
		if receiverType == nil {
			return // Error already reported
		}

		// Handle different types: NamespaceType (deprecated), ModuleType, PackageType
		switch rt := receiverType.(type) {
		case *NamespaceType:
			// Legacy namespace-based import
			importedNs, found := c.cu.Namespaces[rt.Name]
			if !found {
				line, col := token.FindLineAndCol(n.Object.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("internal error: could not find namespace %s", rt.Name))
				return
			}
			if _, found := importedNs.Members[n.Member.Name]; !found {
				line, col := token.FindLineAndCol(n.Member.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined member '%s' in namespace '%s'", n.Member.Name, rt.Name))
			}
		case *ModuleType:
			// Module-based import (file-based)
			importedMod, found := c.cu.Modules[rt.Name]
			if !found {
				line, col := token.FindLineAndCol(n.Object.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("internal error: could not find module %s", rt.Name))
				return
			}
			if _, found := importedMod.Members[n.Member.Name]; !found {
				line, col := token.FindLineAndCol(n.Member.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined member '%s' in module '%s'", n.Member.Name, rt.Name))
			}
		case *PackageType:
			// Package-based import (directory)
			pkg, found := c.cu.Packages[rt.Path]
			if !found {
				line, col := token.FindLineAndCol(n.Object.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("internal error: could not find package %s", rt.Name))
				return
			}
			module, found := pkg.Modules[n.Member.Name]
			if !found {
				line, col := token.FindLineAndCol(n.Member.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined module '%s' in package '%s'", n.Member.Name, rt.Name))
				return
			}
			// Now check for the rule in the module
			_ = module // Module found, rule lookup would be next level (e.g., pkg.Module.rule)
		default:
			line, col := token.FindLineAndCol(n.Object.Pos(), c.cu.Sources[mod.Name])
			c.cu.AddError(mod.Name, line, col, fmt.Sprintf("expected a namespace, module, or package, but got %s", receiverType.String()))
		}
	}
}

func (c *Checker) typeOf(expr ast.Expr, mod *Module) Type {
	switch e := expr.(type) {
	case *ast.Ident:
		if typ, found := mod.Types[e.Name]; found {
			return typ
		}
		line, col := token.FindLineAndCol(e.Pos(), c.cu.Sources[mod.Name])
		c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined identifier: %s", e.Name))
		return nil
	case *ast.StringLit:
		return String
	case *ast.RegexLit:
		return Regexp
	case *ast.ExternalValue:
		return External
	case *ast.MemberExpr:
		receiverType := c.typeOf(e.Object, mod)
		if receiverType == nil {
			return nil
		}

		// Handle different receiver types
		switch rt := receiverType.(type) {
		case *NamespaceType:
			importedNs := c.cu.Namespaces[rt.Name]
			if importedNs == nil {
				line, col := token.FindLineAndCol(e.Object.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("internal error: could not find namespace '%s'", rt.Name))
				return nil
			}
			if memberType, found := importedNs.Types[e.Member.Name]; found {
				return memberType
			}
			line, col := token.FindLineAndCol(e.Member.Pos(), c.cu.Sources[mod.Name])
			c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined member '%s' in namespace '%s'", e.Member.Name, rt.Name))
			return nil

		case *ModuleType:
			importedMod := c.cu.Modules[rt.Name]
			if importedMod == nil {
				line, col := token.FindLineAndCol(e.Object.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("internal error: could not find module '%s'", rt.Name))
				return nil
			}
			if memberType, found := importedMod.Types[e.Member.Name]; found {
				return memberType
			}
			line, col := token.FindLineAndCol(e.Member.Pos(), c.cu.Sources[mod.Name])
			c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined member '%s' in module '%s'", e.Member.Name, rt.Name))
			return nil

		case *PackageType:
			// Accessing a module in a package: pkg.Module
			pkg := c.cu.Packages[rt.Path]
			if pkg == nil {
				line, col := token.FindLineAndCol(e.Object.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("internal error: could not find package '%s'", rt.Name))
				return nil
			}
			module, found := pkg.Modules[e.Member.Name]
			if !found {
				line, col := token.FindLineAndCol(e.Member.Pos(), c.cu.Sources[mod.Name])
				c.cu.AddError(mod.Name, line, col, fmt.Sprintf("undefined module '%s' in package '%s'", e.Member.Name, rt.Name))
				return nil
			}
			// Return the module's types - the next MemberExpr would access rules within the module
			// For now, we return a placeholder type representing the module
			return &ModuleType{Name: module.Name}

		default:
			line, col := token.FindLineAndCol(e.Object.Pos(), c.cu.Sources[mod.Name])
			c.cu.AddError(mod.Name, line, col, fmt.Sprintf("expected a namespace, module, or package, but got %s", receiverType.String()))
			return nil
		}
	default:
		return nil
	}
}
