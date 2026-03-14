package server

import (
	"encoding/json"
	"grammar/ast"
	"grammar/log"
)

func (s *Server) handleDocumentSymbol(id int, rawMsg map[string]any) {
	method := "textDocument/documentSymbol"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params DocumentSymbolParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/documentSymbol", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal documentSymbol params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal documentSymbol params", method)
		return
	}

	uriStr := params.TextDocument.URI.String()
	ns, ok := s.checker.CompilationUnit().Namespaces[uriStr]
	if !ok || ns == nil || ns.File == nil {
		s.logger.Debug("handleDocumentSymbol: No AST available", log.Fields{"uri": uriStr})
		s.sendResponse(id, method, nil, nil) // No AST available
		return
	}
	srcRunes, ok := s.checker.CompilationUnit().Sources[uriStr]
	if !ok {
		s.logger.Debug("handleDocumentSymbol: No source available", log.Fields{"uri": uriStr})
		s.sendResponse(id, method, nil, nil) // No source available
		return
	}

	symbols := []DocumentSymbol{}
	for _, decl := range ns.File.Decls {
		var symbol *DocumentSymbol

		switch d := decl.(type) {
		case *ast.RuleDecl:
			symbol = &DocumentSymbol{
				Name: d.Name.Name,
				Kind: SymbolKindField,
				Range: Range{
					Start: PosToPosition(d.Pos(), srcRunes),
					End:   PosToPosition(d.End(), srcRunes),
				},
				SelectionRange: Range{
					Start: PosToPosition(d.Name.Pos(), srcRunes),
					End:   PosToPosition(d.Name.End(), srcRunes),
				},
			}
		case *ast.BindingDecl:
			symbol = &DocumentSymbol{
				Name: d.Name.Name,
				Kind: SymbolKindVariable,
				Range: Range{
					Start: PosToPosition(d.Pos(), srcRunes),
					End:   PosToPosition(d.End(), srcRunes),
				},
				SelectionRange: Range{
					Start: PosToPosition(d.Name.Pos(), srcRunes),
					End:   PosToPosition(d.Name.End(), srcRunes),
				},
			}
		}

		if symbol != nil {
			symbols = append(symbols, *symbol)
		}
	}

	s.sendResponse(id, method, symbols, nil)
	// Logged by sendResponse: s.logger.Printf("sent %d document symbols for %s", len(symbols), filepath.Base(params.TextDocument.URI.Path))
}
