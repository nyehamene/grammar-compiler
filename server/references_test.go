package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"testing"
)

func TestReferences(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
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
	if err := json.Unmarshal(mustMarshal(h, msg["result"]), &locations); err != nil {
		t.Fatalf("Failed to unmarshal locations: %v", err)
	}

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
	if err := json.Unmarshal(mustMarshal(h, msgFromUsage["result"]), &locationsFromUsage); err != nil {
		t.Fatalf("Failed to unmarshal locations from usage: %v", err)
	}

	if len(locationsFromUsage) != 3 {
		t.Fatalf("Expected 3 references from usage, got %d", len(locationsFromUsage))
	}
	assertNoUnhandledMessages(h, &logBuf)
}
