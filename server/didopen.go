package server

import (
	"context"
	"encoding/json"
	"fmt"
	"grammar/log"
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

	s.logger.Info("Opened and stored document", log.Fields{
		"document": filepath.Base(params.TextDocument.URI.Path),
		"version":  params.TextDocument.Version,
	})
	doc := &document{
		text: []rune(params.TextDocument.Text),
	}
	s.documents[params.TextDocument.URI] = doc
	s.publishDiagnostics(ctx, params.TextDocument.URI)
	return nil
}
