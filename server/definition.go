package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"path/filepath"
)

func (s *Server) handleDefinition(id int, msg map[string]any) {
	var params DefinitionParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal definition params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal definition params")
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/definition")
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.log.Printf("handleDefinition: Document not open %s", params.TextDocument.URI.String())
		s.sendResponse(id, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)
	uriStr := params.TextDocument.URI.String()

	ns, ok := s.checker.CompilationUnit().Namespaces[uriStr]
	if !ok || ns == nil || ns.File == nil {
		s.log.Printf("handleDefinition: No AST available for %s.", uriStr)
		s.sendResponse(id, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, parent := FindNodeAt(ns.File, pos)
	if node == nil {
		s.sendResponse(id, nil, nil)
		return
	}

	defDecl := s.checker.FindDefinition(ns, node, parent)
	if defDecl == nil {
		s.sendResponse(id, nil, nil)
		return
	}

	// Find the namespace (and thus, the file) that contains the definition.
	var defNsURI string
	for path, ns := range s.checker.CompilationUnit().Namespaces {
		for _, decl := range ns.File.Decls {
			if decl == defDecl {
				defNsURI = path
				break
			}
		}
		if defNsURI != "" {
			break
		}
	}

	if defNsURI == "" {
		s.log.Printf("Could not find file for definition.")
		s.sendResponse(id, nil, nil)
		return
	}

	defUri, err := ParseURI(defNsURI)
	if err != nil {
		s.sendErrorResponse(id, InternalError, fmt.Sprintf("could not parse URI for definition: %s", defNsURI))
		return
	}

	defSrcRunes, ok := s.checker.Sources()[defNsURI]
	if !ok {
		s.sendErrorResponse(id, InternalError, "could not find source for definition file")
		return
	}

	// The position of the definition is the position of its name identifier.
	var defNameNode ast.Node
	switch d := defDecl.(type) {
	case *ast.RuleDecl:
		defNameNode = d.Name
	case *ast.BindingDecl:
		defNameNode = d.Name
	default:
		s.sendResponse(id, nil, nil)
		return
	}

	location := Location{
		URI: defUri,
		Range: Range{
			Start: PosToPosition(defNameNode.Pos(), defSrcRunes),
			End:   PosToPosition(defNameNode.End(), defSrcRunes),
		},
	}

	s.sendResponse(id, location, nil)
	s.log.Printf("sent definition: %s %s", filepath.Base(params.TextDocument.URI.Path), location.URI)
}
