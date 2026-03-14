package server

import (
	"encoding/json"
	"grammar/ast"
	"grammar/log"
)

func (s *Server) handleReferences(id int, rawMsg map[string]any) {
	method := "textDocument/references"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params ReferenceParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/references", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal references params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal references params", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.logger.Debug("handleReferences: Document not open", log.Fields{"uri": params.TextDocument.URI.String()})
		s.sendResponse(id, method, nil, nil)
		return
	}
	srcRunes := []rune(content)
	uriStr := params.TextDocument.URI.String()

	pos := PositionToPos(params.Position, srcRunes)

	// Find all reference nodes
	refNodes := s.checker.FindReferences(uriStr, pos)
	if refNodes == nil {
		s.sendResponse(id, method, []Location{}, nil) // Return empty list
		return
	}

	locations := []Location{}
	for _, refNode := range refNodes {
		// Find the file/namespace/module this reference belongs to
		var foundURI string
		var foundSrc []rune

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
				foundSrc = s.checker.CompilationUnit().Sources[path]
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
					foundSrc = s.checker.CompilationUnit().Sources[path]
					break
				}
			}
		}

		if foundURI != "" {
			uri, err := ParseURI(foundURI)
			if err != nil {
				continue // Skip if URI is invalid
			}
			loc := Location{
				URI: uri,
				Range: Range{
					Start: PosToPosition(refNode.Pos(), foundSrc),
					End:   PosToPosition(refNode.End(), foundSrc),
				},
			}
			locations = append(locations, loc)
		}
	}

	// The `includeDeclaration` flag is handled because FindReferences naturally
	// finds the declaration if it's an identifier.
	s.sendResponse(id, method, locations, nil)
	// Logged by sendResponse: s.logger.Printf("sent %d references for %s", len(locations), filepath.Base(params.TextDocument.URI.Path))
}
