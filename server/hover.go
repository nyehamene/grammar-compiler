package server

import (
	"encoding/json"
	"grammar/ast"
	"grammar/check"
	"path/filepath"
	//"log" // Removed log import
)

func (s *Server) handleHover(id int, msg map[string]any) {
	var params HoverParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal hover params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal hover params")
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/hover")
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.log.Printf("handleHover: Document not open %s", params.TextDocument.URI.String())
		s.sendResponse(id, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)

	ns, ok := s.checker.CompilationUnit().Namespaces[params.TextDocument.URI.String()]
	if !ok || ns == nil || ns.File == nil {
		s.log.Printf("handleHover: No AST available for %s. ok: %t", params.TextDocument.URI.String(), ok)
		s.sendResponse(id, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	s.log.Printf("handleHover: Position: %+v, token.Pos: %d", params.Position, pos)

	node, parent := ast.FindNodeAt(ns.File, pos)
	if node == nil {
		s.log.Printf("handleHover: No node found at position %d", pos)
		s.sendResponse(id, nil, nil)
		return
	}
	s.log.Printf("handleHover: Found node: %+v, Parent: %+v", node, parent)

	var typ check.Type
	if ident, isIdent := node.(*ast.Ident); isIdent {
		if p, isMember := parent.(*ast.MemberExpr); isMember && p.Member == ident {
			s.log.Printf("handleHover: Hovering over member of MemberExpr: %+v", p)
			typ = s.checker.TypeOf(p, ns) // p is the MemberExpr 'b.rule_b'
		} else {
			s.log.Printf("handleHover: Hovering over standalone ident: %+v", ident)
			typ = s.checker.TypeOf(ident, ns)
		}
	} else if expr, isExpr := node.(ast.Expr); isExpr {
		s.log.Printf("handleHover: Hovering over non-ident expr: %+v", expr)
		typ = s.checker.TypeOf(expr, ns)
	}
	s.log.Printf("handleHover: Determined type: %+v", typ)

	if typ == nil {
		s.log.Printf("handleHover: Type is nil, sending null response.")
		s.sendResponse(id, nil, nil)
		return
	}

	hover := Hover{
		Contents: MarkupContent{
			Kind:  MarkupKindPlainText,
			Value: typ.String(),
		},
	}

	s.sendResponse(id, hover, nil)
	s.log.Printf("send hover: %s %s", filepath.Base(params.TextDocument.URI.Path), typ.String())
}
