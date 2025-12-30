package server

import (
	"context"
	"encoding/json"
	"fmt"
)

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

func (s *Server) handleDidClose(ctx context.Context, rawMsg map[string]any) error {
	var params DidCloseTextDocumentParams
	if rawMsg["params"] == nil {
		return fmt.Errorf("missing params for textDocument/didClose notification")
	}
	encodedParams, err := json.Marshal(rawMsg["params"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(encodedParams, &params); err != nil {
		return err
	}

	delete(s.documents, params.TextDocument.URI)
	s.logger.Printf("Closed and removed document: %s", params.TextDocument.URI)

	return nil
}