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

func TestDidOpenNotification(t *testing.T) {
	// Sample didOpen notification from a client
	message := `{
		"jsonrpc": "2.0",
		"method": "textDocument/didOpen",
		"params": {
			"textDocument": {
				"uri": "file:///path/to/test.grammar",
				"languageId": "grammar",
				"version": 1,
				"text": "rule Foo = \"bar\";"
			}
		}
	}`

	contentLength := len(message)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", contentLength)

	var in bytes.Buffer
	var out strings.Builder

	srv := server.NewServer(&in, &out)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		srv.Start()
		cancel()
	}()

	in.WriteString(header)
	in.WriteString(message)

	<-ctx.Done()

	response := out.String()

	if len(response) != 0 {
		t.Fatalf("Server produced an unexpected response for notification: %s", response)
	}

	// Additional check: verify the log output
	// Since notifications don't send responses, we can only check the logs for side effects.
	// This would require reading the log file, which is more complex for a unit test.
	// For now, confirming no LSP response is sufficient.
}
