# Grammar

[![Go Version](https://img.shields.io/github/go-mod/go-version/better-tools/better-grammar)](https://golang.org/doc/install)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A domain-specific language for describing programming language syntax.

## Overview

Grammar is a tool and DSL for defining the syntax of programming languages. It provides:

- **Checker** - Validate grammar files and report diagnostic errors
- **Formatter** - Format grammar files with configurable style
- **Diff** - Compare grammar files at token or AST level
- **Print** - Output tokens or AST nodes for debugging
- **Language Server** - IDE support via LSP (completion, hover, go-to-definition, etc.)

## Installation

### From Source

```bash
go install .
```

### Build from Repository

```bash
git clone https://github.com/better-tools/better-grammar.git
cd better-grammar
go build -o grammar .
```

## CLI Interface

```
Usage: grammar COMMAND

A tool for describing the syntax of programming languages.

Command:
fmt      Format grammar files inplace.
check    Validate grammar files and report diagnostic errors.
print    Print tokens or Ast nodes.
diff     Compare 2 input files.
lsp      Start a language server.
help     Print this message.
version  Print version information.

An grammar file is any file with the file extension .grammar.
```

## Quick Start

```bash
# Check a grammar file
grammar check mylang.grammar

# Format a grammar file
grammar fmt mylang.grammar

# Start the language server
grammar lsp

# Compare two files
grammar diff --ast file1.grammar file2.grammar
```

## Grammar Syntax

A grammar file (`.grammar`) defines language syntax using rules and bindings.

### Rules

```grammar
// A rule defines a named production
greeting = "Hello" | "Hi";

// Rules can reference other rules
full_greeting = greeting " World";
```

### Bindings (Imports)

```grammar
// Import another grammar file
mylang = @import("base.grammar");

// Reference rules from imported files
base_rule = mylang.identifier;
```

### Expressions

```grammar
// String literals
str = "hello";

// Regular expressions
regex = /[a-z]+/;

// Groups
grouped = ( "a" "b" );

// Repetition: zero more
repeated = { "a" };

// Optional
optional = [ "b" ];

// External rule
external = $external | <external>;
```

### Directives

```grammar
// Package declaration (for directory-as-package)
@package("mypackage");

// Import from package directory
pkg = @import("mylib");
```

## Commands

### `grammar check`

Validate grammar files and report errors:

```bash
grammar check path/to/file.grammar
grammar check path/to/directory/
```

### `grammar fmt`

Format grammar files:

```bash
grammar fmt path/to/file.grammar     # Format in place
grammar fmt --stdout file.grammar    # Output to stdout
grammar fmt --stdin                  # Format from stdin
```

### `grammar diff`

Compare grammar files:

```bash
grammar diff --tokens file1.grammar file2.grammar
grammar diff --ast file1.grammar file2.grammar
```

### `grammar print`

Print tokens or AST for debugging:

```bash
grammar print --token file.grammar
grammar print --ast file.grammar
```

### `grammar lsp`

Start the Language Server Protocol server for IDE integration:

```bash
grammar lsp                         # Default (no logging)
grammar lsp --log server.log        # Log to file
grammar lsp --log-format json       # JSON log format
grammar lsp --log-level info        # Set log level
```

## Language Server Features

The Grammar language server provides IDE support including:

- **Completion** - Auto-complete rule names, imports, and package members
- **Hover** - Show type information on hover
- **Go to Definition** - Navigate to rule definitions
- **Find References** - Find all usages of a rule
- **Document Symbols** - Outline view of grammar file
- **Diagnostics** - Real-time error checking
- **Document Links** - Navigate between imported files
- **Formatting** - Document formatting support

## Project Structure

```
.
├── cmd/             # CLI commands
│   ├── check/       # Validation command
│   ├── fmt/         # Formatter command
│   ├── lsp/         # Language server
│   ├── diff/        # Diff command
│   └── print/       # Print command
├── ast/             # Abstract Syntax Tree
├── check/           # Semantic checking
├── token/           # Tokenizer
├── log/             # Logging utilities
├── server/          # LSP implementation
└── command/         # CLI help text
```

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on the [GitHub repository](https://github.com/better-tools/better-grammar).
