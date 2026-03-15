package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"testing"
	"grammar/testutil"
)

func TestDefinition(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// 1. Open documents
	bContent := "rule_b = \"from b\";"
	bURI, _ := server.ParseURI("file:///b.grammar")
	h.send(newDidOpenNotification(bURI, bContent, 1))
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	aContent := "b = @import(\"b.grammar\");\nlocal_rule = b.rule_b;\nfinal_rule = local_rule;"
	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, aContent, 1))
	msg = h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	// 2. Test cross-file definition
	id := 1
	var definitionParams any = server.DefinitionParams{
		TextDocument: server.TextDocumentIdentifier{URI: aURI},
		Position:     server.Position{Line: 1, Character: 15}, // on 'rule_b'
	}
	h.send(newRequest(id, "textDocument/definition", &definitionParams))

	msg = h.read()
	assertResponseID(h, msg, id)

	var location server.Location
	if err := json.Unmarshal(mustMarshal(h, msg["result"])), &location); err != nil {
		t.Fatalf("Failed to unmarshal location: %v", err)
	}

	testutil.AssertSnapshotJSON(t, "definition/cross_file_definition", location)

	if location.URI != bURI {
		t.Errorf("Expected URI %s, got %s", bURI, location.URI)
	}
	if location.Range.Start.Line != 0 || location.Range.Start.Character != 0 {
		t.Errorf("Expected range start at 0:0, got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
	}

	// 3. Test same-file definition
	id = 2
	var localDefinitionParams any = server.DefinitionParams{
		TextDocument: server.TextDocumentIdentifier{URI: aURI},
		Position:     server.Position{Line: 2, Character: 15}, // on 'local_rule'
	}
	h.send(newRequest(id, "textDocument/definition", &localDefinitionParams))

	msg = h.read()
	assertResponseID(h, msg, id)
	if err := json.Unmarshal(mustMarshal(h, msg["result"])), &location); err != nil {
		t.Fatalf("Failed to unmarshal location: %v", err)
	}

	testutil.AssertSnapshotJSON(t, "definition/same_file_definition", location)

	if location.URI != aURI {
		t.Errorf("Expected URI %s, got %s", aURI, location.URI)
	}
	if location.Range.Start.Line != 1 || location.Range.Start.Character != 0 {
		t.Errorf("Expected range start at 1:0, got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
	}
	assertNoUnhandledMessages(h, &logBuf)
}

