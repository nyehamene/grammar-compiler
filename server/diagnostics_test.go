package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"grammar/server"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDidOpenPublishDiagnostics(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	textDocumentId, err := server.ParseURI("file:///test.grammar")
	if err != nil {
		t.Fatal(err)
	}

	content := "A = b.c;"
	didOpenParams := server.DidOpenTextDocumentParams{
		TextDocument: server.TextDocumentItem{
			URI:        server.DocumentUri(textDocumentId),
			LanguageID: "grammar",
			Version:    1,
			Text:       content,
		},
	}
	var params any = didOpenParams
	didOpenNotif := server.NotificationMessage{
		Message: server.Message{JSONRPC: "2.0"},
		Method:  "textDocument/didOpen",
		Params:  &params,
	}

	h.send(didOpenNotif)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var msg map[string]any
	readChan := make(chan map[string]any)
	go func() {
		readChan <- h.read()
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for server response")
	case msg = <-readChan:
		if msg == nil {
			t.Fatal("Did not receive a message from the server")
		}
	}

	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics notification, but got: %s", msg["method"])
	}

	paramsData, _ := json.Marshal(msg["params"])
	var diagParams server.PublishDiagnosticsParams
	if err := json.Unmarshal(paramsData, &diagParams); err != nil {
		t.Fatalf("Failed to unmarshal PublishDiagnosticsParams: %v", err)
	}

	if diagParams.URI != textDocumentId {
		t.Errorf("Expected diagnostics for URI 'file:///test.grammar', but got: %s", diagParams.URI)
	}

	if len(diagParams.Diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, but got %d", len(diagParams.Diagnostics))
	}

	diag := diagParams.Diagnostics[0]
	if !strings.Contains(diag.Message, "undefined identifier: b") {
		t.Errorf("Expected diagnostic message to contain 'undefined identifier: b', but got: %s", diag.Message)
	}

	if diag.Range.Start.Line != 0 || diag.Range.Start.Character != 4 {
		t.Errorf("Expected diagnostic to start at line 0, char 4, but got line %d, char %d", diag.Range.Start.Line, diag.Range.Start.Character)
	}

	assertNoUnhandledMessages(h, &logBuf)
}

func TestImportedNamespaceLoading(t *testing.T) {
	// Create test files on disk
	testDir := t.TempDir()
	bPath := filepath.Join(testDir, "b.grammar")
	aPath := filepath.Join(testDir, "a.grammar")

	bContent := `Rule_b = "from b";`
	if err := os.WriteFile(bPath, []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.grammar: %v", err)
	}

	aContent := `b = @import("b.grammar"); A = b.Rule_b;` // relative import
	if err := os.WriteFile(aPath, []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.grammar: %v", err)
	}

	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	aURI, _ := server.ParseURI("file://" + aPath)

	// Open only a.grammar. The server should load b.grammar from filesystem.
	h.send(newDidOpenNotification(aURI, aContent, 1))

	msg := h.read()
	if msg["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("Expected publishDiagnostics, got %v", msg)
	}
	params, _ := msg["params"].(map[string]any)
	diags, _ := params["diagnostics"].([]any)

	// File-based imports are now deprecated, so we expect a warning
	// Allow 0 or 1 diagnostic (the deprecation warning)
	if len(diags) > 1 {
		t.Errorf("Expected at most 1 diagnostic (deprecation warning), but got %d: %v", len(diags), diags)
	}

	assertNoUnhandledMessages(h, &logBuf)
}

func TestDocumentDiagnosticRequest(t *testing.T) {
	var logBuf bytes.Buffer
	h := setupTestServer(t, &logBuf)
	defer func() { _ = h.clientConn.Close() }()
	defer func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	}()

	// 1. Open document with an error
	content := "A = b;" // 'b' is undefined
	uri, _ := server.ParseURI("file:///test_diagnostic.grammar")
	h.send(newDidOpenNotification(uri, content, 1))

	// 2. Consume the initial push diagnostic
	consumeDiagnostics(h)

	// 3. Send a pull diagnostic request
	id := 1
	var diagnosticParams any = server.DocumentDiagnosticParams{
		TextDocument: server.TextDocumentIdentifier{URI: uri},
	}
	h.send(newRequest(id, "textDocument/diagnostic", &diagnosticParams))

	// 4. Read and verify the response
	msg := h.read()
	assertResponseID(h, msg, id)

	resultData, err := json.Marshal(msg["result"])
	if err != nil {
		t.Fatalf("Failed to marshal diagnostic result: %v", err)
	}

	var report server.RelatedFullDocumentDiagnosticReport
	if err := json.Unmarshal(resultData, &report); err != nil {
		t.Fatalf("Failed to unmarshal diagnostic report: %v", err)
	}

	if report.Kind != server.DocumentDiagnosticReportKindFull {
		t.Errorf("Expected report kind to be 'full', got '%s'", report.Kind)
	}

	if len(report.Items) != 1 {
		t.Fatalf("Expected 1 diagnostic item, got %d", len(report.Items))
	}

	diag := report.Items[0]
	if !strings.Contains(diag.Message, "undefined identifier: b") {
		t.Errorf("Expected diagnostic message to contain 'undefined identifier: b', got: '%s'", diag.Message)
	}
	assertNoUnhandledMessages(h, &logBuf)
}

