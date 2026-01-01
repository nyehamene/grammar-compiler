package token

import (
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
			input:   `/\n\t\d\s\(\)\[\]\{\}\\.\//`,
			valid:   true,
			wantVal: `/\n\t\d\s\(\)\[\]\{\}\\.\//`,
		},
		{
			name:    "valid trailing escape",
			input:   `/abc\\/`,
			valid:   true,
			wantVal: `/abc\\/`,
		},
		{
			name:    "valid star plus question",
			input:   `/a\*b\+c\?/`,
			valid:   true,
			wantVal: `/a\*b\+c\?/`,
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

func TestScanString(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		valid   bool
		wantVal string
	}{
		{
			name:    "valid escaped quote",
			input:   `"foo\"bar"`,
			valid:   true,
			wantVal: `"foo\"bar"`,
		},
		{
			name:    "valid other escapes",
			input:   `"hello\n\t\r\\world"`,
			valid:   true,
			wantVal: `"hello\n\t\r\\world"`,
		},
		{
			name:    "valid escaped single quote",
			input:   `"It\'s fine"`,
			valid:   true,
			wantVal: `"It\'s fine"`,
		},
		{
			name:    "valid trailing escape",
			input:   `"final\\"`,
			valid:   true,
			wantVal: `"final\\"`,
		},
		{
			name:    "invalid escape sequence",
			input:   `"\a"`,
			valid:   false,
			wantVal: `"\a"`,
		},
		{
			name:    "unterminated string",
			input:   `"abc`,
			valid:   false,
			wantVal: `"abc`,
		},
		{
			name:    "unterminated with escape",
			input:   `"abc\"`,
			valid:   false,
			wantVal: `"abc\"`,
		},
		{
			name:    "string with newline",
			input:   "\"abc\ndef\"",
			valid:   false,
			wantVal: `"abc`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcRunes := []rune(tc.input)
			tokenizer := NewTokenizer(srcRunes, false, false)
			tokens := tokenizer.Scan()

			if len(tokens) == 0 {
				t.Fatalf("Expected at least 1 token, got 0")
			}

			stringToken := tokens[0]
			if stringToken.Kind != String {
				t.Fatalf("Expected first token to be STRING, got %s", stringToken.Kind)
			}

			if (stringToken.State == Valid) != tc.valid {
				t.Errorf("Expected token validity to be %t, but got %t", tc.valid, (stringToken.State == Valid))
			}

			gotVal := Literal(stringToken, srcRunes)
			if gotVal != tc.wantVal {
				t.Errorf("Expected token value to be %q, got %q", tc.wantVal, gotVal)
			}
		})
	}
}

func TestScanExternal(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		valid   bool // For external values, it's always valid at tokenizer level if it starts with $
		wantVal string
	}{
		{
			name:    "simple external value",
			input:   "$foo",
			valid:   true,
			wantVal: "$foo",
		},
		{
			name:    "external value with underscore",
			input:   "$bar_baz",
			valid:   true,
			wantVal: "$bar_baz",
		},
		{
			name:    "external value with digits",
			input:   "$qux123",
			valid:   true,
			wantVal: "$qux123",
		},
		{
			name:    "just dollar sign (parser will error)",
			input:   "$",
			valid:   true, // Tokenizer still considers it valid External kind
			wantVal: "$",
		},
		{
			name:    "dollar sign followed by non-ident char",
			input:   "$1foo",
			valid:   true, // Tokenizer produces $1, parser will likely error on '1' being invalid ident start
			wantVal: "$1foo",
		},
		{
			name:    "external value followed by space",
			input:   "$foo ",
			valid:   true,
			wantVal: "$foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcRunes := []rune(tc.input)
			tokenizer := NewTokenizer(srcRunes, false, false)
			tokens := tokenizer.Scan()

			if len(tokens) == 0 {
				t.Fatalf("Expected at least 1 token, got 0")
			}

			externalToken := tokens[0]
			if externalToken.Kind != External {
				t.Fatalf("Expected first token to be EXTERNAL, got %s", externalToken.Kind)
			}

			if (externalToken.State == Valid) != tc.valid {
				t.Errorf("Expected token validity to be %t, but got %t", tc.valid, (externalToken.State == Valid))
			}

			gotVal := Literal(externalToken, srcRunes)
			if gotVal != tc.wantVal {
				t.Errorf("Expected token value to be %q, got %q", tc.wantVal, gotVal)
			}
		})
	}
}
