package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"testing"
)

func TestRename(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
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
	if err := json.Unmarshal(mustMarshal(h, msg["result"]), &prepareRange); err != nil {
		t.Fatalf("Failed to unmarshal prepare rename range: %v", err)
	}
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
	if err := json.Unmarshal(mustMarshal(h, msg["result"]), &workspaceEdit); err != nil {
		t.Fatalf("Failed to unmarshal workspace edit: %v", err)
	}

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
	assertNoUnhandledMessages(h, &logBuf)
}

func TestRenamePackage(t *testing.T) {
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

	m1Content := `@package("pkg");
rule_to_rename = "m1";`
	m1Path := pkgDir + "/module_a.grammar"
	os.WriteFile(m1Path, []byte(m1Content), 0644)

	m1URI, _ := server.ParseURI("file://" + m1Path)
	h.send(newDidOpenNotification(m1URI, m1Content, 1))
	consumeDiagnostics(h)

	t.Run("prepare rename on rule in package", func(t *testing.T) {
		id := 1
		var prepareParams any = server.PrepareRenameParams{
			TextDocument: server.TextDocumentIdentifier{URI: m1URI},
			Position:     server.Position{Line: 1, Character: 1},
		}
		h.send(newRequest(id, "textDocument/prepareRename", &prepareParams))
		msg := h.read()
		assertResponseID(h, msg, id)
		t.Logf("PrepareRename result: %v", msg["result"])
	})

	t.Run("rename rule across packages", func(t *testing.T) {
		libDir := testDir + "/lib"
		os.MkdirAll(libDir, 0755)

		libContent := `@package("lib");
lib_rule = "lib";`
		libPath := libDir + "/lib.grammar"
		os.WriteFile(libPath, []byte(libContent), 0644)

		libURI, _ := server.ParseURI("file://" + libPath)
		h.send(newDidOpenNotification(libURI, libContent, 1))
		consumeDiagnostics(h)

		appContent := `lib_pkg = @import("../lib");
use_rule = lib_pkg.lib.lib_rule;`
		appPath := testDir + "/app.grammar"
		os.WriteFile(appPath, []byte(appContent), 0644)

		appURI, _ := server.ParseURI("file://" + appPath)
		h.send(newDidOpenNotification(appURI, appContent, 1))
		consumeDiagnostics(h)

		id := 2
		var prepareParams any = server.PrepareRenameParams{
			TextDocument: server.TextDocumentIdentifier{URI: libURI},
			Position:     server.Position{Line: 1, Character: 1},
		}
		h.send(newRequest(id, "textDocument/prepareRename", &prepareParams))
		msg := h.read()
		assertResponseID(h, msg, id)
		t.Logf("PrepareRename result for lib_rule: %v", msg["result"])
	})
	assertNoUnhandledMessages(h, &logBuf)
}
