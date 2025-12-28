package server_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"grammar/server"
	"io"
	"os"
	"path/filepath"
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
func setupTestServer(t *testing.T, logOut io.Writer) *lspTestHarness {
	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()

	serv := server.NewServerWithLogger(serverRead, serverWrite, logOut)

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
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

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

func TestImportedNamespaceLoading(t *testing.T) {
	// Create test files on disk
	testDir := t.TempDir()
	bPath := filepath.Join(testDir, "b.grammar")
	aPath := filepath.Join(testDir, "a.grammar")

	bContent := `rule_b = "from b";`
	if err := os.WriteFile(bPath, []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.grammar: %v", err)
	}

	aContent := fmt.Sprintf(`b = @import("b.grammar"); a = b.rule_b;`) // relative import
	if err := os.WriteFile(aPath, []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.grammar: %v", err)
	}

	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	aURI, _ := server.ParseURI("file://" + aPath)

	// Open only a.grammar. The server should load b.grammar from filesystem.
	h.send(newDidOpenNotification(aURI, aContent, 1))

	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}
	params, _ := msg["params"].(map[string]any)
	diags, _ := params["diagnostics"].([]any)

	if len(diags) != 0 {
		t.Errorf("Expected no diagnostics, but got %d: %v", len(diags), diags)
	}
}

func TestHover(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// Setup for cross-file tests
	bContent := "rule_b = \"from b\";"
	bURI, _ := server.ParseURI("file:///b.grammar")
	h.send(newDidOpenNotification(bURI, bContent, 1))
	consumeDiagnostics(h)

	aContent := `
b = @import("b.grammar");
str = "hello";
re = /[a-z]/;
ext = $foo;
prod_a = "a";
prod_b = prod_a;
prod_c = b.rule_b;
`
	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, aContent, 2))
	consumeDiagnostics(h)

	testCases := []struct {
		name          string
		uri           server.DocumentUri
		position      server.Position
		expectedType  string
		expectedValue string
	}{
		{"string literal", aURI, server.Position{Line: 2, Character: 7}, "string", `"hello"`},
		{"regexp literal", aURI, server.Position{Line: 3, Character: 6}, "regexp", `/[a-z]/`},
		{"external value", aURI, server.Position{Line: 4, Character: 7}, "external", `$foo`},
		{"local production", aURI, server.Position{Line: 6, Character: 9}, "production", `"a";`},
		{"binding", aURI, server.Position{Line: 1, Character: 0}, "namespace", "b.grammar"},
		{"imported member", aURI, server.Position{Line: 7, Character: 11}, "production", `"from b";`},
	}

	idCounter := 1
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := idCounter
			var hoverReqParams any = server.HoverParams{
				TextDocumentPositionParams: server.TextDocumentPositionParams{
					TextDocument: server.TextDocumentIdentifier{URI: tc.uri},
					Position:     tc.position,
				},
			}
			h.send(newRequest(id, "textDocument/hover", &hoverReqParams))

			msg := h.read()
			assertResponseID(h, msg, id)

			resultData, err := json.Marshal(msg["result"])
			if err != nil {
				t.Fatalf("Failed to marshal hover result: %v", err)
			}
			var hover server.Hover
			json.Unmarshal(resultData, &hover)

			if hover.Contents.Kind != server.MarkupKindMarkdown {
				t.Errorf("Expected hover kind to be Markdown, got %s", hover.Contents.Kind)
			}

			expectedHoverValue := fmt.Sprintf("(%s)\n\n```grammar\n%s\n```\n", tc.expectedType, tc.expectedValue)
			if hover.Contents.Value != expectedHoverValue {
				t.Errorf("Expected hover value:\n%s\nGot:\n%s", expectedHoverValue, hover.Contents.Value)
			}

			idCounter++
		})
	}
}

