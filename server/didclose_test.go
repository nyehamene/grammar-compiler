package server_test

import (
	"bytes"
	"context"
	"fmt"
	"grammar/log"
	"grammar/server"
	"strings"
	"testing"
	"time"
)

func TestDidCloseNotification(t *testing.T) {
	docURI := "file:///path/to/close_test.grammar"
	initialContent := "Foo = bar;"

	// Step 1: Initialize the server
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

	// Step 3: Simulate didClose notification
	didCloseMessage := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "textDocument/didClose",
		"params": {
			"textDocument": {
				"uri": "%s"
			}
		}
	}`, docURI)

	var in bytes.Buffer
	var out strings.Builder
	var logOut bytes.Buffer
	defer func() {
		if t.Failed() {
			t.Log(logOut.String())
		}
	}()

	srv := server.NewServer(&in, &out, log.NewConsoleLogger(&logOut, log.DEBUG))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		srv.Start()
		cancel()
	}()

	// Send Initialize message
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(initializeMessage)))
	in.WriteString(initializeMessage)

	// Send didOpen message
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(didOpenMessage)))
	in.WriteString(didOpenMessage)

	// Send didClose message
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(didCloseMessage)))
	in.WriteString(didCloseMessage)

	// Give a little time for the server to process didOpen
	time.Sleep(100 * time.Millisecond)

	<-ctx.Done()

	// Verify that the document is no longer in the server's documents map
	textDocumentId, err := server.ParseURI(docURI)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := srv.GetDocumentContent(textDocumentId)
	if ok {
		t.Fatalf("Document with URI %s was found in server after didClose", docURI)
	}
}
