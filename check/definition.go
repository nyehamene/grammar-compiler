package check

import (
	"grammar/ast"
)

// FindDefinition finds the declaration of a given node.
func (c *Checker) FindDefinition(mod *Module, node ast.Node, parent ast.Node) ast.Decl {
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
		receiverType := c.typeOf(memberExpr.Object, mod)
		if receiverType == nil {
			return nil // Error already reported by checker
		}

		// Handle different receiver types: NamespaceType (deprecated), ModuleType, PackageType
		switch rt := receiverType.(type) {
		case *NamespaceType:
			importedNs, found := c.cu.Namespaces[rt.Name]
			if !found {
				return nil
			}
			if def, found := importedNs.Members[ident.Name]; found {
				return def
			}
		case *ModuleType:
			importedMod, found := c.cu.Modules[rt.Name]
			if !found {
				return nil
			}
			if def, found := importedMod.Members[ident.Name]; found {
				return def
			}
		case *PackageType:
			pkg, found := c.cu.Packages[rt.Path]
			if !found {
				return nil
			}
			// First level: module access (pkg.Module)
			mod, found := pkg.Modules[ident.Name]
			if !found {
				return nil
			}
			// Module found, return nil for now (would need another level of member access for rules)
			_ = mod
		}
		return nil
	}

	// Case 3: It's a simple identifier used in an expression, like the `bar` in `foo = bar;`
	if def, found := mod.Members[ident.Name]; found {
		return def
	}

	return nil
}
