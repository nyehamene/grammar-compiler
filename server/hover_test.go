package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"grammar/server"
	"os"
	"testing"
)

func TestHover(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// Setup for cross-file tests
	bContent := "rule_b = \"from b\";"
	bURI, _ := server.ParseURI("file:///b.grammar")
	h.send(newDidOpenNotification(bURI, bContent, 1))
	consumeDiagnostics(h)

	aContent := `
b = @import("b.grammar");
str = "hello";
re = /[a-z]/;
ext = $foo;
prod_a = "a";
prod_b = prod_a;
prod_c = b.rule_b;
`
	aURI, _ := server.ParseURI("file:///a.grammar")
	h.send(newDidOpenNotification(aURI, aContent, 2))
	consumeDiagnostics(h)

	testCases := []struct {
		name          string
		uri           server.DocumentUri
		position      server.Position
		expectedType  string
		expectedValue string
	}{
		{"string literal", aURI, server.Position{Line: 2, Character: 7}, "string", `"hello"`},
		{"regexp literal", aURI, server.Position{Line: 3, Character: 6}, "regexp", `/[a-z]/`},
		{"external value", aURI, server.Position{Line: 4, Character: 7}, "external", `$foo`},
		{"local production", aURI, server.Position{Line: 6, Character: 9}, "production", `"a";`},
		{"binding", aURI, server.Position{Line: 1, Character: 0}, "module", "b.grammar"},
		{"imported member", aURI, server.Position{Line: 7, Character: 11}, "production", `"from b";`},
	}

	idCounter := 1
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := idCounter
			var hoverReqParams any = server.HoverParams{
				TextDocumentPositionParams: server.TextDocumentPositionParams{
					TextDocument: server.TextDocumentIdentifier{URI: tc.uri},
					Position:     tc.position,
				},
			}
			h.send(newRequest(id, "textDocument/hover", &hoverReqParams))

			msg := h.read()
			assertResponseID(h, msg, id)

			resultData, err := json.Marshal(msg["result"])
			if err != nil {
				t.Fatalf("Failed to marshal hover result: %v", err)
			}
			var hover server.Hover
			json.Unmarshal(resultData, &hover)

			if hover.Contents.Kind != server.MarkupKindMarkdown {
				t.Errorf("Expected hover kind to be Markdown, got %s", hover.Contents.Kind)
			}

			expectedHoverValue := fmt.Sprintf("(%s)\n\n```grammar\n%s\n```\n", tc.expectedType, tc.expectedValue)
			if hover.Contents.Value != expectedHoverValue {
				t.Errorf("Expected hover value:\n%s\nGot:\n%s", expectedHoverValue, hover.Contents.Value)
			}

			idCounter++
		})
	}
	assertNoUnhandledMessages(h, &logBuf)
}

func TestHoverPackage(t *testing.T) {
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

	m1Content := `@package("mypackage");
rule_m1 = "from m1";`
	m1Path := pkgDir + "/module_a.grammar"
	os.WriteFile(m1Path, []byte(m1Content), 0644)

	m1URI, _ := server.ParseURI("file://" + m1Path)
	h.send(newDidOpenNotification(m1URI, m1Content, 1))
	consumeDiagnostics(h)

	mainContent := `pkg = @import("pkg");
result = pkg.module_a.rule_m1;`
	mainPath := testDir + "/main.grammar"
	os.WriteFile(mainPath, []byte(mainContent), 0644)

	mainURI, _ := server.ParseURI("file://" + mainPath)
	h.send(newDidOpenNotification(mainURI, mainContent, 1))
	consumeDiagnostics(h)

	t.Run("hover on @package directive", func(t *testing.T) {
		id := 1
		var hoverReqParams any = server.HoverParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: m1URI},
				Position:     server.Position{Line: 0, Character: 0},
			},
		}
		h.send(newRequest(id, "textDocument/hover", &hoverReqParams))

		msg := h.read()
		assertResponseID(h, msg, id)

		resultData, _ := json.Marshal(msg["result"])
		var hover server.Hover
		json.Unmarshal(resultData, &hover)

		t.Logf("Hover on @package: %s", hover.Contents.Value)
	})

	t.Run("hover on package binding", func(t *testing.T) {
		id := 2
		var hoverReqParams any = server.HoverParams{
			TextDocumentPositionParams: server.TextDocumentPositionParams{
				TextDocument: server.TextDocumentIdentifier{URI: mainURI},
				Position:     server.Position{Line: 0, Character: 0},
			},
		}
		h.send(newRequest(id, "textDocument/hover", &hoverReqParams))

		msg := h.read()
		assertResponseID(h, msg, id)

		resultData, _ := json.Marshal(msg["result"])
		var hover server.Hover
		json.Unmarshal(resultData, &hover)

		t.Logf("Hover on package binding: %s", hover.Contents.Value)
	})
	assertNoUnhandledMessages(h, &logBuf)
}
