package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"grammar/check"
	"grammar/log"
	"grammar/token"
	"unicode"
)

func (s *Server) handleCompletion(id int, rawMsg map[string]any) {
	method := "textDocument/completion"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params CompletionParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/completion", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal completion params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal completion params", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.logger.Debug("handleCompletion: Document not open", log.Fields{"uri": params.TextDocument.URI.String()})
		s.sendResponse(id, method, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)

	mod, ok := s.checker.CompilationUnit().Namespaces[params.TextDocument.URI.String()]
	if !ok || mod == nil || mod.File == nil {
		// Try Modules
		if mod, modOk := s.checker.CompilationUnit().Modules[params.TextDocument.URI.String()]; modOk && mod != nil && mod.File != nil {
			mod = mod
			ok = true
		}
	}
	if !ok || mod == nil || mod.File == nil {
		s.logger.Debug("handleCompletion: No AST available", log.Fields{"uri": params.TextDocument.URI.String(), "ok": ok})
		s.sendResponse(id, method, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, parent := ast.FindNodeAt(mod.File, pos)

	s.logger.Debug("Completion node at position", log.Fields{"node": fmt.Sprintf("%#v", node), "parent": fmt.Sprintf("%#v", parent)})

	items := s.getCompletiomod(node, parent, mod, s.checker.CompilationUnit(), params.Position, srcRunes)

	completionList := CompletionList{
		IsIncomplete: false,
		Items:        items,
	}

	s.sendResponse(id, method, completionList, nil)
}

func (s *Server) getCompletiomod(node, parent ast.Node, mod *check.Module, cu *check.CompilationUnit, cursorPosition Position, srcRunes []rune) []CompletionItem {
	var items []CompletionItem

	offset := PositionToPos(cursorPosition, srcRunes)

	// Check if cursor is after a '.' (member access)
	if offset > 0 && srcRunes[offset-1] == '.' {
		// Find the start of the identifier preceding the dot.
		identEnd := offset - 1 // Position of the dot
		identStart := identEnd
		for identStart > 0 &&
			(unicode.IsLetter(srcRunes[identStart-1]) ||
				unicode.IsDigit(srcRunes[identStart-1]) ||
				srcRunes[identStart-1] == '_') {
			identStart--
		}

		if identStart < identEnd { // If an identifier was found before the dot
			identName := string(srcRunes[identStart:identEnd])
			s.logger.Debug("Completion: Identified receiver", log.Fields{"identName": identName})

			// Create a dummy Ident node for type checking. Position doesn't matter much here.
			dummyIdent := &ast.Ident{NamePos: token.NoPos, Name: identName}
			typ := s.checker.TypeOf(dummyIdent, mod)
			s.logger.Debug("Completion: Type of receiver", log.Fields{"identName": identName, "type": fmt.Sprintf("%T", typ)})

			// Handle NamespaceType (deprecated), ModuleType, or PackageType
			switch rt := typ.(type) {
			case *check.NamespaceType:
				s.logger.Debug("Completion: Receiver is NamespaceType", log.Fields{"name": rt.Name})
				importedNs, found := cu.Namespaces[rt.Name]
				if !found {
					s.logger.Debug("Completion: Imported namespace not found in CU", log.Fields{"name": rt.Name})
					return nil
				}
				s.logger.Debug("Completion: Found imported namespace", log.Fields{"name": rt.Name, "members": importedNs.Members})
				for name, decl := range importedNs.Members {
					if _, isRule := decl.(*ast.RuleDecl); isRule {
						s.logger.Debug("Completion: Adding member", log.Fields{"name": name})
						items = append(items, CompletionItem{
							Label: name,
							Kind:  FunctionCompletion,
						})
					}
				}
			case *check.ModuleType:
				s.logger.Debug("Completion: Receiver is ModuleType", log.Fields{"name": rt.Name})
				importedMod, found := cu.Modules[rt.Name]
				if !found {
					s.logger.Debug("Completion: Imported module not found in CU", log.Fields{"name": rt.Name})
					return nil
				}
				s.logger.Debug("Completion: Found imported module", log.Fields{"name": rt.Name, "members": importedMod.Members})
				for name, decl := range importedMod.Members {
					if _, isRule := decl.(*ast.RuleDecl); isRule {
						s.logger.Debug("Completion: Adding member", log.Fields{"name": name})
						items = append(items, CompletionItem{
							Label: name,
							Kind:  FunctionCompletion,
						})
					}
				}
			case *check.PackageType:
				s.logger.Debug("Completion: Receiver is PackageType", log.Fields{"name": rt.Name})
				pkg, found := cu.Packages[rt.Path]
				if !found {
					s.logger.Debug("Completion: Package not found in CU", log.Fields{"name": rt.Name})
					return nil
				}
				s.logger.Debug("Completion: Found package", log.Fields{"name": rt.Name, "modules": pkg.Modules})
				for name := range pkg.Modules {
					s.logger.Debug("Completion: Adding module", log.Fields{"name": name})
					items = append(items, CompletionItem{
						Label: name,
						Kind:  ClassCompletion,
					})
				}
			default:
				s.logger.Debug("Completion: Receiver is not a known type", log.Fields{"identName": identName, "type": fmt.Sprintf("%T", typ)})
				return nil
			}

			return items // Only return member completiomod if this context applies

		}

		s.logger.Debug("Completion: No identifier found before the dot.", nil)
	}

	s.logger.Debug("Completion: Falling back to rule body/top-level", log.Fields{"modName": mod.Name})
	s.logger.Debug("Completion: Namespace members", log.Fields{"members": mod.Members})

	for name, decl := range mod.Members {
		if _, isRule := decl.(*ast.RuleDecl); isRule {
			items = append(items, CompletionItem{
				Label: name,
				Kind:  FunctionCompletion,
			})
		}

		if _, isBinding := decl.(*ast.BindingDecl); isBinding {
			items = append(items, CompletionItem{
				Label: name,
				Kind:  ModuleCompletion,
			})
		}
	}

	return items
}
