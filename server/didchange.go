package server

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	documentUri := params.TextDocument.URI.String()
	s.log.Printf("Updated document via didChange: %d %s", params.TextDocument.Version, filepath.Base(documentUri))

	// For full sync, the new text is the first and only content change.
	s.checker.CompilationUnit().RemoveNamespace(documentUri)
	s.documents[params.TextDocument.URI] = params.ContentChanges[0].Text
	s.publishDiagnostics(ctx, params.TextDocument.URI)
	return nil
}
