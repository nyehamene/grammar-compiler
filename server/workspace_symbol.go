package server

import (
	"encoding/json"
	"grammar/ast"
	"strings"
)

func (s *Server) handleWorkspaceSymbol(id int, msg map[string]any) {
	var params WorkspaceSymbolParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal workspaceSymbol params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal workspaceSymbol params")
			return
		}
	} else {
		// Params can be missing, in which case the query is empty and all symbols should be returned.
		params.Query = ""
	}

	query := strings.ToLower(params.Query)
	symbols := []SymbolInformation{}

	s.log.Printf("handling workspace symbol request with query: '%s'", query)

	for uriStr, ns := range s.checker.CompilationUnit().Namespaces {
		if ns == nil || ns.File == nil {
			continue
		}
		srcRunes, ok := s.checker.CompilationUnit().Sources[uriStr]
		if !ok {
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

	s.sendResponse(id, symbols, nil)
	s.log.Printf("sent %d workspace symbols for query '%s'", len(symbols), query)
}
