package server_test

import (
	"bytes"
	"encoding/json"
	"grammar/server"
	"os"
	"testing"
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
	consumeDiagnostics(h)

	aContent := "b = @import(\"b.grammar\");\nlocal_rule = b.rule_b;\nfinal_rule = local_rule;"
	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, aContent, 1))
	consumeDiagnostics(h)

	// 2. Test cross-file definition
	id := 1
	var definitionParams any = server.DefinitionParams{
		TextDocument: server.TextDocumentIdentifier{URI: aURI},
		Position:     server.Position{Line: 1, Character: 15}, // on 'rule_b'
	}
	h.send(newRequest(id, "textDocument/definition", &definitionParams))

	msg := h.read()
	assertResponseID(h, msg, id)

	var location server.Location
	if err := json.Unmarshal(mustMarshal(h, msg["result"]), &location); err != nil {
		t.Fatalf("Failed to unmarshal location: %v", err)
	}

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
	if err := json.Unmarshal(mustMarshal(h, msg["result"]), &location); err != nil {
		t.Fatalf("Failed to unmarshal location: %v", err)
	}

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

	m1Content := `@package("pkg");
rule_from_m1 = "m1";`
	m1Path := pkgDir + "/module_a.grammar"
	os.WriteFile(m1Path, []byte(m1Content), 0644)

	m1URI, _ := server.ParseURI("file://" + m1Path)
	h.send(newDidOpenNotification(m1URI, m1Content, 1))
	consumeDiagnostics(h)

	mainContent := `pkg = @import("pkg");
result = pkg.module_a.rule_from_m1;`
	mainPath := testDir + "/main.grammar"
	os.WriteFile(mainPath, []byte(mainContent), 0644)

	mainURI, _ := server.ParseURI("file://" + mainPath)
	h.send(newDidOpenNotification(mainURI, mainContent, 1))
	consumeDiagnostics(h)

	t.Run("definition of @package directive", func(t *testing.T) {
		id := 1
		var definitionParams any = server.DefinitionParams{
			TextDocument: server.TextDocumentIdentifier{URI: m1URI},
			Position:     server.Position{Line: 0, Character: 0},
		}
		h.send(newRequest(id, "textDocument/definition", &definitionParams))

		msg := h.read()
		assertResponseID(h, msg, id)
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
		t.Logf("Definition result for package member: %v", msg["result"])
	})
	assertNoUnhandledMessages(h, &logBuf)
}
