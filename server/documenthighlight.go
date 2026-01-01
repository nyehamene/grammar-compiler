package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"grammar/token"
)

func handleDocumentHighlight(s *Server, id int, msg map[string]any) {
	method := "textDocument/documentHighlight"

	var params DocumentHighlightParams
	paramsBytes, err := json.Marshal(msg["params"])
	if err != nil {
		s.sendErrorResponse(id, InvalidParams, fmt.Sprintf("Failed to marshal params: %v", err), method)
		return
	}
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		s.sendErrorResponse(id, InvalidParams, fmt.Sprintf("Failed to unmarshal params: %v", err), method)
		return
	}

	doc, ok := s.documents[params.TextDocument.URI]
	if !ok {
		s.sendErrorResponse(id, InvalidRequest, fmt.Sprintf("Document not found: %s", params.TextDocument.URI), method)
		return
	}

	s.checker.Check(params.TextDocument.URI.String()) // Ensure compilation unit is up-to-date
	cu := s.checker.CompilationUnit()
	if cu == nil {
		s.sendResponse(id, method, nil, nil)
		return
	}

	ns, ok := cu.Namespaces[params.TextDocument.URI.String()]
	if !ok || ns.File == nil {
		s.sendResponse(id, method, nil, nil)
		return
	}

	pos := PositionToPos(params.Position, doc.text)

	node, parent := ast.FindNodeAt(ns.File, pos)
	if node == nil {
		s.sendResponse(id, method, nil, nil)
		return
	}

	ident, ok := node.(*ast.Ident)
	if !ok {
		s.sendResponse(id, method, nil, nil)
		return
	}

	// Find the declaration of the symbol under the cursor
	decl := s.checker.FindDefinition(ns, ident, parent)
	if decl == nil {
		s.sendResponse(id, method, nil, nil)
		return
	}

	// Find all references *within the current file* to the identified declaration
	uses := findLocalReferences(ns.File, decl)
	if len(uses) == 0 {
		s.sendResponse(id, method, nil, nil)
		return
	}

	highlights := make([]DocumentHighlight, 0, len(uses))
	for _, useIdent := range uses {
		rng, err := TokenRangeToLSPRange(useIdent.Pos(), useIdent.End(), doc.text)
		if err != nil {
			s.logger.Printf("failed to convert token range to LSP range: %v", err)
			continue
		}

		kind := getHighlightKind(decl, useIdent)
		highlights = append(highlights, DocumentHighlight{
			Range: rng,
			Kind:  &kind,
		})
	}

	s.sendResponse(id, method, highlights, nil)
}

// findLocalReferences traverses the AST of a single file to find all ast.Ident nodes
// that refer to the target declaration.
func findLocalReferences(file *ast.File, targetDecl ast.Decl) []*ast.Ident {
	var references []*ast.Ident

	ast.Walk(file, func(node, parent ast.Node) {
		if id, ok := node.(*ast.Ident); ok {
			// This is a simplified check. In a full type checker, you'd resolve 'id'
			// to its object and compare it with the object of 'targetDecl'.
			// For now, we compare names and assume targetDecl is unambiguous within the file.
			var targetName string
			switch d := targetDecl.(type) {
			case *ast.RuleDecl:
				targetName = d.Name.Name
			case *ast.BindingDecl:
				targetName = d.Name.Name
			}

			if id.Name == targetName {
				references = append(references, id)
			}
		}
	})
	return references
}

// getHighlightKind determines if a given usage (ident) is a 'Write' (declaration) or 'Read' operation.
func getHighlightKind(decl ast.Decl, ident *ast.Ident) DocumentHighlightKind {
	// Compare the position of the usage with the position of the declaration.
	// If they are the same, it's the declaration itself (a 'write').
	var declPos token.Pos
	switch d := decl.(type) {
	case *ast.RuleDecl:
		declPos = d.Name.Pos()
	case *ast.BindingDecl:
		declPos = d.Name.Pos()
	default:
		return Text // Should not happen for valid declarations
	}

	if ident.Pos() == declPos {
		return Write
	}
	return Read
}
