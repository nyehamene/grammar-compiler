package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"grammar/log"
)

func (s *Server) handleDefinition(id int, msg map[string]any) {
	method := "unknown"
	if m, ok := msg["method"].(string); ok {
		method = m
	}
	var params DefinitionParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal definition params", method)
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal definition params", method)
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/definition", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.logger.Debug("handleDefinition: Document not open", log.Fields{"uri": params.TextDocument.URI.String()})
		s.sendResponse(id, method, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)
	uriStr := params.TextDocument.URI.String()

	// First check Namespaces (deprecated), then Modules
	ns, nsOk := s.checker.CompilationUnit().Namespaces[uriStr]
	if !nsOk || ns == nil || ns.File == nil {
		// Check Modules
		if mod, modOk := s.checker.CompilationUnit().Modules[uriStr]; modOk && mod != nil && mod.File != nil {
			// Use Module directly
			ns = mod
			nsOk = true
		}
	}

	if !nsOk || ns == nil || ns.File == nil {
		s.logger.Debug("handleDefinition: No AST available", log.Fields{"uri": uriStr})
		s.sendResponse(id, method, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, parent := ast.FindNodeAt(ns.File, pos)
	if node == nil {
		s.sendResponse(id, method, nil, nil)
		return
	}

	defDecl := s.checker.FindDefinition(ns, node, parent)
	if defDecl == nil {
		s.sendResponse(id, method, nil, nil)
		return
	}

	// Find the namespace/module (and thus, the file) that contains the definition.
	var defNsURI string

	// First check Namespaces (deprecated)
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

	// If not found in Namespaces, check Modules
	if defNsURI == "" {
		for path, mod := range s.checker.CompilationUnit().Modules {
			if mod.File != nil {
				for _, decl := range mod.File.Decls {
					if decl == defDecl {
						defNsURI = path
						break
					}
				}
			}
			if defNsURI != "" {
				break
			}
		}
	}

	if defNsURI == "" {
		s.logger.Debug("Could not find file for definition.", nil)
		s.sendResponse(id, method, nil, nil)
		return
	}

	defUri, err := ParseURI(defNsURI)
	if err != nil {
		s.sendErrorResponse(id, InternalError, fmt.Sprintf("could not parse URI for definition: %s", defNsURI), method)
		return
	}

	defSrcRunes, ok := s.checker.Sources()[defNsURI]
	if !ok {
		s.sendErrorResponse(id, InternalError, "could not find source for definition file", method)
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
		s.sendResponse(id, method, nil, nil)
		return
	}

	location := Location{
		URI: defUri,
		Range: Range{
			Start: PosToPosition(defNameNode.Pos(), defSrcRunes),
			End:   PosToPosition(defNameNode.End(), defSrcRunes),
		},
	}

	s.sendResponse(id, method, location, nil)
}
