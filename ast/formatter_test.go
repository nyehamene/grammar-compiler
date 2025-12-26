package ast

import (
	"grammar/token"
	"testing"
)

func TestFormatter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "Rule Grouping - Alignment of consecutive rules",
			input: `
a = "a";
bar = "bar";

b = "a";
name = "name";
`,
			want: `a   = "a";
bar = "bar";

b    = "a";
name = "name";
`,
		},
		{
			name: "Add space around grouping symbols",
			input: `
one = (foo);
two = [bar];
xxx = {baz};
`,
			want: `one = ( foo );
two = [ bar ];
xxx = { baz };
`,
		},
		{
			name:  "Add space around alternative separator",
			input: `foo = a|b;`,
			want: `foo = a | b;
`,
		},
		{
			name: "Production Group - Alignment of expression",
			input: `
choice = one
   | two
   | three ;
`,
			want: `choice = one
       | two
       | three
       ;
`,
		},
		{
			name: "Trim consecutive blank lines",
			input: `
foo = "one";


bar = "bar";
`,
			want: `foo = "one";

bar = "bar";
`,
		},
		{
			name: "Combined Rule 1 and 4",
			input: `
x = "one";
xxxx = "one"
  | "two"
  | "three";
xx = "two";
`,
			want: `x    = "one";
xxxx = "one"
     | "two"
     | "three"
     ;
xx   = "two";
`,
		},
		{
			name: "Example from implementation notes",
			input: `
/// @var
document = @import("document.grammar");
component = @import("component.grammar");

/// @ast
Source = DocumentNamespace | ComponentNamespace;
DocumentNamespace = document;
ComponentNamespace = component;
`,
			want: `/// @var
document  = @import("document.grammar");
component = @import("component.grammar");

/// @ast
Source             = DocumentNamespace | ComponentNamespace;
DocumentNamespace  = document;
ComponentNamespace = component;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcRunes := []rune(tt.input)
			tokenizer := token.NewTokenizer(srcRunes, false, false) // Do not skip comments or newlines
			tokens := tokenizer.Scan()

			formatterParser := NewFormatterParser(tokens, srcRunes)
			formatFile, err := formatterParser.Parse()
			if err != nil {
				t.Fatalf("FormatterParser error: %v", err)
			}

			formatter := NewFormatter(formatFile)
			got := formatter.Format()

			if got != tt.want {
				t.Errorf("Formatter.Format() mismatch for test %s.\nGot:\n%q\nWant:\n%q", tt.name, got, tt.want)
			}
		})
	}
}
