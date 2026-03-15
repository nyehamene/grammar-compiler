package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"testing"
	"grammar/testutil"
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
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

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

	testutil.AssertSnapshotJSON(t, "document_symbol/basic_symbols", symbols)

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
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	id := 1
	var symbolParams any = server.DocumentSymbolParams{
		TextDocument: server.TextDocumentIdentifier{URI: uri},
	}
	h.send(newRequest(id, "textDocument/documentSymbol", &symbolParams))

	msg := h.read()
	assertResponseID(h, msg, id)

	var symbols []server.DocumentSymbol
	json.Unmarshal(mustMarshal(h, msg["result"]), &symbols)

	testutil.AssertSnapshotJSON(t, "document_symbol/package_symbols", symbols)

	if len(symbols) != 3 { // @package, rule_a, rule_b
		t.Fatalf("Expected 3 document symbols, got %d", len(symbols))
	}

	// Assert @package directive symbol
	pkgSymbol := symbols[0]
	if pkgSymbol.Name != "mypackage" {
		t.Errorf("Expected first symbol name to be 'mypackage', got %s", pkgSymbol.Name)
	}
	if pkgSymbol.Kind != server.SymbolKindPackage { // Assuming SymbolKindPackage for @package
		t.Errorf("Expected first symbol kind to be Package, got %d", pkgSymbol.Kind)
	}

	// Assert rule_a symbol
	ruleASymbol := symbols[1]
	if ruleASymbol.Name != "rule_a" {
		t.Errorf("Expected second symbol name to be 'rule_a', got %s", ruleASymbol.Name)
	}
	if ruleASymbol.Kind != server.SymbolKindField { // Assuming SymbolKindField for rules
		t.Errorf("Expected second symbol kind to be Field, got %d", ruleASymbol.Kind)
	}

	// Assert rule_b symbol
	ruleBSymbol := symbols[2]
	if ruleBSymbol.Name != "rule_b" {
		t.Errorf("Expected third symbol name to be 'rule_b', got %s", ruleBSymbol.Name)
	}
	if ruleBSymbol.Kind != server.SymbolKindField { // Assuming SymbolKindField for rules
		t.Errorf("Expected third symbol kind to be Field, got %d", ruleBSymbol.Kind)
	}
	assertNoUnhandledMessages(h, &logBuf)
}
