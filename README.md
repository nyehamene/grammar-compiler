# Grammar

A language for describing programming language syntax.

## Syntax

A grammar is composed of 3 kinds of declaration.

1. Comment:

  A comment begins with 2 forward-slash (/) and is terminated by a newline.

2. Binding

  A binding imports the content of a grammar file and binds it to an identifier
  that can be used to reference declarations in the imported file.

  A binding declaration is terminated by a semicolon.

3. Rule

  A rule is the main component of a grammar file. It is composed from 2 kinds of
  values.

  - Terminal

    A terminal is one of the following:

    1) A string literal. Any value surrounded by double quotes.
    2) A regular expression. Any value surrounded  by forward slashes.

  - Non-terminal

    A non-terminal is one of the following:

    1) An identify that reference a rule declared in the current file.
    2) A member access on a binding variable that references a rule declared
       in an imported file.

  A rule declaration is terminated by a semicolon.

See the content of ./grammar.txt for a description of the syntax the grammar.

See files ./example directory for an example of a templating language
written in this grammar.

## Implementation

The project will implement a command line tool that implements
the following features.

* A __checker__ for validating syntax and semantics
* A __formatter__ for formatting grammar files
* A __language server__ for the grammar language

See the content of ./command directory for the structure and format of
the commandline arguments accepted by the tool.

### Implementation - Phases

The content of ./implementation directory will guide the implementation
process and describes what deliverables for each stage. Stages are numbered
in sequence starting from 1 to N.

### Implementation - Language

The program will be implemented in __Go Programming Langauge__.
But produce a executable that functions as described in ./command/grammar.txt.

The implementation should following the following project structure.

```sh
cmd/fmt/fmt.go
cmd/check/check.go
cmd/lsp/lsp.go
cmd/diff/diff.go
cmd/print/print.go
cmd/cmd.go
token/token.go
token/tokenizer.go
ast/ast.go
ast/parser.go
main.go
go.mod
```

### Stages
- [x] Skeleton
- [x] Version
- [x] Help
- [x] Tokenizer
- [x] Print token
- [ ] Diff token
- [ ] Parser
- [ ] Print ast
- [ ] Diff ast
- [ ] Formatter
- [ ] Language server
- [ ] Treesitter
- [ ] Print token (color)

