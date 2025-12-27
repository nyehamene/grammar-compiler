package server

import (
	"context"
	"grammar/check"
	"grammar/token"
)

func (s *Server) publishDiagnostics(ctx context.Context, uri DocumentUri) {
	content, ok := s.GetDocumentContent(uri)
	if !ok {
		s.log.Printf("no content: %s", uri.Path)
		return
	}

	documentUri := uri.String()

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

	s.notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}
