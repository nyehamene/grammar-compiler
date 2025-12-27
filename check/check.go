package check

import (
	"fmt"
	"grammar/ast"
)

// Checker holds the state for the type-checking process.
type Checker struct {
	cu *CompilationUnit
}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{
		cu: NewCompilationUnit(),
	}
}

// Sources returns the source code of all files processed by the checker.
func (c *Checker) Sources() map[string][]rune {
	return c.cu.Sources
}

// Check initiates the checking process for a given path.
func (c *Checker) Check(path string) error {
	ns, err := c.cu.LoadFile(path)
	if err != nil {
		// Parsing and file reading errors are handled in LoadFile/LoadSource
		// and added to the error list. We just need to check if the namespace
		// could be loaded at all.
	}
	if ns != nil {
		c.checkNode(ns.File, ns)
	}
	return c.cu.Err()
}

// CheckSource initiates the checking process for a given source content.
func (c *Checker) CheckSource(content []byte, path string) error {
	ns, err := c.cu.LoadSource(content, path)
	if err != nil {
		// Errors are handled in LoadSource
	}
	if ns != nil {
		c.checkNode(ns.File, ns)
	}
	return c.cu.Err()
}

func (c *Checker) checkNode(node ast.Node, ns *Namespace) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Decls {
			c.checkNode(decl, ns)
		}
	case *ast.RuleDecl:
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
	case *ast.MemberExpr:
		receiverType := c.typeOf(n.Object, ns)
		if receiverType == nil {
			return // Error already reported
		}
		nsType, ok := receiverType.(*NamespaceType)
		if !ok {
			c.cu.AddError(ns.Name, n.Object.Pos(), fmt.Sprintf("expected a namespace, but got %s", receiverType.String()))
			return
		}
		importedNs, found := c.cu.Namespaces[nsType.Name]
		if !found {
			c.cu.AddError(ns.Name, n.Object.Pos(), fmt.Sprintf("internal error: could not find namespace %s", nsType.Name))
			return
		}
		if _, found := importedNs.Members[n.Member.Name]; !found {
			c.cu.AddError(ns.Name, n.Member.Pos(), fmt.Sprintf("undefined member '%s' in namespace '%s'", n.Member.Name, nsType.Name))
		}
	}
}

func (c *Checker) typeOf(expr ast.Expr, ns *Namespace) Type {
	switch e := expr.(type) {
	case *ast.Ident:
		if typ, found := ns.Types[e.Name]; found {
			return typ
		}
		c.cu.AddError(ns.Name, e.Pos(), fmt.Sprintf("undefined identifier: %s", e.Name))
		return nil
	case *ast.StringLit:
		return String
	case *ast.RegexLit:
		return Regexp
	case *ast.MemberExpr:
		receiverType := c.typeOf(e.Object, ns)
		if receiverType == nil {
			return nil
		}
		nsType, ok := receiverType.(*NamespaceType)
		if !ok {
			c.cu.AddError(ns.Name, e.Object.Pos(), fmt.Sprintf("expected a namespace, but got %s", receiverType.String()))
			return nil
		}
		importedNs := c.cu.Namespaces[nsType.Name]
		if memberType, found := importedNs.Types[e.Member.Name]; found {
			return memberType
		}
		c.cu.AddError(ns.Name, e.Member.Pos(), fmt.Sprintf("undefined member '%s' in namespace '%s'", e.Member.Name, nsType.Name))
		return nil
	default:
		return nil
	}
}
