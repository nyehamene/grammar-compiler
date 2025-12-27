package server

import (
	"encoding/json"
)

// InitializeParams represents the parameters of an `initialize` request.
type InitializeParams struct {
	ProcessID    *int               `json:"processId,omitempty"`
	RootURI      *DocumentUri       `json:"rootUri,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities"`
	// ... other fields
}

// ClientCapabilities represents the capabilities of the client.
type ClientCapabilities struct {
	// For now, we don't need to parse the client capabilities in detail.
}

// InitializeResult represents the result of an `initialize` request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerCapabilities represents the capabilities of the server.
type ServerCapabilities struct {
	TextDocumentSync   *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	HoverProvider      bool                     `json:"hoverProvider,omitempty"`
	DefinitionProvider bool                     `json:"definitionProvider,omitempty"`
	ReferencesProvider bool                     `json:"referencesProvider,omitempty"`
	// ... other server capabilities
}

// TextDocumentSyncOptions defines how the server synchronizes text documents.
type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose,omitempty"`
	Change    int  `json:"change,omitempty"` // Full, Incremental
}

const (
	// TextDocumentSyncKindNone means documents are not synced at all.
	TextDocumentSyncKindNone = 0
	// TextDocumentSyncKindFull means documents are synced by sending the full content on change.
	TextDocumentSyncKindFull = 1
	// TextDocumentSyncKindIncremental means documents are synced by sending incremental changes.
	TextDocumentSyncKindIncremental = 2
)

// ServerInfo represents information about the server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

func handleInitializeRequest(s *Server, id int, msg map[string]any) {
	var params InitializeParams
	paramsBytes, err := json.Marshal(msg["params"])
	if err != nil {
		s.sendErrorResponse(id, InvalidParams, "Failed to marshal initialize params")
		return
	}
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		s.log.Println(err)
		s.sendErrorResponse(id, InvalidParams, "Invalid initialize params")
		return
	}

	s.log.Printf("Client capabilities: %v", params.Capabilities)

	// Respond with server capabilities
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    TextDocumentSyncKindFull,
			},
			HoverProvider:      true,
			DefinitionProvider: true,
			ReferencesProvider: true,
		},
		ServerInfo: &ServerInfo{
			Name:    "grammar-lsp",
			Version: "0.1.0",
		},
	}
	s.sendResponse(id, result, nil)
	s.log.Printf("initialize result: %d", id)
}