func TestDefinitionPackage(t *testing.T) {
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

	m1Content := `@package(\"pkg\");
rule_from_m1 = \"m1\";`
	m1Path := pkgDir + "/module_a.grammar"
	os.WriteFile(m1Path, []byte(m1Content), 0644)

	m1URI, _ := server.ParseURI("file://" + m1Path)
	h.send(newDidOpenNotification(m1URI, m1Content, 1))
	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	mainContent := `pkg = @import("pkg");
result = pkg.module_a.rule_from_m1;`
	mainPath := testDir + "/main.grammar"
	os.WriteFile(mainPath, []byte(mainContent), 0644)

	mainURI, _ := server.ParseURI("file://" + mainPath)
	h.send(newDidOpenNotification(mainURI, mainContent, 1))
	msg = h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}

	t.Run("definition of @package directive", func(t *testing.T) {
		id := 1
		var definitionParams any = server.DefinitionParams{
			TextDocument: server.TextDocumentIdentifier{URI: m1URI},
			Position:     server.Position{Line: 0, Character: 0},
		}
		h.send(newRequest(id, "textDocument/definition", &definitionParams))

		msg := h.read()
		assertResponseID(h, msg, id)
		testutil.AssertSnapshotJSON(t, "definition/package_directive_definition", msg["result"])
		t.Logf("Definition result for @package: %v", msg["result"])
	})

	t.Run("definition in same package", func(t *testing.T) {
		id := 2
		var definitionParams any = server.DefinitionParams{
			TextDocument: server.TextDocumentIdentifier{URI: mainURI},
			Position:     server.Position{Line: 1, Character: 25},
		}
		h.send(newRequest(id, "textDocument/definition", &definitionParams))

		msg := h.read()
		assertResponseID(h, msg, id)
		testutil.AssertSnapshotJSON(t, "definition/same_package_definition", msg["result"])
		t.Logf("Definition result for package member: %v", msg["result"])

		var location server.Location
		if err := json.Unmarshal(mustMarshal(h, msg["result"])), &location); err != nil {
			t.Fatalf("Failed to unmarshal location: %v", err)
		}

		if location.URI != m1URI {
			t.Errorf("Expected URI %s, got %s", m1URI, location.URI)
		}
		if location.Range.Start.Line != 1 || location.Range.Start.Character != 0 {
			t.Errorf("Expected range start at 1:0, got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
		}
	})

	t.Run("definition across packages", func(t *testing.T) {
		// lib/utils.grammar: @package("lib"); helper = "helper";
		// app/main.grammar: @package("app"); libPkg = @import("../lib"); use_lib = libPkg.utils.helper;
		libDir := testDir + "/lib"
		appDir := testDir + "/app"
		os.MkdirAll(libDir, 0755)
		os.MkdirAll(appDir, 0755)

		libContent := `@package(\"lib\"); helper = \"helper\";`
		libPath := libDir + "/utils.grammar"
		os.WriteFile(libPath, []byte(libContent), 0644)
		libURI, _ := server.ParseURI("file://" + libPath)
		h.send(newDidOpenNotification(libURI, libContent, 1))
		msgLib := h.read()
		if msgLib["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgLib)
		}

		appContent := `@package(\"app\"); libPkg = @import(\"../lib\"); appRule = libPkg.utils.helper;`
		appPath := appDir + "/main.grammar"
		os.WriteFile(appPath, []byte(appContent), 0644)
		appURI, _ := server.ParseURI("file://" + appPath)
		h.send(newDidOpenNotification(appURI, appContent, 1))
		msgApp := h.read()
		if msgApp["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics, got %v", msgApp)
		}

		id := 3
		var definitionParams any = server.DefinitionParams{
			TextDocument: server.TextDocumentIdentifier{URI: appURI},
			Position:     server.Position{Line: 0, Character: 60}, // on 'helper' in 'libPkg.utils.helper'
		}
		h.send(newRequest(id, "textDocument/definition", &definitionParams))

		msg := h.read()
		assertResponseID(h, msg, id)

		var location server.Location
		if err := json.Unmarshal(mustMarshal(h, msg["result"])), &location); err != nil {
			t.Fatalf("Failed to unmarshal location: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "definition/across_packages_definition", location)

		if location.URI != libURI {
			t.Errorf("Expected URI %s, got %s", libURI, location.URI)
		}
		if location.Range.Start.Line != 0 || location.Range.Start.Character != 17 { // Start of 'helper' in libContent
			t.Errorf("Expected range start at 0:17, got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
		}
	})

	t.Run("definition of module access", func(t *testing.T) {
		id := 4
		var definitionParams any = server.DefinitionParams{
			TextDocument: server.TextDocumentIdentifier{URI: mainURI},
			Position:     server.Position{Line: 1, Character: 16}, // on 'module_a' in 'pkg.module_a.rule_from_m1'
		}
		h.send(newRequest(id, "textDocument/definition", &definitionParams))

		msg := h.read()
		assertResponseID(h, msg, id)

		var location server.Location
		if err := json.Unmarshal(mustMarshal(h, msg["result"])), &location); err != nil {
			t.Fatalf("Failed to unmarshal location: %v", err)
		}

		testutil.AssertSnapshotJSON(t, "definition/module_access_definition", location)

		if location.URI != m1URI {
			t.Errorf("Expected URI %s, got %s", m1URI, location.URI)
		}
		if location.Range.Start.Line != 0 || location.Range.Start.Character != 0 {
			t.Errorf("Expected range start at 0:0 (start of module), got %d:%d", location.Range.Start.Line, location.Range.Start.Character)
		}
	})
	assertNoUnhandledMessages(h, &logBuf)
}