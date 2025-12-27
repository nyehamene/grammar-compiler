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

func TestHover(t *testing.T) {
	h := setupTestServer(t)
	defer h.clientConn.Close()

	// 1. Open a document that imports another.
	bContent := "rule_b = \"from b\";"
	bURI, _ := server.ParseURI("file:///b.grammar")
	bDidOpenParams := server.DidOpenTextDocumentParams{
		TextDocument: server.TextDocumentItem{
			URI: bURI, Text: bContent, Version: 1,
		},
	}
	var bParams any = bDidOpenParams
	h.send(server.NotificationMessage{
		Message: server.Message{JSONRPC: "2.0"},
		Method:  "textDocument/didOpen",
		Params:  &bParams,
	})
	diags := h.read() // Consume diagnostics
	diagsParams := diags["params"].(map[string]any)
	diag := diagsParams["diagnostics"].([]any)
	if len(diag) != 0 {
		t.Fatalf("server reported an error: %s", diag)
	}

	aContent := "b = @import(\"b.grammar\");\na = b.rule_b;"
	aURI, _ := server.ParseURI("file:///a.grammar")
	aDidOpenParams := server.DidOpenTextDocumentParams{
		TextDocument: server.TextDocumentItem{
			URI: aURI, Text: aContent, Version: 1,
		},
	}
	var aParams any = aDidOpenParams
	h.send(server.NotificationMessage{
		Message: server.Message{JSONRPC: "2.0"},
		Method:  "textDocument/didOpen",
		Params:  &aParams,
	})
	diags = h.read() // Consume diagnostics
	diagsParams = diags["params"].(map[string]any)
	diag = diagsParams["diagnostics"].([]any)
	if len(diag) != 0 {
		t.Fatalf("server reported an error: %s", diag)
	}

	// 2. Send hover request
	id := 1
	var hoverReqParams any = server.HoverParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: aURI},
			Position:     server.Position{Line: 1, Character: 5}, // Position of 'b.rule_b'
		},
	}
	h.send(server.RequestMessage{
		Message: server.Message{JSONRPC: "2.0"},
		ID:      &id,
		Method:  "textDocument/hover",
		Params:  &hoverReqParams,
	})

	// 3. Read and verify response
	msg := h.read()
	if msg["id"] == nil || msg["id"].(float64) != float64(id) {
		t.Fatalf("Expected response for request %d, got %v", id, msg)
	}

	resultData, _ := json.Marshal(msg["result"])
	var hover server.Hover
	json.Unmarshal(resultData, &hover)

	if hover.Contents.Kind != server.MarkupKindPlainText {
		t.Errorf("Expected hover kind to be plaintext, got %s", hover.Contents.Kind)
	}
	if hover.Contents.Value != "production" {
		t.Errorf("Expected hover value to be 'production', got %s", hover.Contents.Value)
	}
}

func TestDefinition(t *testing.T) {
	h := setupTestServer(t)
	defer h.clientConn.Close()

	// 1. Open documents
	bContent := "rule_b = \"from b\";"
	bURI, _ := server.ParseURI("file:///b.grammar")
	h.send(newDidOpenNotification(bURI, bContent, 1))
	consumeDiagnostics(h)

	aContent := "b = @import(\"b.grammar\");\nlocal_rule = b.rule_b;\nfinal_rule = local_rule;"
	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, aContent, 1))
	consumeDiagnostics(h)

	// 2. Test cross-file definition
	id := 1
	var definitionParams any = server.DefinitionParams{
		TextDocument: server.TextDocumentIdentifier{URI: aURI},
		Position:     server.Position{Line: 1, Character: 15}, // on 'rule_b'
	}
	h.send(newRequest(id, "textDocument/definition", &definitionParams))

	msg := h.read()
	assertResponseID(h, msg, id)

	var location server.Location
	json.Unmarshal(mustMarshal(h, msg["result"]), &location)

	if location.URI != bURI {
		t.Errorf("Expected URI %s, got %s", bURI, location.URI)
	}
	if location.Range.Start.Line != 0 || location.Range.Start.Character != 0 {
		t.Errorf("Expected range start at 0:0, got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
	}

	// 3. Test same-file definition
	id = 2
	var localDefinitionParams any = server.DefinitionParams{
		TextDocument: server.TextDocumentIdentifier{URI: aURI},
		Position:     server.Position{Line: 2, Character: 15}, // on 'local_rule'
	}
	h.send(newRequest(id, "textDocument/definition", &localDefinitionParams))

	msg = h.read()
	assertResponseID(h, msg, id)
	json.Unmarshal(mustMarshal(h, msg["result"]), &location)

	if location.URI != aURI {
		t.Errorf("Expected URI %s, got %s", aURI, location.URI)
	}
	if location.Range.Start.Line != 1 || location.Range.Start.Character != 0 {
		t.Errorf("Expected range start at 1:0, got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
	}
}

func TestReferences(t *testing.T) {
	h := setupTestServer(t)
	defer h.clientConn.Close()

	// 1. Open documents
	commonContent := "export_rule = \"hello\";"
	commonURI, _ := server.ParseURI("file:///common.grammar")
	h.send(newDidOpenNotification(commonURI, commonContent, 1))
	consumeDiagnostics(h)

	userAContent := "common = @import(\"common.grammar\");\na_rule = common.export_rule;"
	userAURI, _ := server.ParseURI("file:///user_a.grammar")
	h.send(newDidOpenNotification(userAURI, userAContent, 1))
	consumeDiagnostics(h)

	userBContent := "common = @import(\"common.grammar\");\nb_rule = common.export_rule;"
	userBURI, _ := server.ParseURI("file:///user_b.grammar")
	h.send(newDidOpenNotification(userBURI, userBContent, 1))
	consumeDiagnostics(h)

	// 2. Test references from the definition
	id := 1
	var refParams any = server.ReferenceParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: commonURI},
			Position:     server.Position{Line: 0, Character: 1}, // on 'export_rule'
		},
		Context: server.ReferenceContext{IncludeDeclaration: true},
	}
	h.send(newRequest(id, "textDocument/references", &refParams))

	msg := h.read()
	assertResponseID(h, msg, id)

	var locations []server.Location
	json.Unmarshal(mustMarshal(h, msg["result"]), &locations)

	if len(locations) != 3 {
		t.Fatalf("Expected 3 references, got %d", len(locations))
	}

	// 3. Test references from a usage
	id = 2
	var refParamsFromUsage any = server.ReferenceParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: userAURI},
			Position:     server.Position{Line: 1, Character: 18}, // on 'export_rule' in user_a
		},
		Context: server.ReferenceContext{IncludeDeclaration: true},
	}
	h.send(newRequest(id, "textDocument/references", &refParamsFromUsage))

	msgFromUsage := h.read()
	assertResponseID(h, msgFromUsage, id)
	var locationsFromUsage []server.Location
	json.Unmarshal(mustMarshal(h, msgFromUsage["result"]), &locationsFromUsage)

	if len(locationsFromUsage) != 3 {
		t.Fatalf("Expected 3 references from usage, got %d", len(locationsFromUsage))
	}
}

// --- Test Helpers ---

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
