package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"testing"
)

func TestWorkspaceSymbol(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
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
	if err := json.Unmarshal(mustMarshal(h, msg["result"]), &symbols); err != nil {
		t.Fatalf("Failed to unmarshal workspace symbols: %v", err)
	}

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
	assertNoUnhandledMessages(h, &logBuf)
}
