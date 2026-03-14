package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"grammar/server"
	"strings"
	"testing"
	"time"
)

func TestInitializeRequest(t *testing.T) {
	// Sample initialize request from a client
	message := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"processId": 1234,
			"rootUri": "file:///path/to/workspace",
			"capabilities": {}
		}
	}`

	contentLength := len(message)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", contentLength)

	var in bytes.Buffer
	var out strings.Builder
	var logOut bytes.Buffer
	defer func() {
		if t.Failed() {
			t.Log(logOut.String())
		}
	}()

	srv := server.NewServer(&in, &out, server.NewWriterLogger(&logOut))

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

	if len(response) == 0 {
		t.Fatal("Server did not produce a response")
	}

	// Parse the response to get the JSON content
	parts := strings.SplitN(response, "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatalf("Invalid response format: %s", response)
	}

	var resp server.ResponseMessage
	if err := json.Unmarshal([]byte(parts[1]), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got: %v", resp.Error)
	}
}
