package check

import (
	"grammar/ast"
	"grammar/token"
)

// FindReferences finds all references to the symbol at the given position.
func (c *Checker) FindReferences(uri string, pos token.Pos) []ast.Node {
	ns, ok := c.cu.Namespaces[uri]
	if !ok || ns == nil || ns.File == nil {
		return nil
	}

	node, parent := ast.FindNodeAt(ns.File, pos)
	if node == nil {
		return nil
	}

	// 1. Find the canonical declaration of the symbol under the cursor.
	defDecl := c.FindDefinition(ns, node, parent)
	if defDecl == nil {
		return nil
	}

	var references []ast.Node
	// 2. Iterate through all files in the compilation unit.
	for _, currentNs := range c.cu.Namespaces {
		if currentNs == nil || currentNs.File == nil {
			continue
		}

		// 3. Walk the AST of the current file and find identifiers that refer to the same declaration.
		ast.Walk(currentNs.File, func(n, p ast.Node) {
			if n == nil {
				return
			}

			// We only care about identifiers
			ident, isIdent := n.(*ast.Ident)
			if !isIdent {
				return
			}

			// Check if this identifier resolves to our target declaration.
			foundDecl := c.FindDefinition(currentNs, ident, p)
			if foundDecl == defDecl {
				references = append(references, ident)
			}
		})
	}

	return references
}
