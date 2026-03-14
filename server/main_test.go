package server_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"grammar/log"
	"grammar/server"
	"io"
	"strings"
	"testing"
)

// testLogger implements server.Logger for testing
type testLogger struct {
	out io.Writer
}

func newTestLogger(out io.Writer) *testLogger {
	return &testLogger{out: out}
}

func (l *testLogger) Print(v any) {
	if l.out == nil {
		return
	}
	fmt.Fprintln(l.out, v)
}

func (l *testLogger) Printf(format string, v ...any) {
	if l.out == nil {
		return
	}
	_, _ = fmt.Fprintf(l.out, format+"\n", v...)
}

// lspTestHarness holds the client and server for an integration test.
type lspTestHarness struct {
	clientConn io.ReadWriteCloser
	server     *server.Server
	t          *testing.T
}

// setupTestServer creates a new server with an in-memory client connection.
func setupTestServer(t *testing.T, logOut io.Writer) *lspTestHarness {
	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()

	logger := log.NewConsoleLogger(logOut, log.DEBUG)

	serv := server.NewServer(serverRead, serverWrite, logger)

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

// assertNoUnhandledMessages checks the log output for any unhandled log type messages.
func assertNoUnhandledMessages(h *lspTestHarness, logBuf *bytes.Buffer) {
	logContent := logBuf.String()
	if strings.Contains(logContent, "unhandled log type") || strings.Contains(logContent, "unhandled raw message") {
		h.t.Errorf("Found unhandled log messages in server output:\n%s", logContent)
	}
}

func newDidOpenNotification(uri server.DocumentUri, content string, version int) server.NotificationMessage {
	params := server.DidOpenTextDocumentParams{
		TextDocument: server.TextDocumentItem{
			URI: uri, Text: content, Version: version, LanguageID: "grammar",
		},
	}
	var p any = params
	return server.NotificationMessage{
		Message: server.Message{JSONRPC: "2.0"},
		Method:  "textDocument/didOpen",
		Params:  &p,
	}
}

// newDidChangeNotification creates a didChange notification.
func newDidChangeNotification(uri server.DocumentUri, content string, version int) server.NotificationMessage {
	params := server.DidChangeTextDocumentParams{
		TextDocument: server.VersionedTextDocumentIdentifier{
			URI:     uri, // Correctly assign URI
			Version: version,
		},
		ContentChanges: []server.TextDocumentContentChangeEvent{
			{
				Text: content,
			},
		},
	}
	var p any = params
	return server.NotificationMessage{
		Message: server.Message{JSONRPC: "2.0"},
		Method:  "textDocument/didChange",
		Params:  &p,
	}
}

func newRequest(id int, method string, params *any) server.RequestMessage {
	return server.RequestMessage{
		Message: server.Message{JSONRPC: "2.0"},
		ID:      &id,
		Method:  method,
		Params:  params,
	}
}

func consumeDiagnostics(h *lspTestHarness) {
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		h.t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}
	params, _ := msg["params"].(map[string]any)
	diags, _ := params["diagnostics"].([]any)
	if len(diags) > 0 {
		h.t.Logf("Consumed %d diagnostics", len(diags))
	}
}

func assertResponseID(h *lspTestHarness, msg map[string]any, expectedID int) {
	if msg["id"] == nil {
		h.t.Fatalf("Response is missing ID, expected %d. Got: %v", expectedID, msg)
	}
	if int(msg["id"].(float64)) != expectedID {
		h.t.Fatalf("Expected response for request %d, got ID %v", expectedID, msg["id"])
	}
}

func mustMarshal(h *lspTestHarness, v any) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		h.t.Fatalf("Failed to marshal: %v", err)
	}
	return bytes
}

func newInitializeRequest(id int, capabilities server.ClientCapabilities) server.RequestMessage {
	params := server.InitializeParams{
		Capabilities: capabilities,
	}
	var p any = params
	return newRequest(id, "initialize", &p)
}
