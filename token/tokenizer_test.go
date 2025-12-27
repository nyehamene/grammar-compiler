package token

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	// Create a dummy file for testing various token types
	testSrc := `document = "document"; param = "param"; num = 123; ident = /[a-zA-Z_][a-zA-Z0-9_]*/;`
	srcRunes := []rune(testSrc)
	tokenizer := NewTokenizer(srcRunes, false, false)
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
		{EOF, ""},
	}

	if len(tokens) != len(expectedTokens) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	for i, exp := range expectedTokens {
		tok := tokens[i]
		if tok.Kind != exp.kind {
			t.Errorf("Token %d: Expected kind %s, got %s (value: %q)", i, exp.kind, tok.Kind, Literal(tok, srcRunes))
		}
		val := Literal(tok, srcRunes)
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
	files, err := os.ReadDir(exampleDir)
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

				tokenizer := NewTokenizer(srcRunes, false, false)
				tokens := tokenizer.Scan()

				for _, tok := range tokens {
					if tok.State == Invalid {
						t.Errorf("File %s: Found invalid token %q", file.Name(), Literal(tok, srcRunes))
					}
				}
			})
		}
	}
}

func TestScanRegex(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		valid   bool
		wantVal string
	}{
		{
			name:    "valid escaped slash",
			input:   `/[^\/]/`,
			valid:   true,
			wantVal: `/[^\/]/`,
		},
		{
			name:    "valid other escapes",
			input:   `/\n\t\d\s\(\)\[\]\{\}\.\\/`,
			valid:   true,
			wantVal: `/\n\t\d\s\(\)\[\]\{\}\.\\/`,
		},
		{
			name:    "valid trailing escape",
			input:   `/abc\\/`,
			valid:   true,
			wantVal: `/abc\\/`,
		},
		{
			name:    "invalid escape sequence",
			input:   `/\a/`,
			valid:   false,
			wantVal: `/\a/`,
		},
		{
			name:    "unterminated regex",
			input:   `/abc`,
			valid:   false,
			wantVal: `/abc`,
		},
		{
			name:    "unterminated with escape",
			input:   `/abc\/`,
			valid:   false,
			wantVal: `/abc\/`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcRunes := []rune(tc.input)
			tokenizer := NewTokenizer(srcRunes, false, false)
			tokens := tokenizer.Scan()

			// We expect 2 tokens: Regex and EOF
			if len(tokens) != 2 {
				t.Fatalf("Expected 2 tokens, got %d", len(tokens))
			}

			regexToken := tokens[0]
			if regexToken.Kind != Regex {
				t.Fatalf("Expected first token to be REGEX, got %s", regexToken.Kind)
			}

			if (regexToken.State == Valid) != tc.valid {
				t.Errorf("Expected token validity to be %t, but got %t", tc.valid, (regexToken.State == Valid))
			}

			gotVal := Literal(regexToken, srcRunes)
			if gotVal != tc.wantVal {
				t.Errorf("Expected token value to be %q, got %q", tc.wantVal, gotVal)
			}
		})
	}
}