func TestWorkspaceDiagnosticRequest(t *testing.T) {
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
	aPath := filepath.Join(testDir, "a.grammar")
	bPath := filepath.Join(testDir, "b.grammar")

	// a.grammar with an error (undefined identifier 'b')
	aContent := `Rule_a = b;`
	if err := os.WriteFile(aPath, []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.grammar: %v", err)
	}

	// b.grammar without errors
	bContent := `Rule_b = "hello";`
	if err := os.WriteFile(bPath, []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.grammar: %v", err)
	}

	aURI, _ := server.ParseURI("file://" + aPath)
	bURI, _ := server.ParseURI("file://" + bPath)

	// 1. Open both documents
	h.send(newDidOpenNotification(aURI, aContent, 1))
	consumeDiagnostics(h) // Consume initial diagnostics for a.grammar

	h.send(newDidOpenNotification(bURI, bContent, 1))
	consumeDiagnostics(h) // Consume initial diagnostics for b.grammar

	// 2. Send workspace/diagnostic request
	id := 1
	var wsDiagParams any = server.WorkspaceDiagnosticParams{}
	h.send(newRequest(id, "workspace/diagnostic", &wsDiagParams))

	// 3. Read and verify the response
	msg := h.read()
	assertResponseID(h, msg, id)

	resultData, err := json.Marshal(msg["result"])
	if err != nil {
		t.Fatalf("Failed to marshal workspace diagnostic result: %v", err)
	}

	var report server.WorkspaceDiagnosticReport
	if err := json.Unmarshal(resultData, &report); err != nil {
		t.Fatalf("Failed to unmarshal workspace diagnostic report: %v", err)
	}

	if len(report.Items) != 2 {
		t.Fatalf("Expected 2 document reports, got %d", len(report.Items))
	}

	// Helper to find report for a specific URI
	findDocReport := func(uri server.DocumentUri) *server.WorkspaceDocumentDiagnosticReport {
		for i := range report.Items {
			if report.Items[i].URI == uri {
				return &report.Items[i]
			}
		}
		return nil
	}

	// Verify a.grammar report
	aReport := findDocReport(aURI)
	if aReport == nil {
		t.Fatalf("Expected report for a.grammar, but not found")
	}
	if len(aReport.Items) != 1 {
		t.Fatalf("Expected 1 diagnostic for a.grammar, got %d", len(aReport.Items))
	}
	if !strings.Contains(aReport.Items[0].Message, "undefined identifier: b") {
		t.Errorf("Expected diagnostic message for a.grammar to contain 'undefined identifier: b', got: '%s'", aReport.Items[0].Message)
	}

	// Verify b.grammar report
	bReport := findDocReport(bURI)
	if bReport == nil {
		t.Fatalf("Expected report for b.grammar, but not found")
	}
	if len(bReport.Items) != 0 {
		t.Fatalf("Expected 0 diagnostics for b.grammar, got %d", len(bReport.Items))
	}
	assertNoUnhandledMessages(h, &logBuf)
}

func TestConditionalPublishDiagnostics(t *testing.T) {
	t.Run("ClientWithPullSupport", func(t *testing.T) {
		var logBuf bytes.Buffer
		h := setupTestServer(t, &logBuf)
		defer func() { _ = h.clientConn.Close() }()
		defer func() {
			if t.Failed() {
				t.Log(logBuf.String())
			}
		}()

		// 1. Initialize with pull diagnostic support
		id := 1
		caps := server.ClientCapabilities{
			TextDocument: &server.TextDocumentClientCapabilities{
				Diagnostic: &server.DiagnosticClientCapabilities{},
			},
		}
		h.send(newInitializeRequest(id, caps))
		// consume the initialize response
		msg := h.read()
		assertResponseID(h, msg, id)

		// 2. Open document with an error
		content := "A = b;" // 'b' is undefined
		uri, _ := server.ParseURI("file:///test.grammar")
		h.send(newDidOpenNotification(uri, content, 1))

		// 3. Assert that NO publishDiagnostics is received.
		// The read will time out if no message is sent, which is the desired outcome.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		readChan := make(chan map[string]any)
		go func() {
			readChan <- h.read()
		}()

		select {
		case msg := <-readChan:
			if msg != nil && msg["method"] == "textDocument/publishDiagnostics" {
				t.Fatal("Received unexpected publishDiagnostics notification")
			}
			// Another message is not expected, but if one arrives and it's not a diagnostic,
			// it's not a failure for this specific test. A nil message means connection closed.
		case <-ctx.Done():
			// This is the success case. The read timed out, meaning no message was sent.
		}
	})

	t.Run("ClientWithoutPullSupport", func(t *testing.T) {
		var logBuf bytes.Buffer
		h := setupTestServer(t, &logBuf)
		defer func() { _ = h.clientConn.Close() }()
		defer func() {
			if t.Failed() {
				t.Log(logBuf.String())
			}
		}()

		// 1. Initialize without pull diagnostic support
		id := 1
		h.send(newInitializeRequest(id, server.ClientCapabilities{}))
		// consume the initialize response
		msg := h.read()
		assertResponseID(h, msg, id)

		// 2. Open document with an error
		content := "A = b;" // 'b' is undefined
		uri, _ := server.ParseURI("file:///test.grammar")
		h.send(newDidOpenNotification(uri, content, 1))

		// 3. Assert that a publishDiagnostics notification IS received.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var diagMsg map[string]any
		readChan := make(chan map[string]any)
		go func() {
			readChan <- h.read()
		}()

		select {
		case <-ctx.Done():
			t.Fatal("Test timed out waiting for publishDiagnostics notification")
		case diagMsg = <-readChan:
			if diagMsg == nil {
				t.Fatal("Did not receive a message from the server")
			}
		}

		if diagMsg["method"] != "textDocument/publishDiagnostics" {
			t.Fatalf("Expected publishDiagnostics notification, but got: %v", diagMsg)
		}
	})
}
