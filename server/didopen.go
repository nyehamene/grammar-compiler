package server

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
)

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

func (s *Server) handleDidOpen(ctx context.Context, rawMsg map[string]any) error {
	var params DidOpenTextDocumentParams
	if rawMsg["params"] == nil {
		return fmt.Errorf("missing params for textDocument/didOpen notification")
	}
	encodedParams, err := json.Marshal(rawMsg["params"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(encodedParams, &params); err != nil {
		return err
	}

	s.logger.Printf("Opened and stored document: %s version %d", filepath.Base(params.TextDocument.URI.Path), params.TextDocument.Version)
	s.documents[params.TextDocument.URI] = params.TextDocument.Text
	s.publishDiagnostics(ctx, params.TextDocument.URI)
	return nil
}