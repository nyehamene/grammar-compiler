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
	// Sample didChange notification from a client (full sync)
	message := `{
		"jsonrpc": "2.0",
		"method": "textDocument/didChange",
		"params": {
			"textDocument": {
				"uri": "file:///path/to/test.grammar",
				"version": 2
			},
			"contentChanges": [
				{
					"text": "rule Foo = \"baz\";"
				}
			]
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
}
