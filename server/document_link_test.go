package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentLinkResolveRequest(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// Create test files on disk
	testDir := t.TempDir()
	bPath := filepath.Join(testDir, "b.grammar")
	aPath := filepath.Join(testDir, "a.grammar")

	bContent := `rule_b = "from b";`
	if err := os.WriteFile(bPath, []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.grammar: %v", err)
	}

	aContent := `binding_b = @import("b.grammar");`
	if err := os.WriteFile(aPath, []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.grammar: %v", err)
	}

	aURI, _ := server.ParseURI("file://" + aPath)
	bURI, _ := server.ParseURI("file://" + bPath)

	// 1. Open a.grammar
	h.send(newDidOpenNotification(aURI, aContent, 1))
	consumeDiagnostics(h) // Consume initial diagnostics for a.grammar

	// 2. Send documentLink request for a.grammar
	id := 1
	var linkParams any = server.DocumentLinkParams{
		TextDocument: server.TextDocumentIdentifier{URI: aURI},
	}
	h.send(newRequest(id, "textDocument/documentLink", &linkParams))

	// 3. Read and verify the unresolved response
	msg := h.read()
	assertResponseID(h, msg, id)

	resultData, err := json.Marshal(msg["result"])
	if err != nil {
		t.Fatalf("Failed to marshal documentLink result: %v", err)
	}

	var links []server.DocumentLink
	if err := json.Unmarshal(resultData, &links); err != nil {
		t.Fatalf("Failed to unmarshal document links: %v", err)
	}

	if len(links) != 1 {
		t.Fatalf("Expected 1 document link, got %d", len(links))
	}

	unresolvedLink := links[0]
	if unresolvedLink.Target != (server.DocumentUri{}) {
		t.Errorf("Expected unresolved link target to be empty, got %s", unresolvedLink.Target)
	}
	if unresolvedLink.Data == nil {
		t.Fatal("Expected unresolved link to have data, but it was nil")
	}

	// 4. Send documentLink/resolve request
	id = 2
	var resolveParams any = unresolvedLink
	h.send(newRequest(id, "documentLink/resolve", &resolveParams))

	// 5. Read and verify the resolved response
	msg = h.read()
	assertResponseID(h, msg, id)

	resultData, err = json.Marshal(msg["result"])
	if err != nil {
		t.Fatalf("Failed to marshal resolved documentLink result: %v", err)
	}

	var resolvedLink server.DocumentLink
	if err := json.Unmarshal(resultData, &resolvedLink); err != nil {
		t.Fatalf("Failed to unmarshal resolved document link: %v", err)
	}

	if resolvedLink.Target != bURI {
		t.Errorf("Expected resolved link target %v, got %v", bURI, resolvedLink.Target)
	}
	if resolvedLink.Tooltip != bURI.String() {
		t.Errorf("Expected resolved link tooltip %v, got %v", bURI.String(), resolvedLink.Tooltip)
	}
	assertNoUnhandledMessages(h, &logBuf)
}

func TestDocumentLinkPackage(t *testing.T) {
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

	t.Run("document link for package import", func(t *testing.T) {
		moduleContent := `@package("pkg");
rule_m = "m";`
		modulePath := pkgDir + "/module.grammar"
		os.WriteFile(modulePath, []byte(moduleContent), 0644)

		mainContent := `pkg = @import("pkg");`
		mainPath := testDir + "/main.grammar"
		os.WriteFile(mainPath, []byte(mainContent), 0644)

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		consumeDiagnostics(h)

		id := 1
		var linkParams any = server.DocumentLinkParams{
			TextDocument: server.TextDocumentIdentifier{URI: mainURI},
		}
		h.send(newRequest(id, "textDocument/documentLink", &linkParams))

		msg := h.read()
		assertResponseID(h, msg, id)

		resultData, _ := json.Marshal(msg["result"])
		var links []server.DocumentLink
		json.Unmarshal(resultData, &links)

		t.Logf("Found %d document links for package import", len(links))
		for _, l := range links {
			t.Logf("Link: target=%s, tooltip=%s", l.Target, l.Tooltip)
		}
	})

	t.Run("document link for deprecated file import", func(t *testing.T) {
		testDir2 := t.TempDir()
		oldContent := `old_rule = "old";`
		oldPath := testDir2 + "/old.grammar"
		os.WriteFile(oldPath, []byte(oldContent), 0644)

		mainContent := `old = @import("old.grammar");`
		mainPath := testDir2 + "/main.grammar"
		os.WriteFile(mainPath, []byte(mainContent), 0644)

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		consumeDiagnostics(h)

		id := 2
		var linkParams any = server.DocumentLinkParams{
			TextDocument: server.TextDocumentIdentifier{URI: mainURI},
		}
		h.send(newRequest(id, "textDocument/documentLink", &linkParams))

		msg := h.read()
		assertResponseID(h, msg, id)

		resultData, _ := json.Marshal(msg["result"])
		var links []server.DocumentLink
		json.Unmarshal(resultData, &links)

		t.Logf("Found %d document links for file import", len(links))
		for _, l := range links {
			t.Logf("Link: target=%s, tooltip=%s", l.Target, l.Tooltip)
		}
	})
}
