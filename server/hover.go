package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"grammar/check"
	"grammar/token"
	"path/filepath"
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
	node, parent := ast.FindNodeAt(ns.File, pos)
	if node == nil {
		s.log.Printf("handleHover: No node found at position %d", pos)
		s.sendResponse(id, nil, nil)
		return
	}
	s.log.Printf("handleHover: Found node: %+v, Parent: %+v", node, parent)

	var typ string
	var value string

	switch n := node.(type) {
	case *ast.StringLit:
		typ = "string"
		value = sourceOf(n, srcRunes)
	case *ast.RegexLit:
		typ = "regexp"
		value = sourceOf(n, srcRunes)
	case *ast.ExternalValue:
		typ = "external"
		value = sourceOf(n, srcRunes)
	case *ast.Ident:
		if memberExpr, isMember := parent.(*ast.MemberExpr); isMember && memberExpr.Member == n {
			receiverType := s.checker.TypeOf(memberExpr.Object, ns)
			if nsType, ok := receiverType.(*check.NamespaceType); ok {
				if importedNs, found := s.checker.CompilationUnit().Namespaces[nsType.Name]; found {
					if decl, found := importedNs.Members[n.Name]; found {
						if ruleDecl, ok := decl.(*ast.RuleDecl); ok {
							typ = "production"
							importedSrc, srcFound := s.checker.Sources()[nsType.Name]
							if srcFound {
								value = sourceOf(ruleDecl.Body, importedSrc) + ";"
							}
						}
					}
				}
			}
		} else {
			identType := s.checker.TypeOf(n, ns)
			if identType != nil {
				switch identType.(type) {
				case check.ProductionType:
					if decl, found := ns.Members[n.Name]; found {
						if ruleDecl, ok := decl.(*ast.RuleDecl); ok {
							typ = "production"
							value = sourceOf(ruleDecl.Body, srcRunes) + ";"
						}
					}
				case *check.NamespaceType:
					if decl, found := ns.Members[n.Name]; found {
						if bindingDecl, ok := decl.(*ast.BindingDecl); ok {
							typ = "namespace"
							path := bindingDecl.Path.Value
							value = path[1 : len(path)-1] // remove quotes
						}
					}
				}
			}
		}
	}
	var hoverContent string
	if typ != "" && value != "" {
		hoverContent = fmt.Sprintf("(%s)\n\n```grammar\n%s\n```\n", typ, value)
	}

	if hoverContent == "" {
		s.sendResponse(id, nil, nil)
		return
	}

	hover := Hover{
		Contents: MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: hoverContent,
		},
	}

	s.sendResponse(id, hover, nil)
	s.log.Printf("send hover: %s %s", filepath.Base(params.TextDocument.URI.Path), typ)
}

func sourceOf(nodeOrSlice interface{}, src []rune) string {
	var start, end token.Pos

	switch v := nodeOrSlice.(type) {
	case ast.Node:
		start = v.Pos()
		end = v.End()
	case []ast.Expr:
		if len(v) == 0 {
			return ""
		}
		start = v[0].Pos()
		end = v[len(v)-1].End()
	default:
		return ""
	}

	if start < 0 || int(end) > len(src) || start > end {
		return ""
	}
	return string(src[start:end])
}
