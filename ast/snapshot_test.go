package ast

import (
	"grammar/testutil"
	"grammar/token"
	"testing"
)

func TestParserSnapshot(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "parser_simple_rules",
			input: `
a = "a";
b = "b";
`,
		},
		{
			name: "parser_imports",
			input: `
a = @import("foo.grammar");
b = a.rule;
`,
		},
		{
			name: "parser_alternatives",
			input: `
foo = "a" | "b" | "c";
`,
		},
		{
			name: "parser_groups",
			input: `
foo = ( "a" "b" );
bar = ( "x" | "y" ) "z";
`,
		},
		{
			name: "parser_import_predicate",
			input: `
foo = @import("test.grammar");
`,
		},
		{
			name: "parser_rule_reference",
			input: `
foo = bar;
bar = "b";
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srcRunes := []rune(tc.input)
			tokenizer := token.NewTokenizer(srcRunes, false, false)
			tokens := tokenizer.Scan()
			parser := NewParser(tokens, srcRunes)
			file, err := parser.ParseFile()

			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			testutil.AssertSnapshotJSON(t, tc.name, file)
		})
	}
}

func TestFormatterSnapshot(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "formatter_simple_rules",
			input: `a = "a";b = "b";`,
		},
		{
			name:  "formatter_alternatives_with_spaces",
			input: `foo = a|b|c;`,
		},
		{
			name:  "formatter_groups",
			input: `foo = (a b);bar = (x|y)z;`,
		},
		{
			name:  "formatter_alignment",
			input: `a = "a";longname = "b";`,
		},
		{
			name:  "formatter_imports",
			input: `a = @import("foo.grammar");`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srcRunes := []rune(tc.input)
			tokenizer := token.NewTokenizer(srcRunes, false, false)
			tokens := tokenizer.Scan()

			formatterParser := NewFormatterParser(tokens, srcRunes)
			formatFile, err := formatterParser.Parse()
			if err != nil {
				t.Fatalf("Format error: %v", err)
			}

			formatter := NewFormatter(formatFile)
			formatted := formatter.Format()

			testutil.AssertSnapshotText(t, tc.name, formatted)
		})
	}
}
