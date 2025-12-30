package server

import (
	"encoding/json"
	"grammar/ast"
	"path/filepath"
	"strings"
)

func (s *Server) handleWorkspaceSymbol(id int, rawMsg map[string]any) {
	method := "workspace/symbol"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params WorkspaceSymbolParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for workspace/symbol", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal workspaceSymbol params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal workspaceSymbol params", method)
		return
	}

	query := strings.ToLower(params.Query)
	symbols := []SymbolInformation{}

	s.logger.Printf("handling workspace symbol request with query: '%s'", query)

	for uriStr, ns := range s.checker.CompilationUnit().Namespaces {
		if ns == nil || ns.File == nil {
			continue
		}
		srcRunes, ok := s.checker.CompilationUnit().Sources[uriStr]
		if !ok {
			s.logger.Printf("handleWorkspaceSymbol: no source available for '%s'", filepath.Base(uriStr))
			continue
		}
		uri, err := ParseURI(uriStr)
		if err != nil {
			continue
		}

		for _, decl := range ns.File.Decls {
			var name string
			var kind SymbolKind
			var nameNode ast.Node

			switch d := decl.(type) {
			case *ast.RuleDecl:
				name = d.Name.Name
				kind = SymbolKindFunction
				nameNode = d.Name
			case *ast.BindingDecl:
				name = d.Name.Name
				kind = SymbolKindVariable
				nameNode = d.Name
			}

			if name != "" && strings.Contains(strings.ToLower(name), query) {
				symbol := SymbolInformation{
					Name: name,
					Kind: kind,
					Location: Location{
						URI: uri,
						Range: Range{
							Start: PosToPosition(nameNode.Pos(), srcRunes),
							End:   PosToPosition(nameNode.End(), srcRunes),
						},
					},
				}
				symbols = append(symbols, symbol)
			}
		}
	}

	s.sendResponse(id, method, symbols, nil)
	// Logged by sendResponse: s.logger.Printf("sent %d workspace symbols for query '%s'", len(symbols), query)
}
