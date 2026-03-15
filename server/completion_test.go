package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"slices"
	"testing"
	"grammar/testutil"
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
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	// Initial content for a.grammar (valid state)
	initialAContent := `
b = @import("b.grammar");
prod_a = "a";
prod_b = ""; // initially valid

`

	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, initialAContent, 1))
	msg = h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}
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
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: aURI},
				Position:     server.Position{Line: 3, Character: 11}, // after 'b.'
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/member_completion", completionList)

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
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: aURI},
				Position:     server.Position{Line: 4, Character: 9}, // after '='
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])

		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/rule_body_completion", completionList)

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

func TestCompletionPackage(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)

	defer func() { _ = h.clientConn.Close() }()

	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// Create test files on disk for package resolution
	testDir := t.TempDir()
	pkgDir := testDir + "/pkg"
	os.MkdirAll(pkgDir, 0755)

	// Create package modules
	m1Content := `@package("pkg");
rule_m1 = "from m1";`
	m1Path := pkgDir + "/module_a.grammar"
	if err := os.WriteFile(m1Path, []byte(m1Content), 0644); err != nil {
		t.Fatalf("Failed to write module_a.grammar: %v", err)
	}

	m2Content := `@package("pkg");
rule_m2 = "from m2";`
	m2Path := pkgDir + "/module_b.grammar"
	if err := os.WriteFile(m2Path, []byte(m2Content), 0644); err != nil {
		t.Fatalf("Failed to write module_b.grammar: %v", err)
	}

	m1URI, _ := server.ParseURI("file://" + m1Path)
	h.send(newDidOpenNotification(m1URI, m1Content, 1))
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	m2URI, _ := server.ParseURI("file://" + m2Path)
	h.send(newDidOpenNotification(m2URI, m2Content, 1))
	msg = h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	idCounter := 1

	t.Run("package module completion", func(t *testing.T) {
		mainContent := `pkg = @import("pkg");
result = pkg.`
		mainPath := testDir + "/main.grammar"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main.grammar: %v", err)
		}

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: mainURI},
				Position:     server.Position{Line: 1, Character: 13}, // after 'pkg.'
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/package_module_completion", completionList)

		// Should show module_a and module_b
		if len(completionList.Items) < 1 {
			t.Logf("Completion items: %v", completionList.Items)
		}
		idCounter++
	})

	t.Run("package member completion", func(t *testing.T) {
		mainContent := `pkg = @import("pkg");
result = pkg.module_a.`
		mainPath := testDir + "/main2.grammar"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main2.grammar: %v", err)
		}

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: mainURI},
				Position:     server.Position{Line: 1, Character: 26}, // after 'pkg.module_a.'
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/package_member_completion", completionList)

		// Should show rule_m1
		if len(completionList.Items) < 1 {
			t.Logf("Completion items: %v", completionList.Items)
		}
		idCounter++
	})

	t.Run("package directory completion", func(t *testing.T) {
		mainContent := `pkg = @import("`
		mainPath := testDir + "/main3.grammar"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main3.grammar: %v", err)
		}

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: mainURI},
				Position:     server.Position{Line: 0, Character: 12}, // inside @import("")
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/package_directory_completion", completionList)

		// Should suggest pkg directory
		t.Logf("Completion items: %v", completionList.Items)
		idCounter++
	})

	t.Run("same-package module completion", func(t *testing.T) {
		// Test @package("pkg"). to get module names in the same package
		mainContent := `@package("pkg").`
		mainPath := testDir + "/main5.grammar"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main5.grammar: %v", err)
		}

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: mainURI},
				Position:     server.Position{Line: 0, Character: 16}, // after '@package("pkg").'
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/same_package_module_completion", completionList)

		// Should suggest module_a and module_b
		t.Logf("Completion items: %v", completionList.Items)
		idCounter++
	})

	t.Run("completion in rule body with package", func(t *testing.T) {
		mainContent := `pkg = @import("pkg");

myrule = `
		mainPath := testDir + "/main4.grammar"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main4.grammar: %v", err)
		}

		mainURI, _ := server.ParseURI("file://" + mainPath)
		h.send(newDidOpenNotification(mainURI, mainContent, 1))
		msg = h.read()
		if msg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msg)
		}
		id := idCounter

		var completionParams any = server.CompletionParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: mainURI},
				Position:     server.Position{Line: 2, Character: 9}, // after '='
			},
		}

		h.send(newRequest(id, "textDocument/completion", &completionParams))
		msg = h.read()
		assertResponseID(h, msg, id)

		resultData, err := json.Marshal(msg["result"])
		if err != nil {
			t.Fatalf("Failed to marshal completion result: %v", err)
		}

		var completionList server.CompletionList
		if err := json.Unmarshal(resultData, &completionList); err != nil {
			t.Fatalf("Failed to unmarshal completion list: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "completion/rule_body_with_package_completion", completionList)

		// Should show pkg
		foundPkg := false
		for _, item := range completionList.Items {
			if item.Label == "pkg" {
				foundPkg = true
				break
			}
		}
		if !foundPkg {
			t.Logf("Expected 'pkg' in completion items, got: %v", completionList.Items)
		}
		idCounter++
	})
	assertNoUnhandledMessages(h, &logBuf)
}