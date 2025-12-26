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

func TestDidCloseNotification(t *testing.T) {
	// First, simulate opening a document
	openMessage := `{
		"jsonrpc": "2.0",
		"method": "textDocument/didOpen",
		"params": {
			"textDocument": {
				"uri": "file:///path/to/close_test.grammar",
				"languageId": "grammar",
				"version": 1,
				"text": "rule Foo = \"bar\";"
			}
		}
	}`

	// Then, simulate closing the same document
	closeMessage := `{ 
		"jsonrpc": "2.0",
		"method": "textDocument/didClose",
		"params": {
			"textDocument": {
				"uri": "file:///path/to/close_test.grammar"
			}
		}
	}`

	var in bytes.Buffer
	var out strings.Builder

	srv := server.NewServer(&in, &out)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		srv.Start()
		cancel()
	}()

	// Send didOpen notification
	openContentLength := len(openMessage)
	openHeader := fmt.Sprintf("Content-Length: %d\r\n\r\n", openContentLength)
	in.WriteString(openHeader)
	in.WriteString(openMessage)

	// Give a little time for the server to process didOpen
	time.Sleep(100 * time.Millisecond)

	// Send didClose notification
	closeContentLength := len(closeMessage)
	closeHeader := fmt.Sprintf("Content-Length: %d\r\n\r\n", closeContentLength)
	in.WriteString(closeHeader)
	in.WriteString(closeMessage)

	<-ctx.Done()

	response := out.String()

	if len(response) != 0 {
		t.Fatalf("Server produced an unexpected response for notification: %s", response)
	}

	// Verify that the document is no longer in the server's documents map
	// This requires access to the server's internal state, which is not directly possible
	// from an external test. For now, rely on log messages or implicit behavior.
	// A better test would expose a method on the server to check for opened documents.
}
