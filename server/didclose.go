package server

import (
	"context"
	"encoding/json"
	"fmt"
)

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

func (s *Server) handleDidClose(ctx context.Context, msg map[string]any) error {
	var params DidCloseTextDocumentParams
	if p, ok := msg["params"]; ok {
		encodedParams, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(encodedParams, &params); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing params for textDocument/didClose notification")
	}

	delete(s.documents, params.TextDocument.URI)
	s.log.Printf("Closed and removed document: %s", params.TextDocument.URI)

	return nil
}
