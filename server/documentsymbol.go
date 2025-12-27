package server

import (
	"encoding/json"
	"grammar/ast"
	"path/filepath"
)

func (s *Server) handleDocumentSymbol(id int, msg map[string]any) {
	var params DocumentSymbolParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal documentSymbol params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal documentSymbol params")
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/documentSymbol")
		return
	}

	uriStr := params.TextDocument.URI.String()
	ns, ok := s.checker.CompilationUnit().Namespaces[uriStr]
	if !ok || ns == nil || ns.File == nil {
		s.log.Printf("handleDocumentSymbol: No AST available for %s.", uriStr)
		s.sendResponse(id, nil, nil) // No AST available
		return
	}
	srcRunes, ok := s.checker.CompilationUnit().Sources[uriStr]
	if !ok {
		s.log.Printf("handleDocumentSymbol: No source available for %s.", uriStr)
		s.sendResponse(id, nil, nil) // No source available
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

	s.sendResponse(id, symbols, nil)
	s.log.Printf("sent %d document symbols for %s", len(symbols), filepath.Base(params.TextDocument.URI.Path))
}