func TestDefinition(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

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
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

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

func TestDocumentSymbol(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// 1. Open document
	content := `
// A binding
b = @import("b.grammar");

// A rule
rule_a = "hello";
`
	uri, _ := server.ParseURI("file:///test.grammar")
	h.send(newDidOpenNotification(uri, content, 1))
	consumeDiagnostics(h)

	// 2. Send request
	id := 1
	var symbolParams any = server.DocumentSymbolParams{
		TextDocument: server.TextDocumentIdentifier{URI: uri},
	}
	h.send(newRequest(id, "textDocument/documentSymbol", &symbolParams))

	// 3. Read and verify response
	msg := h.read()
	assertResponseID(h, msg, id)

	var symbols []server.DocumentSymbol
	json.Unmarshal(mustMarshal(h, msg["result"]), &symbols)

	if len(symbols) != 2 {
		t.Fatalf("Expected 2 document symbols, got %d", len(symbols))
	}

	bindingSymbol := symbols[0]
	if bindingSymbol.Name != "b" {
		t.Errorf("Expected first symbol name to be 'b', got %s", bindingSymbol.Name)
	}
	if bindingSymbol.Kind != server.SymbolKindVariable {
		t.Errorf("Expected first symbol kind to be Variable, got %d", bindingSymbol.Kind)
	}

	ruleSymbol := symbols[1]
	if ruleSymbol.Name != "rule_a" {
		t.Errorf("Expected second symbol name to be 'rule_a', got %s", ruleSymbol.Name)
	}
	if ruleSymbol.Kind != server.SymbolKindField {
		t.Errorf("Expected second symbol kind to be Function, got %d", ruleSymbol.Kind)
	}
	if ruleSymbol.SelectionRange.Start.Line != 5 { // Line numbers are 0-indexed
		t.Errorf("Expected rule selection range to start on line 5, got %d", ruleSymbol.SelectionRange.Start.Line)
	}
}

func TestWorkspaceSymbol(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// 1. Open documents
	contentA := "rule_foo = \"a\";"
	uriA, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(uriA, contentA, 1))
	consumeDiagnostics(h)

	contentB := "rule_bar = \"b\";\nbinding_foo = @import(\"a.grammar\");"
	uriB, _ := server.ParseURI("file:///b.grammar")
	h.send(newDidOpenNotification(uriB, contentB, 1))
	consumeDiagnostics(h)

	// 2. Send request
	id := 1
	var symbolParams any = server.WorkspaceSymbolParams{
		Query: "foo",
	}
	h.send(newRequest(id, "workspace/symbol", &symbolParams))

	// 3. Read and verify response
	msg := h.read()
	assertResponseID(h, msg, id)

	var symbols []server.SymbolInformation
	json.Unmarshal(mustMarshal(h, msg["result"]), &symbols)

	if len(symbols) != 2 {
		t.Fatalf("Expected 2 workspace symbols for query 'foo', got %d", len(symbols))
	}

	foundA := false
	foundB := false
	for _, s := range symbols {
		if s.Name == "rule_foo" && s.Location.URI == uriA {
			foundA = true
		}
		if s.Name == "binding_foo" && s.Location.URI == uriB {
			foundB = true
		}
	}

	if !foundA || !foundB {
		t.Errorf("Did not find expected symbols. Found A: %t, Found B: %t", foundA, foundB)
	}
}

func TestRename(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer h.clientConn.Close()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// 1. Open documents
	commonContent := "rename_rule = \"hello\";"
	commonURI, _ := server.ParseURI("file:///common.grammar")
	h.send(newDidOpenNotification(commonURI, commonContent, 1))
	consumeDiagnostics(h)

	userContent := "common = @import(\"common.grammar\");\nlocal = common.rename_rule;"
	userURI, _ := server.ParseURI("file:///user.grammar")
	h.send(newDidOpenNotification(userURI, userContent, 1))
	consumeDiagnostics(h)

	// 2. Test prepareRename
	id := 1
	var prepareParams any = server.PrepareRenameParams{
		TextDocument: server.TextDocumentIdentifier{URI: userURI},
		Position:     server.Position{Line: 1, Character: 18}, // on 'rename_rule'
	}
	h.send(newRequest(id, "textDocument/prepareRename", &prepareParams))
	msg := h.read()
	assertResponseID(h, msg, id)
	var prepareRange server.Range
	json.Unmarshal(mustMarshal(h, msg["result"]), &prepareRange)
	if prepareRange.Start.Character != 15 || prepareRange.End.Character != 26 {
		t.Fatalf("Expected prepareRename range to be 15-26, got %d-%d", prepareRange.Start.Character, prepareRange.End.Character)
	}

	// 3. Test rename
	id = 2
	newName := "renamed_rule"
	var renameParams any = server.RenameParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: userURI},
			Position:     server.Position{Line: 1, Character: 18}, // on 'rename_rule'
		},
		NewName: newName,
	}
	h.send(newRequest(id, "textDocument/rename", &renameParams))
	msg = h.read()
	assertResponseID(h, msg, id)
	var workspaceEdit server.WorkspaceEdit
	json.Unmarshal(mustMarshal(h, msg["result"]), &workspaceEdit)

	if len(workspaceEdit.Changes) != 2 {
		t.Fatalf("Expected edits in 2 files, got %d", len(workspaceEdit.Changes))
	}
	commonEdits, ok := workspaceEdit.Changes[commonURI.String()]
	if !ok || len(commonEdits) != 1 {
		t.Fatalf("Expected 1 edit in common.grammar")
	}
	if commonEdits[0].NewText != newName {
		t.Errorf("Wrong new name in common.grammar: got %s, want %s", commonEdits[0].NewText, newName)
	}

	userEdits, ok := workspaceEdit.Changes[userURI.String()]
	if !ok || len(userEdits) != 1 {
		t.Fatalf("Expected 1 edit in user.grammar")
	}
	if userEdits[0].NewText != newName {
		t.Errorf("Wrong new name in user.grammar: got %s, want %s", userEdits[0].NewText, newName)
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
