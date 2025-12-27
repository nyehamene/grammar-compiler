package ast

import (
	"grammar/token"
	"os"
	"strings"
	"testing"
)

func TestParseInvalidToken(t *testing.T) {
	testCases := []struct {
		path    string
		message string
	}{
		{"../testdata/parser/invalid_token/unterminated_string.grammar", "unterminated string literal"},
		{"../testdata/parser/invalid_token/unterminated_regex.grammar", "unterminated regex literal"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			content, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}
			srcRunes := []rune(string(content))
			tokenizer := token.NewTokenizer(srcRunes, false, false)
			tokens := tokenizer.Scan()
			parser := NewParser(tokens, srcRunes)
			_, err = parser.ParseFile()

			if err == nil {
				t.Fatalf("Expected a parsing error, but got none")
			}

			if !strings.Contains(err.Error(), tc.message) {
				t.Errorf("Error message mismatch, got: '%s', want contains: '%s'", err.Error(), tc.message)
			}
		})
	}
}
