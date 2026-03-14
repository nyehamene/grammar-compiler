package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"testing"
)

func TestDocumentSymbol(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
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
	assertNoUnhandledMessages(h, &logBuf)
}

func TestDocumentSymbolPackage(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	testDir := t.TempDir()
	pkgDir := testDir + "/pkg"
	os.MkdirAll(pkgDir, 0755)

	content := `@package("mypackage");
rule_a = "a";
rule_b = "b";`
	modulePath := pkgDir + "/module.grammar"
	os.WriteFile(modulePath, []byte(content), 0644)

	uri, _ := server.ParseURI("file://" + modulePath)
	h.send(newDidOpenNotification(uri, content, 1))
	consumeDiagnostics(h)

	id := 1
	var symbolParams any = server.DocumentSymbolParams{
		TextDocument: server.TextDocumentIdentifier{URI: uri},
	}
	h.send(newRequest(id, "textDocument/documentSymbol", &symbolParams))

	msg := h.read()
	assertResponseID(h, msg, id)

	var symbols []server.DocumentSymbol
	json.Unmarshal(mustMarshal(h, msg["result"]), &symbols)

	t.Logf("Found %d symbols", len(symbols))
	for _, s := range symbols {
		t.Logf("Symbol: %s, Kind: %d", s.Name, s.Kind)
	}
	assertNoUnhandledMessages(h, &logBuf)
}
