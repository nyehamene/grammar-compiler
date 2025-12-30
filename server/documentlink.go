package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"net/url"
	"path/filepath"
)

func (s *Server) handleDocumentLink(id int, rawMsg map[string]any) {
	method := "textDocument/documentLink"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params DocumentLinkParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/documentLink", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal documentLink params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal documentLink params", method)
		return
	}

	uriStr := params.TextDocument.URI.String()
	ns, ok := s.checker.CompilationUnit().Namespaces[uriStr]
	if !ok || ns == nil || ns.File == nil {
		s.logger.Printf("handleDocumentLink: No AST available for %s.", uriStr)
		s.sendResponse(id, method, []DocumentLink{}, nil) // No AST available, return empty list
		return
	}
	srcRunes, ok := s.checker.CompilationUnit().Sources[uriStr]
	if !ok {
		s.logger.Printf("handleDocumentLink: No source available for %s.", uriStr)
		s.sendResponse(id, method, []DocumentLink{}, nil) // No source available, return empty list
		return
	}

	var links []DocumentLink

	ast.Walk(ns.File, func(node, parent ast.Node) {
		if bindingDecl, ok := node.(*ast.BindingDecl); ok {
			if bindingDecl.Path != nil {
				// The path is a string literal, e.g., "my_file.grammar"
				importPathLiteral := bindingDecl.Path.Value
				// Remove quotes
				importPath := importPathLiteral[1 : len(importPathLiteral)-1]

				// Resolve the import path to an absolute URI
				targetURI, err := resolveImportPath(uriStr, importPath)
				if err != nil {
					s.logger.Printf("handleDocumentLink: failed to resolve import path '%s': %v", importPath, err)
					return // Continue walking the AST, skip this link
				}

				links = append(links, DocumentLink{
					Range: Range{
						Start: PosToPosition(bindingDecl.Path.Pos(), srcRunes),
						End:   PosToPosition(bindingDecl.Path.End(), srcRunes),
					},
					Target: targetURI,
				})
			}
		}
	})

	s.sendResponse(id, method, links, nil)
	// s.logger.Printf("sent %d document links for %s", len(links), params.TextDocument.URI.Path) // Logged by sendResponse
}

// resolveImportPath resolves a relative import path to an absolute DocumentUri.
// This logic is adapted from check/compilation.go to be used in the server package.
func resolveImportPath(baseURIStr, importPath string) (DocumentUri, error) {
	baseURI, err := url.Parse(baseURIStr)
	if err != nil {
		return DocumentUri{}, fmt.Errorf("failed to parse base URI: %w", err)
	}

	// If the base URI has no scheme or is not a file URI, treat it as a file path.
	// This handles cases where baseURIStr might be a simple file path string.
	if baseURI.Scheme == "" || baseURI.Scheme == "file" {
		baseDir := filepath.Dir(baseURI.Path)
		absPath := filepath.Join(baseDir, importPath)
		// Clean the path (e.g., remove ../)
		absPath = filepath.Clean(absPath)
		resolvedURI, err := ParseURI("file://" + absPath)
		if err != nil {
			return DocumentUri{}, fmt.Errorf("failed to construct file URI for '%s': %w", absPath, err)
		}
		return resolvedURI, nil
	}

	// For other schemes (e.g., http, git), resolve as a URL.
	impURI, err := url.Parse(importPath)
	if err != nil {
		return DocumentUri{}, fmt.Errorf("invalid import path '%s': %w", importPath, err)
	}
	resolvedURL := baseURI.ResolveReference(impURI)
	resolvedDocURI, err := ParseURI(resolvedURL.String())
	if err != nil {
		return DocumentUri{}, fmt.Errorf("failed to parse resolved URL '%s': %w", resolvedURL.String(), err)
	}
	return resolvedDocURI, nil
}
