package server

import (
	"encoding/json"
	"grammar/ast"
	"grammar/check"
	"grammar/token" // Import token package
	"path/filepath"
	"unicode" // Import unicode package
)

func (s *Server) handleCompletion(id int, msg map[string]any) {
	var params CompletionParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal completion params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal completion params")
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/completion")
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.log.Printf("handleCompletion: Document not open %s", params.TextDocument.URI.String())
		s.sendResponse(id, nil, nil) // Document not open
		return
	}
	srcRunes := []rune(content)

	ns, ok := s.checker.CompilationUnit().Namespaces[params.TextDocument.URI.String()]
	if !ok || ns == nil || ns.File == nil {
		s.log.Printf("handleCompletion: No AST available for %s. ok: %t", params.TextDocument.URI.String(), ok)
		s.sendResponse(id, nil, nil) // No AST available
		return
	}

	pos := PositionToPos(params.Position, srcRunes)
	node, parent := ast.FindNodeAt(ns.File, pos)

	s.log.Printf("Completion node at position: %#v %#v", node, parent)

	items := s.getCompletions(node, parent, ns, s.checker.CompilationUnit(), params.Position, srcRunes)

	completionList := CompletionList{
		IsIncomplete: false,
		Items:        items,
	}

	s.sendResponse(id, completionList, nil)
	s.log.Printf("send completion: %s", filepath.Base(params.TextDocument.URI.Path))
}

func (s *Server) getCompletions(node, parent ast.Node, ns *check.Namespace, cu *check.CompilationUnit, cursorPosition Position, srcRunes []rune) []CompletionItem {
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
			s.log.Printf("Completion: Identified receiver: %s", identName)

			// Create a dummy Ident node for type checking. Position doesn't matter much here.
			dummyIdent := &ast.Ident{NamePos: token.NoPos, Name: identName}
			typ := s.checker.TypeOf(dummyIdent, ns)
			s.log.Printf("Completion: Type of receiver (%s): %T", identName, typ)

			nsType, ok := typ.(*check.NamespaceType)
			if !ok {
				s.log.Printf("Completion: Receiver %s is not NamespaceType, got %T", identName, typ)
				return nil
			}

			s.log.Printf("Completion: Receiver is NamespaceType: %s", nsType.Name)
			importedNs, found := cu.Namespaces[nsType.Name]
			if !found {
				s.log.Printf("Completion: Imported namespace %s not found in CU", nsType.Name)
				return nil
			}

			s.log.Printf("Completion: Found imported namespace: %s. Members: %v", nsType.Name, importedNs.Members)

			for name, decl := range importedNs.Members {
				if _, isRule := decl.(*ast.RuleDecl); isRule {
					s.log.Printf("Completion: Adding member: %s", name)
					items = append(items, CompletionItem{
						Label: name,
						Kind:  FunctionCompletion,
					})
				}
			}

			return items // Only return member completions if this context applies

		}

		s.log.Printf("Completion: No identifier found before the dot.")
	}

	s.log.Printf("Completion: Falling back to rule body/top-level for %s", ns.Name)
	s.log.Printf("Completion: Namespace members: %v", ns.Members)

	for name, decl := range ns.Members {
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
