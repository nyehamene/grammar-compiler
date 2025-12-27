package server_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"grammar/server"
	"io"
	"strings"
	"testing"
	"time"
)

// lspTestHarness holds the client and server for an integration test.
type lspTestHarness struct {
	clientConn io.ReadWriteCloser
	server     *server.Server
	t          *testing.T
}

// setupTestServer creates a new server with an in-memory client connection.
func setupTestServer(t *testing.T) *lspTestHarness {
	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()

	serv := server.NewServer(serverRead, serverWrite)

	clientConn := &inMemoryConn{
		Reader: clientRead,
		Writer: clientWrite,
		Closer: serverWrite,
	}

	go serv.Start()

	return &lspTestHarness{
		clientConn: clientConn,
		server:     serv,
		t:          t,
	}
}

// inMemoryConn is a simple struct to combine a reader, writer, and closer.
type inMemoryConn struct {
	io.Reader
	io.Writer
	io.Closer
}

// send sends a message from the client to the server.
func (h *lspTestHarness) send(msg any) {
	encoded, err := json.Marshal(msg)
	if err != nil {
		h.t.Fatalf("Failed to marshal message: %v", err)
	}

	payload := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(encoded), encoded)
	_, err = h.clientConn.Write([]byte(payload))
	if err != nil {
		h.t.Fatalf("Failed to write to server: %v", err)
	}
}

// read reads a message from the server.
func (h *lspTestHarness) read() map[string]any {
	reader := bufio.NewReader(h.clientConn)
	var contentLen int64

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				h.t.Log("Server closed connection")
				return nil
			}
			h.t.Fatalf("Failed to read from server: %v", err)
		}
		if strings.HasPrefix(line, "Content-Length:") {
			_, err := fmt.Sscanf(line, "Content-Length: %d", &contentLen)
			if err != nil {
				h.t.Fatalf("Failed to parse Content-Length: %v", err)
			}
		}
		if line == "\r\n" {
			break
		}
	}

	if contentLen == 0 {
		h.t.Fatal("Received message with content length 0")
	}

	content := make([]byte, contentLen)
	_, err := io.ReadFull(reader, content)
	if err != nil {
		h.t.Fatalf("Failed to read message content: %v", err)
	}

	var msg map[string]any
	if err := json.Unmarshal(content, &msg); err != nil {
		h.t.Fatalf("Failed to unmarshal message from server: %v \nContent: %s", err, string(content))
	}
	return msg
}

func TestDidOpenPublishDiagnostics(t *testing.T) {
	h := setupTestServer(t)
	defer h.clientConn.Close()

	textDocumentId, err := server.ParseURI("file:///test.grammar")
	if err != nil {
		t.Fatal(err)
	}

	content := "a = b.c;"
	didOpenParams := server.DidOpenTextDocumentParams{
		TextDocument: server.TextDocumentItem{
			URI:        server.DocumentUri(textDocumentId),
			LanguageID: "grammar",
			Version:    1,
			Text:       content,
		},
	}
	var params any = didOpenParams
	didOpenNotif := server.NotificationMessage{
		Message: server.Message{JSONRPC: "2.0"},
		Method:  "textDocument/didOpen",
		Params:  &params,
	}

	h.send(didOpenNotif)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var msg map[string]any
	readChan := make(chan map[string]any)
	go func() {
		readChan <- h.read()
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for server response")
	case msg = <-readChan:
		if msg == nil {
			t.Fatal("Did not receive a message from the server")
		}
	}

	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics notification, but got: %s", msg["method"])
	}

	paramsData, _ := json.Marshal(msg["params"])
	var diagParams server.PublishDiagnosticsParams
	json.Unmarshal(paramsData, &diagParams)

	if diagParams.URI != textDocumentId {
		t.Errorf("Expected diagnostics for URI 'file:///test.grammar', but got: %s", diagParams.URI)
	}

	if len(diagParams.Diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, but got %d", len(diagParams.Diagnostics))
	}

	diag := diagParams.Diagnostics[0]
	if !strings.Contains(diag.Message, "undefined identifier: b") {
		t.Errorf("Expected diagnostic message to contain 'undefined identifier: b', but got: %s", diag.Message)
	}

	if diag.Range.Start.Line != 0 || diag.Range.Start.Character != 4 {
		t.Errorf("Expected diagnostic to start at line 0, char 4, but got line %d, char %d", diag.Range.Start.Line, diag.Range.Start.Character)
	}
}
