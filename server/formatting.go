package server

import (
	"encoding/json"
	"fmt"
	"grammar/ast"
	"grammar/token"
	"path/filepath"
	"strings"
)

// DocumentFormattingParams params for textDocument/formatting
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions for formatting a document.
type FormattingOptions struct {
	TabSize                int  `json:"tabSize"`
	InsertSpaces           bool `json:"insertSpaces"`
	TrimTrailingWhitespace bool `json:"trimTrailingWhitespace"`
	InsertFinalNewline     bool `json:"insertFinalNewline"`
	TrimFinalNewlines      bool `json:"trimFinalNewlines"`
}

func (s *Server) handleTextDocumentFormatting(id int, msg map[string]any) {
	method := "unknown"
	if m, ok := msg["method"].(string); ok {
		method = m
	}

	var params DocumentFormattingParams
	if p, ok := msg["params"]; ok {
		encodedParams, err := json.Marshal(p)
		if err != nil {
			s.sendErrorResponse(id, InternalError, fmt.Sprintf("Failed to marshal params: %v", err), method)
			return
		}
		if err := json.Unmarshal(encodedParams, &params); err != nil {
			s.sendErrorResponse(id, InternalError, fmt.Sprintf("Failed to unmarshal params: %v", err), method)
			return
		}
	} else {
		s.sendErrorResponse(id, InvalidParams, "missing params for textDocument/formatting", method)
		return
	}

	content, ok := s.GetDocumentContent(params.TextDocument.URI)
	if !ok {
		s.sendErrorResponse(id, InternalError, fmt.Sprintf("document not found: %s", params.TextDocument.URI), method)
		return
	}

	formattedContent, err := formatContent(content)
	if err != nil {
		// In case of a formatting error, we can return no edits.
		// Or we could send a notification to the user.
		s.logger.Printf("error formatting document %s: %v", params.TextDocument.URI, err)
		s.sendResponse(id, method, nil, nil) // No edits
		return
	}

	// Create a text edit for the entire document
	lines := strings.Split(content, "\n")
	endLine := len(lines) - 1
	endChar := len(lines[endLine])
	fullRange := Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: endLine, Character: endChar},
	}

	textEdit := TextEdit{
		Range:   fullRange,
		NewText: formattedContent,
	}

	s.sendResponse(id, method, []TextEdit{textEdit}, nil)
	s.logger.Printf("text-formatting result %s", filepath.Base(params.TextDocument.URI.Path))
}

func formatContent(content string) (string, error) {
	srcRunes := []rune(content)
	if len(srcRunes) == 0 {
		return "", nil
	}

	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()

	formatterParser := ast.NewFormatterParser(tokens, srcRunes)
	formatFile, err := formatterParser.Parse()
	if err != nil {
		return "", err
	}

	formatter := ast.NewFormatter(formatFile)
	return formatter.Format(), nil
}
