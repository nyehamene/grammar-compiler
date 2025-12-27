package server

import (
	"context"
	"encoding/json"
	"fmt"
)

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     DocumentUri `json:"uri"`
	Version int         `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	// For now, we only support full text synchronization.
	// The range and rangeLength fields are ignored.
	Text string `json:"text"`
}

func (s *Server) handleDidChange(ctx context.Context, msg map[string]any) error {
	var params DidChangeTextDocumentParams
	if p, ok := msg["params"]; ok {
		encodedParams, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(encodedParams, &params); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("missing params for textDocument/didChange notification")
	}

	if len(params.ContentChanges) == 0 {
		return fmt.Errorf("no content changes found for textDocument/didChange")
	}

	// For full sync, the new text is the first and only content change.
	s.documents[params.TextDocument.URI] = params.ContentChanges[0].Text
	s.log.Printf("Updated document via didChange: %s", params.TextDocument.URI)
	s.publishDiagnostics(ctx, params.TextDocument.URI)

	return nil
}