# Project Overview

This project is a Go-based command-line tool for working with a custom grammar definition language. The language is designed to describe the syntax of programming languages in a modular and reusable way. The tool provides functionalities for checking, formatting, and inspecting grammar files, as well as a Language Server Protocol (LSP) implementation for editor support.

The grammar language itself is composed of three main declaration types:
*   **Comments**: Standard single-line comments.
*   **Bindings**: A mechanism to import and use other grammar files.
*   **Rules**: The core of the grammar, defining terminals (string literals and regular expressions) and non-terminals (references to other rules).

The project is structured with a clear separation of concerns, including directories for command-line interface definitions, example grammars, and a phased implementation guide.

# Building and Running

While there are no explicit build scripts like `Makefile` or `go.mod` in the current directory, the `README.md` file indicates that this is a Go project. The intended project structure suggests that the main entry point will be `main.go`.

**TODO:** Add explicit build, run, and test commands once the Go source files are created. A typical Go project would use the following commands:

*   **Build:** `go build`
*   **Run:** `go run main.go`
*   **Test:** `go test ./...`

# Development Conventions

The project follows a phased development approach, as outlined in the `implementation` directory. Each markdown file in this directory corresponds to a specific stage of development, starting from the basic command-line skeleton to the full implementation of the language server.

The code is intended to be written in Go, following the project structure described in `README.md`:

```sh
cmd/fmt/fmt.go
cmd/check/check.go
cmd/lsp/lsp.go
cmd/cmd.go
token/token.go
token/tokenizer.go
ast/ast.go
ast/parser.go
main.go
go.mod
```

The example grammar files in the `example` directory demonstrate a modular approach to language design, where complex grammars are composed from smaller, more focused grammar files.
