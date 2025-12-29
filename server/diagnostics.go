package server

import (
	"context"
	"encoding/json"
	"grammar/check"
	"grammar/token"
)

func (s *Server) generateDiagnosticsForURI(uri DocumentUri) []Diagnostic {
	content, ok := s.GetDocumentContent(uri)
	if !ok {
		s.log.Printf("no content for diagnostic generation: %s", uri.Path)
		return []Diagnostic{}
	}

	documentUri := uri.String()

	// Remove old diagnostics and AST for this file to ensure a fresh check.
	s.checker.CompilationUnit().RemoveNamespace(documentUri)

	// Run the checker. This will populate the checker's error list.
	err := s.checker.CheckSource([]byte(content), documentUri)

	// Get the errors from the checker.
	var errors []check.Error
	if err != nil {
		if list, ok := err.(check.ErrorList); ok {
			errors = list
		} else {
			s.log.Printf("unexpected error: '%s'", err)
		}
	}

	diagnostics := []Diagnostic{}
	srcRunes := []rune(content)

	for _, e := range errors {

		line, col := token.FindLineAndCol(int(e.Pos), srcRunes)
		diagnostic := Diagnostic{
			Range: Range{
				Start: Position{Line: line - 1, Character: col - 1},
				End:   Position{Line: line - 1, Character: col},
			},
			Severity: SeverityError,
			Source:   "checker",
			Message:  e.Message,
		}

		diagnostics = append(diagnostics, diagnostic)

		s.log.Printf("diagnostic-(%d:%d)/(%d:%d) %s",
			diagnostic.Range.Start.Line,
			diagnostic.Range.Start.Character,
			diagnostic.Range.End.Line,
			diagnostic.Range.End.Character,
			diagnostic.Message,
		)
	}
	return diagnostics
}

func (s *Server) publishDiagnostics(ctx context.Context, uri DocumentUri) {
	diagnostics := s.generateDiagnosticsForURI(uri)
	s.notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

func (s *Server) handleDocumentDiagnostic(id int, msg map[string]any) {
	var params DocumentDiagnosticParams
	if p, ok := msg["params"]; ok {
		encoded, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, "could not marshal documentDiagnostic params")
			return
		}
		if err := json.Unmarshal(encoded, &params); err != nil {
			s.sendErrorResponse(id, InternalError, "could not unmarshal documentDiagnostic params")
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/diagnostic")
		return
	}

	diagnostics := s.generateDiagnosticsForURI(params.TextDocument.URI)

	report := RelatedFullDocumentDiagnosticReport{
		Kind:  DocumentDiagnosticReportKindFull,
		Items: diagnostics,
	}

	s.sendResponse(id, report, nil)
	s.log.Printf("sent %d diagnostics for %s", len(diagnostics), params.TextDocument.URI.Path)
}