package server

import (
	"context"
	"grammar/check"
	"grammar/token"
)

func (s *Server) publishDiagnostics(ctx context.Context, uri DocumentUri) {
	content, ok := s.GetDocumentContent(uri)
	if !ok {
		// Document not open, nothing to do.
		return
	}

	pathAbsLocal := uri.String()
	if uri.Scheme == "file" {
		pathAbsLocal = uri.Path
	}

	diagnostics := []Diagnostic{}
	errors, _ := check.CheckDocument([]byte(content), pathAbsLocal)

	srcRunes := []rune(content)

	for _, err := range errors {

		s.log.Printf("error document: %s\npath: %s\n", pathAbsLocal, err.Path)

		if err.Path != pathAbsLocal {
			continue
		}

		line, col := token.FindLineAndCol(int(err.Pos), srcRunes)
		diagnostic := Diagnostic{
			Range: Range{
				Start: Position{Line: line - 1, Character: col - 1},
				End:   Position{Line: line - 1, Character: col}, // Highlight one character
			},
			Severity: SeverityError,
			Source:   "checker",
			Message:  err.Message,
		}
		diagnostics = append(diagnostics, diagnostic)
	}

	s.notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}
