package server

import (
	"encoding/json"
	"grammar/ast"
	"path/filepath"
)

func (s *Server) handleReferences(id int, msg map[string]any) {
	var params ReferenceParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal references params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal references params")
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/references")
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.log.Printf("handleReferences: Document not open %s", params.TextDocument.URI.String())
		s.sendResponse(id, nil, nil)
		return
	}
	srcRunes := []rune(content)
	uriStr := params.TextDocument.URI.String()

	pos := PositionToPos(params.Position, srcRunes)

	// Find all reference nodes
	refNodes := s.checker.FindReferences(uriStr, pos)
	if refNodes == nil {
		s.sendResponse(id, []Location{}, nil) // Return empty list
		return
	}

	locations := []Location{}
	for _, refNode := range refNodes {
		// Find the file/namespace this reference belongs to
		var foundURI string
		var foundSrc []rune
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
	s.sendResponse(id, locations, nil)
	s.log.Printf("sent %d references for %s", len(locations), filepath.Base(params.TextDocument.URI.Path))
}
