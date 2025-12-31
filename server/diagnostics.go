package server

import (
	"context"
	"encoding/json"
	"grammar/token"
)

func (s *Server) generateDiagnosticsForURI(uri DocumentUri) []Diagnostic {
	content, ok := s.GetDocumentContent(uri)
	if !ok {
		s.logger.Printf("no content for diagnostic generation: %s", uri.Path)
		return []Diagnostic{}
	}

	documentUri := uri.String()

	// Remove old diagnostics and AST for this file to ensure a fresh check.
	s.checker.CompilationUnit().RemoveNamespace(documentUri)

	// Run the checker. This will populate the checker's error list.
	s.checker.CheckSource([]byte(content), documentUri)

	// Get the errors from the checker.
	errors, _ := s.checker.CompilationUnit().Errors[documentUri]

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

		// Log individual diagnostic here (new requirement)
		s.logger.Print(&diagnostic) // Pass Diagnostic struct for structured logging
	}
	return diagnostics
}

func (s *Server) publishDiagnostics(ctx context.Context, uri DocumentUri) {
	diagnostics := s.generateDiagnosticsForURI(uri)
	s.notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
	// Diagnostic logging is now part of generateDiagnosticsForURI, called before notify.
	// The requirement is to log after sending, but `notify` sends to client.
	// So, we log AFTER notify.
	// Let's re-think: The requirement is "Log diagnostic messages, one per line, after they have been sent to the client"
	// `notify` sends the entire PublishDiagnosticsParams. Individual diagnostics should be logged after that.
	// This means generateDiagnosticsForURI should just return diagnostics, and the caller (publishDiagnostics)
	// should log them after `s.notify`.
	for _, diag := range diagnostics {
		s.logger.Print(&diag)
	}
}

func (s *Server) handleDocumentDiagnostic(id int, rawMsg map[string]any) {
	method := "textDocument/diagnostic"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params DocumentDiagnosticParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for textDocument/diagnostic", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal documentDiagnostic params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal documentDiagnostic params", method)
		return
	}

	diagnostics := s.generateDiagnosticsForURI(params.TextDocument.URI)

	report := RelatedFullDocumentDiagnosticReport{
		Kind:  DocumentDiagnosticReportKindFull,
		Items: diagnostics,
	}

	s.sendResponse(id, method, report, nil)
	// Log individual diagnostic here (new requirement), after sending.
	for _, diag := range diagnostics {
		s.logger.Print(&diag)
	}
}
