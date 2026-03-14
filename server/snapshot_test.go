package server_test

import (
	"bytes"
	"grammar/server"
	"grammar/testutil"
	"os"
	"testing"
)

func createTestFiles(t *testing.T, files map[string]string) (map[string]server.DocumentUri, func()) {
	uris := make(map[string]server.DocumentUri)

	for name := range files {
		uri, _ := server.ParseURI("file:///test_project/" + name)
		uris[name] = uri
	}

	return uris, func() {}
}

func TestCompletionSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `
b = @import("b.grammar");
prod_a = "a";
prod_b = b.
`,
		"b.grammar": `
rule_b = "from b";
rule_c = "from c";
`,
	}

	uris, _ := createTestFiles(t, files)

	h.send(newDidOpenNotification(uris["b.grammar"], files["b.grammar"], 1))
	consumeDiagnostics(h)

	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 2))
	consumeDiagnostics(h)

	id := 1
	var completionParams any = server.CompletionParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: uris["a.grammar"]},
			Position:     server.Position{Line: 2, Character: 11},
		},
	}

	h.send(newRequest(id, "textDocument/completion", &completionParams))
	msg := h.read()
	assertResponseID(h, msg, id)

	result := msg["result"]
	testutil.AssertSnapshotJSON(t, "completion_basic", result)
}

func TestHoverSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `
str = "hello";
prod_a = "a";
`,
	}

	uris, _ := createTestFiles(t, files)
	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 1))
	consumeDiagnostics(h)

	id := 1
	var hoverParams any = server.HoverParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: uris["a.grammar"]},
			Position:     server.Position{Line: 1, Character: 7},
		},
	}

	h.send(newRequest(id, "textDocument/hover", &hoverParams))
	msg := h.read()
	assertResponseID(h, msg, id)

	result := msg["result"]
	testutil.AssertSnapshotJSON(t, "hover_string_literal", result)
}

func TestDefinitionSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `
prod_a = "a";
prod_b = prod_a;
`,
	}

	uris, _ := createTestFiles(t, files)
	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 1))
	consumeDiagnostics(h)

	id := 1
	var defParams any = server.DefinitionParams{
		TextDocument: server.TextDocumentIdentifier{URI: uris["a.grammar"]},
		Position:     server.Position{Line: 2, Character: 11},
	}

	h.send(newRequest(id, "textDocument/definition", &defParams))
	msg := h.read()
	assertResponseID(h, msg, id)

	result := msg["result"]
	testutil.AssertSnapshotJSON(t, "definition_local", result)
}

func TestReferencesSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `
prod_a = "a";
prod_b = prod_a;
prod_c = prod_a;
`,
	}

	uris, _ := createTestFiles(t, files)
	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 1))
	consumeDiagnostics(h)

	id := 1
	var refParams any = server.ReferenceParams{
		TextDocumentPositionParams: server.TextDocumentPositionParams{
			TextDocument: server.TextDocumentIdentifier{URI: uris["a.grammar"]},
			Position:     server.Position{Line: 2, Character: 11},
		},
		Context: server.ReferenceContext{IncludeDeclaration: true},
	}

	h.send(newRequest(id, "textDocument/references", &refParams))
	msg := h.read()
	assertResponseID(h, msg, id)

	result := msg["result"]
	testutil.AssertSnapshotJSON(t, "references_local", result)
}

func TestDiagnosticsSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `
prod_a = "a";
prod_b = undefined_rule;
`,
	}

	uris, _ := createTestFiles(t, files)
	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 1))

	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	result := msg["params"]
	testutil.AssertSnapshotJSON(t, "diagnostics_error", result)
}

func TestDocumentSymbolsSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `
prod_a = "a";
prod_b = "b";
prod_c = prod_a | prod_b;
`,
	}

	uris, _ := createTestFiles(t, files)
	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 1))
	consumeDiagnostics(h)

	id := 1
	var symParams any = server.DocumentSymbolParams{
		TextDocument: server.TextDocumentIdentifier{URI: uris["a.grammar"]},
	}

	h.send(newRequest(id, "textDocument/documentSymbol", &symParams))
	msg := h.read()
	assertResponseID(h, msg, id)

	result := msg["result"]
	testutil.AssertSnapshotJSON(t, "document_symbols_basic", result)
}

func TestFormattingSnapshot(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()

	files := map[string]string{
		"a.grammar": `prod_a = "a";prod_b =    "b";`,
	}

	uris, _ := createTestFiles(t, files)
	h.send(newDidOpenNotification(uris["a.grammar"], files["a.grammar"], 1))
	consumeDiagnostics(h)

	id := 1
	var formatParams any = server.DocumentFormattingParams{
		TextDocument: server.TextDocumentIdentifier{URI: uris["a.grammar"]},
		Options:      server.FormattingOptions{TabSize: 2, InsertSpaces: true},
	}

	h.send(newRequest(id, "textDocument/formatting", &formatParams))
	msg := h.read()
	assertResponseID(h, msg, id)

	result := msg["result"]
	testutil.AssertSnapshotJSON(t, "formatting_basic", result)
}

func init() {
	if os.Getenv("UPDATE_SNAPSHOTS") == "true" {
		os.MkdirAll("testdata/snapshots/completion", 0755)
		os.MkdirAll("testdata/snapshots/hover", 0755)
		os.MkdirAll("testdata/snapshots/definition", 0755)
		os.MkdirAll("testdata/snapshots/references", 0755)
		os.MkdirAll("testdata/snapshots/diagnostics", 0755)
		os.MkdirAll("testdata/snapshots/documentSymbol", 0755)
		os.MkdirAll("testdata/snapshots/formatter", 0755)
	}
}
