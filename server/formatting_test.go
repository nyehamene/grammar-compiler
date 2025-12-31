package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"grammar/server"
	"strings"
	"testing"
	"time"
)

func TestTextDocumentFormatting(t *testing.T) {
	const docURI = "file:///path/to/format_test.grammar"
	unformattedContent := `a="a";
longname = "b";`

	textDocumentId, err := server.ParseURI(docURI)
	if err != nil {
		t.Fatal(err)
	}

	// Using json.Marshal to handle escaping of the content
	didOpenParams := server.DidOpenTextDocumentParams{
		TextDocument: server.TextDocumentItem{
			URI:        textDocumentId,
			LanguageID: "grammar",
			Version:    1,
			Text:       unformattedContent,
		},
	}
	didOpenParamsBytes, _ := json.Marshal(didOpenParams)

	formattingParams := server.DocumentFormattingParams{
		TextDocument: server.TextDocumentIdentifier{
			URI: textDocumentId,
		},
	}
	formattingParamsBytes, _ := json.Marshal(formattingParams)

	initializeMessage := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":123,"rootUri":null,"capabilities":{}}}`
	didOpenMessage := fmt.Sprintf(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":%s}`, string(didOpenParamsBytes))
	formattingMessage := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":%s}`, string(formattingParamsBytes))

	var in bytes.Buffer
	var out bytes.Buffer
	var logOut bytes.Buffer
	defer func() {
		if t.Failed() {
			t.Log(logOut.String())
		}
	}()

	srv := server.NewServer(&in, &out, &logOut)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		srv.Start()
		cancel()
	}()

	// Send messages
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(initializeMessage), initializeMessage))
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(didOpenMessage), didOpenMessage))
	in.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(formattingMessage), formattingMessage))

	<-ctx.Done()

	// Find and decode the formatting response
	responses := strings.Split(out.String(), "Content-Length:")
	var formattingResponse server.ResponseMessage
	found := false
	for _, respStr := range responses {
		if len(respStr) == 0 {
			continue
		}
		// Clean up the response part
		parts := strings.SplitN(respStr, "\r\n\r\n", 2)
		if len(parts) < 2 {
			continue
		}
		var resp server.ResponseMessage
		if err := json.Unmarshal([]byte(parts[1]), &resp); err != nil {
			t.Logf("failed to unmarshal part: %v, content: %s", err, parts[1])
			continue
		}
		if resp.ID != nil && *resp.ID == 2 {
			formattingResponse = resp
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("did not find formatting response in server output: %s", out.String())
	}

	if formattingResponse.Error != nil {
		t.Fatalf("expected no error, got: %+v", formattingResponse.Error)
	}

	resultSlice, ok := (*formattingResponse.Result).([]any)
	if !ok {
		t.Fatalf("result is not a slice, got %T", *formattingResponse.Result)
	}

	if len(resultSlice) != 1 {
		t.Fatalf("expected 1 text edit, got %d", len(resultSlice))
	}

	var textEdit server.TextEdit
	editBytes, err := json.Marshal(resultSlice[0])
	if err != nil {
		t.Fatalf("Failed to marshal text edit: %v", err)
	}
	if err := json.Unmarshal(editBytes, &textEdit); err != nil {
		t.Fatalf("Failed to unmarshal text edit: %v", err)
	}

	expectedFormatted := `a        = "a";
longname = "b";
`
	if textEdit.NewText != expectedFormatted {
		t.Errorf("wrong new text\ngot:\n%q\nwant:\n%q", textEdit.NewText, expectedFormatted)
	}
}
