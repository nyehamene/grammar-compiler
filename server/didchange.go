package server

import (
	"context"
	"encoding/json"
	"fmt"
	"grammar/log"
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

func (s *Server) handleDidChange(ctx context.Context, rawMsg map[string]any) error {
	var params DidChangeTextDocumentParams
	if rawMsg["params"] == nil {
		return fmt.Errorf("missing params for textDocument/didChange notification")
	}
	encodedParams, err := json.Marshal(rawMsg["params"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(encodedParams, &params); err != nil {
		return err
	}

	if len(params.ContentChanges) == 0 {
		return fmt.Errorf("no content changes found for textDocument/didChange")
	}
	documentUri := params.TextDocument.URI.String()
	s.logger.Debug("Updated document via didChange", log.Fields{"version": params.TextDocument.Version, "file": filepath.Base(documentUri)})

	// For full sync, the new text is the first and only content change.
	s.checker.CompilationUnit().RemoveNamespace(documentUri)
	newText := []rune(params.ContentChanges[0].Text)
	s.documents[params.TextDocument.URI] = &document{
		text: newText,
	}
	s.publishDiagnostics(ctx, params.TextDocument.URI)
	return nil
}
