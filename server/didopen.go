package server

import (
	"context"
	"encoding/json"
	"fmt"
)

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        DocumentUri `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

func (s *Server) handleDidOpen(ctx context.Context, msg map[string]any) error {
	var params DidOpenTextDocumentParams
	if p, ok := msg["params"]; ok {
		encodedParams, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(encodedParams, &params); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing params for textDocument/didOpen notification")
	}

	s.documents[params.TextDocument.URI] = params.TextDocument.Text
	s.log.Printf("Opened and stored document: %s", params.TextDocument.URI)
	return nil
}
