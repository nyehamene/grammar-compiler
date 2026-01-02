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
	log     log.Logger
	symbols map[string]*SymbolTable
}

// NewChecker creates a new Checker.
func NewChecker(cu *CompilationUnit, logger log.Logger) *Checker {
	return &Checker{
		cu:      cu,
		log:     logger,
		symbols: make(map[string]*SymbolTable),
	}
}

// CompilationUnit returns the checker's compilation unit.
func (c *Checker) CompilationUnit() *CompilationUnit {
	return c.cu
}

// TypeOf returns the type of an expression in a given namespace.
func (c *Checker) TypeOf(expr ast.Expr, ns *Namespace) Type {
	return c.typeOf(expr, ns)
}

// Sources returns the source code of all files processed by the checker.
func (c *Checker) Sources() map[string][]rune {
	return c.cu.Sources
}

// Check initiates the checking process for a given path.
func (c *Checker) Check(path string) {
	ns, _ := c.cu.LoadFile(path)
	if ns != nil {
		c.symbols[path] = NewSymbolTable(path)
		c.collectSymbols(ns.File, c.symbols[path])
		c.checkNode(ns.File, ns)
	}
}

// CheckSource initiates the checking process for a given source content.
func (c *Checker) CheckSource(content []byte, path string) {
	ns := c.cu.LoadSource(content, path)
	if ns != nil {
		c.symbols[path] = NewSymbolTable(path)
		c.collectSymbols(ns.File, c.symbols[path])
		c.checkNode(ns.File, ns)
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

func (c *Checker) checkNode(node ast.Node, ns *Namespace) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.File:
		st, ok := c.symbols[ns.Name]
		if !ok {
			c.log.Printf("UNREACHABLE")
			return // Should not happen
		}

		for _, decl := range n.Decls {
			c.checkNode(decl, ns)
		}

		// After checking all nodes, analyze the symbol table for unused symbols.
		for _, symbol := range st.Symbols {
			if !symbol.IsUsed {
				line, col := token.FindLineAndCol(symbol.Pos, c.cu.Sources[ns.Name])
				c.cu.AddWarning(ns.Name, line, col, fmt.Sprintf("unused symbol: %s", symbol.Name))
			}
		}

		if len(st.PublicRules) > 1 {
			for _, symbol := range st.PublicRules {
				line, col := token.FindLineAndCol(symbol.Pos, c.cu.Sources[ns.Name])
				c.cu.AddWarning(ns.Name, line, col, fmt.Sprintf("more than one public rule in file: %s", symbol.Name))
			}
		}

	case *ast.RuleDecl:
		if st, ok := c.symbols[ns.Name]; ok {
			if symbol, found := st.Find(n.Name.Name); found && symbol.IsPublic {
				symbol.IsUsed = true
			}
		}
		if n.Body != nil {
			for _, expr := range n.Body {
				c.checkNode(expr, ns)
			}
		}
	case *ast.SequenceExpr:
		for _, expr := range n.Exprs {
			c.checkNode(expr, ns)
		}
	case *ast.AlternativeExpr:
		for _, expr := range n.Exprs {
			c.checkNode(expr, ns)
		}
	case *ast.OptionalExpr:
		c.checkNode(n.Expr, ns)
	case *ast.RepetitionExpr:
		c.checkNode(n.Expr, ns)
	case *ast.GroupExpr:
		c.checkNode(n.Expr, ns)
	case *ast.ExternalValue:
		// No children to check.
	case *ast.Ident:
		if _, found := ns.Members[n.Name]; !found {
			line, col := token.FindLineAndCol(n.Pos(), c.cu.Sources[ns.Name])
			c.cu.AddError(ns.Name, line, col, fmt.Sprintf("undefined identifier: %s", n.Name))
		} else {
			// Mark symbol as used
			if st, ok := c.symbols[ns.Name]; ok {
				if symbol, found := st.Find(n.Name); found {
					symbol.IsUsed = true
				}
			}
		}
	case *ast.MemberExpr:
		// Mark the object of the member expression as used.
		if ident, isIdent := n.Object.(*ast.Ident); isIdent {
			if st, ok := c.symbols[ns.Name]; ok {
				if symbol, found := st.Find(ident.Name); found {
					symbol.IsUsed = true
				}
			}
		}

		receiverType := c.typeOf(n.Object, ns)
		if receiverType == nil {
			return // Error already reported
		}
		nsType, ok := receiverType.(*NamespaceType)
		if !ok {
			line, col := token.FindLineAndCol(n.Object.Pos(), c.cu.Sources[ns.Name])
			c.cu.AddError(ns.Name, line, col, fmt.Sprintf("expected a namespace, but got %s", receiverType.String()))
			return
		}
		importedNs, found := c.cu.Namespaces[nsType.Name]
		if !found {
			line, col := token.FindLineAndCol(n.Object.Pos(), c.cu.Sources[ns.Name])
			c.cu.AddError(ns.Name, line, col, fmt.Sprintf("internal error: could not find namespace %s", nsType.Name))
			return
		}
		if _, found := importedNs.Members[n.Member.Name]; !found {
			line, col := token.FindLineAndCol(n.Member.Pos(), c.cu.Sources[ns.Name])
			c.cu.AddError(ns.Name, line, col, fmt.Sprintf("undefined member '%s' in namespace '%s'", n.Member.Name, nsType.Name))
		} else {
			// This is a reference to a symbol in another file, so we don't mark it as used in the current file.
			_ = n
		}
	}
}

func (c *Checker) typeOf(expr ast.Expr, ns *Namespace) Type {
	switch e := expr.(type) {
	case *ast.Ident:
		if typ, found := ns.Types[e.Name]; found {
			return typ
		}
		line, col := token.FindLineAndCol(e.Pos(), c.cu.Sources[ns.Name])
		c.cu.AddError(ns.Name, line, col, fmt.Sprintf("undefined identifier: %s", e.Name))
		return nil
	case *ast.StringLit:
		return String
	case *ast.RegexLit:
		return Regexp
	case *ast.ExternalValue:
		return External
	case *ast.MemberExpr:
		receiverType := c.typeOf(e.Object, ns)
		if receiverType == nil {
			return nil
		}
		nsType, ok := receiverType.(*NamespaceType)
		if !ok {
			line, col := token.FindLineAndCol(e.Object.Pos(), c.cu.Sources[ns.Name])
			c.cu.AddError(ns.Name, line, col, fmt.Sprintf("expected a namespace, but got %s", receiverType.String()))
			return nil
		}
		importedNs := c.cu.Namespaces[nsType.Name]
		if memberType, found := importedNs.Types[e.Member.Name]; found {
			return memberType
		}
		line, col := token.FindLineAndCol(e.Member.Pos(), c.cu.Sources[ns.Name])
		c.cu.AddError(ns.Name, line, col, fmt.Sprintf("undefined member '%s' in namespace '%s'", e.Member.Name, nsType.Name))
		return nil
	default:
		return nil
	}
}
