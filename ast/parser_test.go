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
		{"../testdata/parser/invalid_token/unterminated_string.grammar", "invalid string literal"},
		{"../testdata/parser/invalid_token/unterminated_regex.grammar", "invalid regex literal"},
		{"../testdata/parser/invalid_token/invalid_regex_escape.grammar", "invalid regex literal"},
		{"../testdata/parser/invalid_token/invalid_string_escape.grammar", "invalid string literal"},
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
func TestParseRecovery(t *testing.T) {
	path := "../testdata/parser/recovery.grammar"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	srcRunes := []rune(string(content))
	tokenizer := token.NewTokenizer(srcRunes, false, false)
	tokens := tokenizer.Scan()

	parser := NewParser(tokens, srcRunes)
	file, err := parser.ParseFile()

	if err == nil {
		t.Fatalf("Expected parsing to produce errors, but it didn't")
	}

	errList, ok := err.(ErrorList)
	if !ok {
		t.Fatalf("Expected err to be of type ErrorList, got %T", err)
	}

	expectedErrors := []string{
		"expected ASSIGN, got IDENT",
		"rule declaration must have a body",
		"unexpected token PIPE",
	}

	if len(errList) != len(expectedErrors) {
		t.Fatalf("Expected %d errors, but got %d. Errors: %v", len(expectedErrors), len(errList), errList)
	}

	for _, expectedErr := range expectedErrors {
		found := false
		for _, actualErr := range errList {
			if strings.Contains(actualErr.Error(), expectedErr) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error containing\n\n\t%s\n\nbut none found.", expectedErr)
		}
	}

	if file == nil {
		t.Fatalf("Expected a non-nil AST, but got nil")
	}

	if len(file.Decls) != 1 {
		t.Fatalf("Expected 1 valid declaration, but got %d", len(file.Decls))
	}

	rule, ok := file.Decls[0].(*RuleDecl)
	if !ok {
		t.Fatalf("Expected a RuleDecl, but got %T", file.Decls[0])
	}
	if rule.Name.Name != "valid_rule" {
		t.Errorf("Expected rule name 'valid_rule', but got '%s'", rule.Name.Name)
	}
}
