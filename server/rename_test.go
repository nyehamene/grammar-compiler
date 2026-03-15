package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"path/filepath" // Added import
	"testing"
	"grammar/testutil"
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
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	userContent := "common = @import(\"common.grammar\");\nlocal = common.rename_rule;"
	userURI, _ := server.ParseURI("file:///user.grammar")
	h.send(newDidOpenNotification(userURI, userContent, 1))
	msgUser := h.read()
	if msgUser["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msgUser)
	}

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

	testutil.AssertSnapshotJSON(t, "rename/cross_file_rename", workspaceEdit)

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
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	t.Run("prepare rename on rule in package", func(t *testing.T) {
		id := 1
		var prepareParams any = server.PrepareRenameParams{
			TextDocument: server.TextDocumentIdentifier{URI: m1URI},
			Position:     server.Position{Line: 1, Character: 1},
		}
		h.send(newRequest(id, "textDocument/prepareRename", &prepareParams))
		msg := h.read()
		assertResponseID(h, msg, id)
		testutil.AssertSnapshotJSON(t, "rename/prepare_rename_rule_in_package", msg["result"])
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
		msgLib := h.read()
		if msgLib["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgLib)
		}

		appContent := `lib_pkg = @import("../lib");
use_rule = lib_pkg.lib.lib_rule;`
		appPath := testDir + "/app.grammar"
		os.WriteFile(appPath, []byte(appContent), 0644)

		appURI, _ := server.ParseURI("file://" + appPath)
		h.send(newDidOpenNotification(appURI, appContent, 1))
		msgApp := h.read()
		if msgApp["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgApp)
		}

		id := 2
		var prepareParams any = server.PrepareRenameParams{
			TextDocument: server.TextDocumentIdentifier{URI: libURI},
			Position:     server.Position{Line: 1, Character: 1},
		}
		h.send(newRequest(id, "textDocument/prepareRename", &prepareParams))
		msg := h.read()
		assertResponseID(h, msg, id)
		testutil.AssertSnapshotJSON(t, "rename/prepare_rename_across_packages", msg["result"])
		t.Logf("PrepareRename result for lib_rule: %v", msg["result"])
	})

	t.Run("rename rule in same package", func(t *testing.T) {
		// pkg/module_b.grammar references pkg/module_a.grammar
		bContent := `@package("pkg");
use_rule_a = module_a.rule_to_rename;`
		bPath := pkgDir + "/module_b.grammar"
		os.WriteFile(bPath, []byte(bContent), 0644)
		bURI, _ := server.ParseURI("file://" + bPath)
		h.send(newDidOpenNotification(bURI, bContent, 1))
		msgB := h.read()
		if msgB["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgB)
		}

		id := 3
		newName := "new_rule_name"
		var renameParams any = server.RenameParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: m1URI},
				Position:     server.Position{Line: 1, Character: 1}, // on 'rule_to_rename' in module_a
			},
			NewName: newName,
		}
		h.send(newRequest(id, "textDocument/rename", &renameParams))
		msg := h.read()
		assertResponseID(h, msg, id)

		var workspaceEdit server.WorkspaceEdit
		if err := json.Unmarshal(mustMarshal(h, msg["result"]), &workspaceEdit); err != nil {
			t.Fatalf("Failed to unmarshal workspace edit: %v", err)
		}

		if len(workspaceEdit.Changes) != 2 {
			t.Fatalf("Expected edits in 2 files, got %d", len(workspaceEdit.Changes))
		}

		// Check changes in module_a.grammar
		editsA, ok := workspaceEdit.Changes[m1URI.String()]
		if !ok || len(editsA) != 1 {
			t.Fatalf("Expected 1 edit in %s", m1URI)
		}
		if editsA[0].NewText != newName {
			t.Errorf("Wrong new name in %s: got %s, want %s", m1URI, editsA[0].NewText, newName)
		}

		// Check changes in module_b.grammar
		editsB, ok := workspaceEdit.Changes[bURI.String()]
		if !ok || len(editsB) != 1 {
			t.Fatalf("Expected 1 edit in %s", bURI)
		}
		if editsB[0].NewText != newName {
			t.Errorf("Wrong new name in %s: got %s, want %s", bURI, editsB[0].NewText, newName)
		}
	})

	t.Run("rename module", func(t *testing.T) {
		// Use a fresh set of files to avoid interference
		tempTestDir := t.TempDir()
		pkgDir := filepath.Join(tempTestDir, "my_package")
		os.MkdirAll(pkgDir, 0755)

		// Original module_a
		oldModuleAName := "module_a"
		oldModuleAPath := filepath.Join(pkgDir, oldModuleAName+".grammar")
		os.WriteFile(oldModuleAPath, []byte(`@package("my_package"); rule_in_a = "a";`), 0644)
		oldModuleAURI, _ := server.ParseURI("file://" + oldModuleAPath)
		h.send(newDidOpenNotification(oldModuleAURI, string([]byte(`@package("my_package"); rule_in_a = "a";`)), 1))
		msgA := h.read()
		if msgA["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgA)
		}

		// module_b references module_a
		moduleBPath := filepath.Join(pkgDir, "module_b.grammar")
		os.WriteFile(moduleBPath, []byte(`@package("my_package"); rule_in_b = `+oldModuleAName+`.rule_in_a;`), 0644)
		moduleBURI, _ := server.ParseURI("file://" + moduleBPath)
		h.send(newDidOpenNotification(moduleBURI, string([]byte(`@package("my_package"); rule_in_b = `+oldModuleAName+`.rule_in_a;`)), 1))
		msgB := h.read()
		if msgB["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgB)
		}

		id := 4
		newModuleName := "module_new_name"
		var renameParams any = server.RenameParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: oldModuleAURI},
				Position:     server.Position{Line: 0, Character: 15}, // on 'module_a' in @package("my_package"); rule_in_a = "a"; (conceptually renaming the module itself)
			},
			NewName: newModuleName,
		}
		h.send(newRequest(id, "textDocument/rename", &renameParams))
		msg := h.read()
		assertResponseID(h, msg, id)

		var workspaceEdit server.WorkspaceEdit
		if err := json.Unmarshal(mustMarshal(h, msg["result"]), &workspaceEdit); err != nil {
			t.Fatalf("Failed to unmarshal workspace edit: %v", err)
		}

		// Expect edits in module_b.grammar to update 'module_a' to 'module_new_name'
		editsB, ok := workspaceEdit.Changes[moduleBURI.String()]
		if !ok || len(editsB) == 0 {
			t.Fatalf("Expected edits in %s", moduleBURI)
		}

		foundEdit := false
		for _, edit := range editsB {
			if edit.NewText == newModuleName {
				foundEdit = true
				break
			}
		}
		if !foundEdit {
			t.Errorf("Expected edit to change module name to '%s', but not found in %v", newModuleName, editsB)
		}
	})
	assertNoUnhandledMessages(h, &logBuf)
}
