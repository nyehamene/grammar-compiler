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
		severity := SeverityError
		if e.Warning {
			severity = SeverityWarning
		}

		line, col := token.FindLineAndCol(e.Pos, srcRunes)
		diagnostic := Diagnostic{
			Range: Range{
				Start: Position{Line: line - 1, Character: col - 1},
				End:   Position{Line: line - 1, Character: col},
			},
			Severity: severity,
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
}

func (s *Server) handleWorkspaceDiagnostic(id int, rawMsg map[string]any) {
	method := "workspace/diagnostic"
	if m, ok := rawMsg["method"].(string); ok {
		method = m
	}

	var params WorkspaceDiagnosticParams
	if rawMsg["params"] == nil {
		s.sendErrorResponse(id, InvalidRequest, "missing params for workspace/diagnostic", method)
		return
	}
	encoded, err := json.Marshal(rawMsg["params"])
	if err != nil {
		s.sendErrorResponse(id, InternalError, "could not marshal workspaceDiagnostic params", method)
		return
	}
	if err := json.Unmarshal(encoded, &params); err != nil {
		s.sendErrorResponse(id, InternalError, "could not unmarshal workspaceDiagnostic params", method)
		return
	}

	workspaceReport := WorkspaceDiagnosticReport{
		Items: make([]WorkspaceDocumentDiagnosticReport, 0, len(s.documents)),
	}

	// For each document the server knows about, generate diagnostics
	for uri := range s.documents {
		diagnostics := s.generateDiagnosticsForURI(uri)

		// For now, we always send a full report. Version and ResultID are not used yet.
		docReport := WorkspaceDocumentDiagnosticReport{
			URI:   uri,
			Kind:  DocumentDiagnosticReportKindFull,
			Items: diagnostics,
		}
		workspaceReport.Items = append(workspaceReport.Items, docReport)
	}

	s.sendResponse(id, method, workspaceReport, nil)
}
