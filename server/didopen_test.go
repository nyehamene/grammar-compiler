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

func TestDidOpenNotification(t *testing.T) {
	const docURI = "file:///path/to/test.grammar"
	const docContent = `Foo = bar;`

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
	}`, docURI, docContent)

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

	<-ctx.Done()

	// Check if the document was added to the server's state
	textDocumentId, err := server.ParseURI(docURI)
	if err != nil {
		t.Fatal(err)
	}

	storedContent, ok := srv.GetDocumentContent(textDocumentId)
	if !ok {
		t.Fatalf("Document was not added to the server's documents map. Log: %s", out.String())
	}

	if storedContent != docContent {
		t.Errorf("Document content mismatch. Got: '%s', Expected: '%s'", storedContent, docContent)
	}
}
