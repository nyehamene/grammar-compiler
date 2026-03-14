package server

import (
	"encoding/json"
	"grammar/ast"
	"grammar/check"
	"grammar/token"
	"unicode" // Import unicode package
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
		s.logger.Printf("handleCompletion: Document not open %s", params.TextDocument.URI.String())
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
		s.logger.Printf("handleCompletion: No AST available for %s. ok: %t", params.TextDocument.URI.String(), ok)
		s.sendResponse(id, method, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, parent := ast.FindNodeAt(mod.File, pos)

	s.logger.Printf("Completion node at position: %#v %#v", node, parent)

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
			s.logger.Printf("Completion: Identified receiver: %s", identName)

			// Create a dummy Ident node for type checking. Position doesn't matter much here.
			dummyIdent := &ast.Ident{NamePos: token.NoPos, Name: identName}
			typ := s.checker.TypeOf(dummyIdent, mod)
			s.logger.Printf("Completion: Type of receiver (%s): %T", identName, typ)

			// Handle NamespaceType (deprecated), ModuleType, or PackageType
			switch rt := typ.(type) {
			case *check.NamespaceType:
				s.logger.Printf("Completion: Receiver is NamespaceType: %s", rt.Name)
				importedNs, found := cu.Namespaces[rt.Name]
				if !found {
					s.logger.Printf("Completion: Imported namespace %s not found in CU", rt.Name)
					return nil
				}
				s.logger.Printf("Completion: Found imported namespace: %s. Members: %v", rt.Name, importedNs.Members)
				for name, decl := range importedNs.Members {
					if _, isRule := decl.(*ast.RuleDecl); isRule {
						s.logger.Printf("Completion: Adding member: %s", name)
						items = append(items, CompletionItem{
							Label: name,
							Kind:  FunctionCompletion,
						})
					}
				}
			case *check.ModuleType:
				s.logger.Printf("Completion: Receiver is ModuleType: %s", rt.Name)
				importedMod, found := cu.Modules[rt.Name]
				if !found {
					s.logger.Printf("Completion: Imported module %s not found in CU", rt.Name)
					return nil
				}
				s.logger.Printf("Completion: Found imported module: %s. Members: %v", rt.Name, importedMod.Members)
				for name, decl := range importedMod.Members {
					if _, isRule := decl.(*ast.RuleDecl); isRule {
						s.logger.Printf("Completion: Adding member: %s", name)
						items = append(items, CompletionItem{
							Label: name,
							Kind:  FunctionCompletion,
						})
					}
				}
			case *check.PackageType:
				s.logger.Printf("Completion: Receiver is PackageType: %s", rt.Name)
				pkg, found := cu.Packages[rt.Path]
				if !found {
					s.logger.Printf("Completion: Package %s not found in CU", rt.Name)
					return nil
				}
				s.logger.Printf("Completion: Found package: %s. Modules: %v", rt.Name, pkg.Modules)
				for name := range pkg.Modules {
					s.logger.Printf("Completion: Adding module: %s", name)
					items = append(items, CompletionItem{
						Label: name,
						Kind:  ClassCompletion,
					})
				}
			default:
				s.logger.Printf("Completion: Receiver %s is not a known type, got %T", identName, typ)
				return nil
			}

			return items // Only return member completiomod if this context applies

		}

		s.logger.Printf("Completion: No identifier found before the dot.")
	}

	s.logger.Printf("Completion: Falling back to rule body/top-level for %s", mod.Name)
	s.logger.Printf("Completion: Namespace members: %v", mod.Members)

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
