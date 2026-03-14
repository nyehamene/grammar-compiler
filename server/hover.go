package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"grammar/check"
	"grammar/log"
	"grammar/token"
)

func (s *Server) handleHover(id int, rawMsg map[string]any) {
	method := "textDocument/hover"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params HoverParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/hover", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal hover params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal hover params", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.logger.Debug("handleHover: Document not open", log.Fields{"uri": params.TextDocument.URI.String()})
		s.sendResponse(id, method, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)

	ns, ok := s.checker.CompilationUnit().Namespaces[params.TextDocument.URI.String()]
	if !ok || ns == nil || ns.File == nil {
		s.logger.Debug("handleHover: No AST available", log.Fields{"uri": params.TextDocument.URI.String(), "ok": ok})
		s.sendResponse(id, method, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, parent := ast.FindNodeAt(ns.File, pos)
	if node == nil {
		s.logger.Debug("handleHover: No node found at position", log.Fields{"position": pos})
		s.sendResponse(id, method, nil, nil)
		return
	}
	s.logger.Debug("handleHover: Found node", log.Fields{"node": fmt.Sprintf("%+v", node), "parent": fmt.Sprintf("%+v", parent)})

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
			// Handle NamespaceType (deprecated), ModuleType, or PackageType
			switch rt := receiverType.(type) {
			case *check.NamespaceType:
				if importedNs, found := s.checker.CompilationUnit().Namespaces[rt.Name]; found {
					if decl, found := importedNs.Members[n.Name]; found {
						if ruleDecl, ok := decl.(*ast.RuleDecl); ok {
							typ = "production"
							importedSrc, srcFound := s.checker.Sources()[rt.Name]
							if srcFound {
								value = sourceOf(ruleDecl.Body, importedSrc) + ";"
							}
						}
					}
				}
			case *check.ModuleType:
				if importedMod, found := s.checker.CompilationUnit().Modules[rt.Name]; found {
					if decl, found := importedMod.Members[n.Name]; found {
						if ruleDecl, ok := decl.(*ast.RuleDecl); ok {
							typ = "production"
							importedSrc, srcFound := s.checker.Sources()[rt.Name]
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
				case *check.ModuleType:
					if decl, found := ns.Members[n.Name]; found {
						if bindingDecl, ok := decl.(*ast.BindingDecl); ok {
							typ = "module"
							path := bindingDecl.Path.Value
							value = path[1 : len(path)-1] // remove quotes
						}
					}
				case *check.PackageType:
					if decl, found := ns.Members[n.Name]; found {
						if bindingDecl, ok := decl.(*ast.BindingDecl); ok {
							typ = "package"
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
		s.sendResponse(id, method, nil, nil)
		return
	}

	hover := Hover{
		Contents: MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: hoverContent,
		},
	}

	s.sendResponse(id, method, hover, nil)
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
