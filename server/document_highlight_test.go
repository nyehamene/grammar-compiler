package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"grammar/server"
	"testing"
)

func TestDocumentHighlight(t *testing.T) {
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
highlight_rule = "a";
ref1 = highlight_rule;
ref2 = highlight_rule;
`
	uri, _ := server.ParseURI("file:///highlight.grammar")
	h.send(newDidOpenNotification(uri, content, 1))
	consumeDiagnostics(h)

	// Positions to test (declaration, ref1, ref2)
	testPositions := []server.Position{
		{Line: 1, Character: 5}, // on 'highlight_rule' declaration
		{Line: 2, Character: 8}, // on 'highlight_rule' first reference
		{Line: 3, Character: 8}, // on 'highlight_rule' second reference
	}

	idCounter := 1
	for _, pos := range testPositions {
		t.Run(fmt.Sprintf("position_%d_%d", pos.Line, pos.Character), func(t *testing.T) {
			// 2. Send request
			id := idCounter
			var highlightParams any = server.DocumentHighlightParams{
				TextDocument: server.TextDocumentIdentifier{URI: uri},
				Position:     pos,
			}
			h.send(newRequest(id, "textDocument/documentHighlight", &highlightParams))

			// 3. Read and verify response
			msg := h.read()
			assertResponseID(h, msg, id)

			var highlights []server.DocumentHighlight
			if err := json.Unmarshal(mustMarshal(h, msg["result"]), &highlights); err != nil {
				t.Fatalf("Failed to unmarshal document highlights: %v", err)
			}

			if len(highlights) != 3 {
				t.Fatalf("Expected 3 document highlights, got %d", len(highlights))
			}

			// Check kinds
			writeFound := 0
			readFound := 0
			for _, h := range highlights {
				if h.Kind == nil {
					t.Error("Highlight kind is nil")
					continue
				}
				switch *h.Kind {
				case server.Write:
					writeFound++
				case server.Read:
					readFound++
				}
			}

			if writeFound != 1 {
				t.Errorf("Expected 1 'Write' highlight, found %d", writeFound)
			}
			if readFound != 2 {
				t.Errorf("Expected 2 'Read' highlights, found %d", readFound)
			}
		})
		idCounter++
	}
	assertNoUnhandledMessages(h, &logBuf)
}
