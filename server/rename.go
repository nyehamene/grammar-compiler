package server

import (
	"encoding/json"
	"grammar/ast"
)

func (s *Server) handlePrepareRename(id int, rawMsg map[string]any) {
	method := "textDocument/prepareRename"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params PrepareRenameParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/prepareRename", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal prepareRename params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal prepareRename params", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.sendResponse(id, method, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)
	uriStr := params.TextDocument.URI.String()

	ns, ok := s.checker.CompilationUnit().Namespaces[uriStr]
	if !ok || ns == nil || ns.File == nil {
		s.sendResponse(id, method, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, _ := ast.FindNodeAt(ns.File, pos)

	// A rename is only valid on an identifier.
	if _, isIdent := node.(*ast.Ident); !isIdent {
		s.sendResponse(id, method, nil, nil) // Not a valid symbol to rename.
		return
	}

	// Return the range of the identifier to be renamed.
	renameRange := Range{
		Start: PosToPosition(node.Pos(), srcRunes),
		End:   PosToPosition(node.End(), srcRunes),
	}

	s.sendResponse(id, method, renameRange, nil)
	// Logged by sendResponse: s.logger.Printf("prepared rename for %s", filepath.Base(uriStr))
}

func (s *Server) handleRename(id int, rawMsg map[string]any) {
	method := "textDocument/rename"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params RenameParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/rename", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal rename params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal rename params", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.sendResponse(id, method, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)
	uriStr := params.TextDocument.URI.String()
	pos := PositionToPos(params.Position, srcRunes)

	// Find all references to the symbol at the given position.
	refNodes := s.checker.FindReferences(uriStr, pos)
	if refNodes == nil {
		s.sendResponse(id, method, nil, nil) // No symbol found or no references.
		return
	}

	workspaceEdit := WorkspaceEdit{
		Changes: make(map[string][]TextEdit),
	}

	// Group edits by file URI.
	for _, refNode := range refNodes {
		// Find the file/namespace/module this reference belongs to.
		var foundURI string

		// First check Namespaces (deprecated)
		for path, ns := range s.checker.CompilationUnit().Namespaces {
			var found bool
			if ns.File != nil {
				ast.Walk(ns.File, func(n, p ast.Node) {
					if n == refNode {
						found = true
					}
				})
			}
			if found {
				foundURI = path
				break
			}
		}

		// If not found in Namespaces, check Modules
		if foundURI == "" {
			for path, mod := range s.checker.CompilationUnit().Modules {
				var found bool
				if mod.File != nil {
					ast.Walk(mod.File, func(n, p ast.Node) {
						if n == refNode {
							found = true
						}
					})
				}
				if found {
					foundURI = path
					break
				}
			}
		}

		if foundURI != "" {
			src, ok := s.checker.CompilationUnit().Sources[foundURI]
			if !ok {
				continue
			}
			edit := TextEdit{
				Range: Range{
					Start: PosToPosition(refNode.Pos(), src),
					End:   PosToPosition(refNode.End(), src),
				},
				NewText: params.NewName,
			}
			workspaceEdit.Changes[foundURI] = append(workspaceEdit.Changes[foundURI], edit)
		}
	}

	s.sendResponse(id, method, workspaceEdit, nil)
	// Logged by sendResponse: s.logger.Printf("sent workspace edit with %d files for rename to '%s'", len(workspaceEdit.Changes), params.NewName)
}
