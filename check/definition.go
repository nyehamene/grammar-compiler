package check

import (
	"grammar/ast"
)

// FindDefinition finds the declaration of a given node.
func (c *Checker) FindDefinition(ns *Namespace, node ast.Node, parent ast.Node) ast.Decl {
	if node == nil {
		return nil
	}

	ident, ok := node.(*ast.Ident)
	if !ok {
		// Not an identifier, can't find definition for it in this context.
		return nil
	}

	// Case 1: The identifier is the name of a declaration (e.g., the `foo` in `foo = "bar";`).
	// The definition is the declaration itself.
	if decl, isDecl := parent.(ast.Decl); isDecl {
		switch d := decl.(type) {
		case *ast.RuleDecl:
			if d.Name == ident {
				return decl
			}
		case *ast.BindingDecl:
			if d.Name == ident {
				return decl
			}
		}
	}

	// Case 2: It's a member access like `b.rule_b`. We are on `rule_b`.
	if memberExpr, isMember := parent.(*ast.MemberExpr); isMember && memberExpr.Member == ident {
		receiverType := c.typeOf(memberExpr.Object, ns)
		if receiverType == nil {
			return nil // Error already reported by checker
		}
		nsType, isNs := receiverType.(*NamespaceType)
		if !isNs {
			return nil
		}
		importedNs, found := c.cu.Namespaces[nsType.Name]
		if !found {
			return nil
		}
		// Look for the member in the imported namespace.
		if def, found := importedNs.Members[ident.Name]; found {
			return def
		}
		return nil
	}

	// Case 3: It's a simple identifier used in an expression, like the `bar` in `foo = bar;`
	if def, found := ns.Members[ident.Name]; found {
		return def
	}

	return nil
}
