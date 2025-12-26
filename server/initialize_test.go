package server_test

import (
	"encoding/json"
	"grammar/server"
	"testing"
)

func TestInitializeRequest(t *testing.T) {
	// Sample initialize request from a client
	requestJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"processId": 1234,
			"rootUri": "file:///path/to/workspace",
			"capabilities": {}
		}
	}`

	// Test Unmarshaling
	var request server.RequestMessage
	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if request.Method != "initialize" {
		t.Errorf("Expected method 'initialize', got %q", request.Method)
	}

	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	var params server.InitializeParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		t.Fatalf("Failed to unmarshal params: %v", err)
	}

	if params.RootURI.String() != "file:///path/to/workspace" {
		t.Errorf("Expected rootUri 'file:///path/to/workspace', got %q", params.RootURI.String())
	}

	// Test Marshaling
	response := server.InitializeResult{
		Capabilities: server.ServerCapabilities{
			TextDocumentSync: &server.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    server.TextDocumentSyncKindFull,
			},
		},
		ServerInfo: &server.ServerInfo{
			Name:    "grammar-lsp",
			Version: "0.1.0",
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// For now, we just check that it marshals without error.
	// A more thorough test would compare the output to a known-good JSON string.
	if len(responseBytes) == 0 {
		t.Error("Marshaled response is empty")
	}
}
