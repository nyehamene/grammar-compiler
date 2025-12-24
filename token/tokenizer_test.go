package token

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	// Create a dummy file for testing various token types
	testSrc := `document = "document"; param = "param"; num = 123; ident = /[a-zA-Z_][a-zA-Z0-9_]*/; import "some/path";`
	srcRunes := []rune(testSrc)
	tokenizer := NewTokenizer(srcRunes)
	tokens := tokenizer.Scan()

	expectedTokens := []struct {
		kind  Kind
		value string
	}{
		{Ident, "document"},
		{Assign, "="},
		{String, `"document"`},
		{Semicolon, ";"},
		{Ident, "param"},
		{Assign, "="},
		{String, `"param"`},
		{Semicolon, ";"},
		{Ident, "num"},
		{Assign, "="},
		{Number, "123"},
		{Semicolon, ";"},
		{Ident, "ident"},
		{Assign, "="},
		{Regex, `/[a-zA-Z_][a-zA-Z0-9_]*/`},
		{Semicolon, ";"},
		{Import, "import"},
		{String, `"some/path"`},
		{Semicolon, ";"},
		{EOF, ""},
	}

	if len(tokens) != len(expectedTokens) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	for i, exp := range expectedTokens {
		tok := tokens[i]
		if tok.Kind != exp.kind {
			t.Errorf("Token %d: Expected kind %s, got %s (value: %q)", i, exp.kind, tok.Kind, tokenizer.Literal(tok, srcRunes))
		}
		val := tokenizer.Literal(tok, srcRunes)
		if val != exp.value {
			t.Errorf("Token %d: Expected value %q, got %q", i, exp.value, val)
		}
		if tok.State != Valid {
			t.Errorf("Token %d: Expected state Valid, got Invalid", i)
		}
	}
}

func TestTokenizerExampleFiles(t *testing.T) {
	exampleDir := "../example" // Corrected path to be relative to the 'token' package
	files, err := ioutil.ReadDir(exampleDir)
	if err != nil {
		t.Fatalf("Failed to read example directory: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".grammar") {
			t.Run(file.Name(), func(t *testing.T) {
				filePath := filepath.Join(exampleDir, file.Name())
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read example file %s: %v", filePath, err)
				}
				srcRunes := []rune(string(fileContent))

				tokenizer := NewTokenizer(srcRunes)
				tokens := tokenizer.Scan()

				for _, tok := range tokens {
					if tok.State == Invalid {
						t.Errorf("File %s: Found invalid token %q", file.Name(), tokenizer.Literal(tok, srcRunes))
					}
				}
			})
		}
	}
}