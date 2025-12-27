package server_test

import (
	"bytes"
	"context"
	"fmt"
	"grammar/server"
	"strings"
	"testing"
	"time"
)

func TestDidChangeNotification(t *testing.T) {
	// Initial document content
	initialContent := "Foo = bar;"
	changedContent := "Foo = baz;"
	docURI := "file:///path/to/test.grammar"

	// Step 1: Initialize the server (this is usually the first message from client)
	initializeMessage := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"processId": 123,
			"rootUri": null,
			"capabilities": {}
		}
	}`

	// Step 2: Simulate didOpen notification
	didOpenMessage := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "textDocument/didOpen",
		"params": {
			"textDocument": {
				"uri": "%s",
				"languageId": "grammar",
				"version": 1,
				"text": "%s"
			}
		}
	}`, docURI, initialContent)

	// Step 3: Simulate didChange notification (full sync)
	didChangeMessage := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "textDocument/didChange",
		"params": {
			"textDocument": {
				"uri": "%s",
				"version": 2
			},
			"contentChanges": [
				{
					"text": "%s"
				}
			]
		}
	}`, docURI, changedContent)

	var in bytes.Buffer
	var out strings.Builder

	srv := server.NewServer(&in, &out)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		srv.Start()
		cancel()
	}()

	// Send Initialize message
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(initializeMessage)))
	in.WriteString(initializeMessage)

	// Send DidOpen message
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(didOpenMessage)))
	in.WriteString(didOpenMessage)

	// Send DidChange message
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(didChangeMessage)))
	in.WriteString(didChangeMessage)

	// Give a little time for the server to process didOpen
	time.Sleep(100 * time.Millisecond)

	// Wait for server to process messages
	<-ctx.Done()

	// Verify that the document content was updated in the server
	textDocumentId, err := server.ParseURI(docURI)
	if err != nil {
		t.Fatal(err)
	}

	updatedContent, ok := srv.GetDocumentContent(textDocumentId)
	if !ok {
		t.Fatalf("Document with URI %s not found in server after didChange", docURI)
	}
	if updatedContent != changedContent {
		t.Errorf("Expected document content to be %q, but got %q", changedContent, updatedContent)
	}
}
