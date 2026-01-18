package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"slices"
	"testing"
)

func TestCompletion(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)

	defer func() { _ = h.clientConn.Close() }()

	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// Setup for cross-file tests
	bContent := `
rule_b = "from b";
rule_c = "from c";

`

	bURI, _ := server.ParseURI("file:///b.grammar")
	h.send(newDidOpenNotification(bURI, bContent, 1))
	consumeDiagnostics(h)

	// Initial content for a.grammar (valid state)
	initialAContent := `
b = @import("b.grammar");
prod_a = "a";
prod_b = ""; // initially valid

`

	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, initialAContent, 1))
	consumeDiagnostics(h)
	idCounter := 1 // Global ID counter for all sub-tests

	t.Run("member completion", func(t *testing.T) {
		// Simulate typing "b." after "prod_b = "
		// Document version is 2
		incompleteAContent := `
b = @import("b.grammar");
prod_a = "a";
prod_b = b.
`

		h.send(newDidChangeNotification(aURI, incompleteAContent, 2)) // version 2
		consumeDiagnostics(h)                                         // Consume diagnostics for the incomplete state
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: aURI},
				Position:     server.Position{Line: 3, Character: 11}, // after 'b.'
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg := h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		if len(completionList.Items) != 2 {
			t.Fatalf("Expected 2 completion items, got %d", len(completionList.Items))
		}

		expectedLabels := []string{"rule_b", "rule_c"}
		for _, item := range completionList.Items {
			found := slices.Contains(expectedLabels, item.Label)
			if !found {
				t.Errorf("Unexpected completion item: %s", item.Label)
			}
		}
		idCounter++
	})

	t.Run("rule body completion", func(t *testing.T) {
		// Restore a valid state (or a different valid state for this test)
		// Document version is 3
		validAContent := `

b = @import("b.grammar");

prod_a = "a";

prod_b = ;

`

		h.send(newDidChangeNotification(aURI, validAContent, 3)) // version 3
		consumeDiagnostics(h)
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: aURI},
				Position:     server.Position{Line: 4, Character: 9}, // after '='
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg := h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])

		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList

		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)

		}

		// Expect 'b' and 'prod_a'

		if len(completionList.Items) != 2 {
			t.Fatalf("Expected 2 completion items, got %d", len(completionList.Items))
		}

		expectedLabels := []string{"b", "prod_a"}

		for _, item := range completionList.Items {
			found := slices.Contains(expectedLabels, item.Label)
			if !found {
				t.Errorf("Unexpected completion item: %s", item.Label)
			}
		}

		idCounter++
	})
	assertNoUnhandledMessages(h, &logBuf)
}
