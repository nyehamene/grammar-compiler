# AGENTS.md - Guide for AI Agents

This document provides guidance for AI agents working on this codebase.

## Project Overview

**Grammar** is a domain-specific language (DSL) for describing programming language syntax. It provides a CLI tool and LSP server for grammar file validation, formatting, and IDE support.

- **Language**: Go 1.25+
- **Module**: `grammar`
- **License**: MIT

## Common Tasks

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with timeout
make test

# Update snapshots
UPDATE_SNAPSHOTS=true go test ./...
```

### Building

```bash
# Build the binary
go build -o grammar .

# Or use make
make build
```

### Linting

```bash
golangci-lint run --timeout 5m
```

## Code Organization

```
.
├── cmd/              # CLI command implementations
│   ├── check/       # Validation command
│   ├── fmt/         # Formatter command
│   ├── lsp/         # Language server command
│   ├── diff/        # Diff command
│   └── print/       # Print tokens/AST command
├── ast/             # Abstract Syntax Tree (parser, formatter)
├── check/           # Semantic checker
├── token/           # Tokenizer
├── log/             # Structured logging
├── server/          # LSP implementation
├── command/         # CLI help text (embedded)
└── testdata/        # Test fixtures
```

## Testing Conventions

### Test File Naming

- Tests colocated with implementation: `*_test.go`
- Snapshot tests: Use `testutil.AssertSnapshotJSON` and `testutil.AssertSnapshotText`

### Snapshot Testing

```go
import "grammar/testutil"

// For JSON snapshots
testutil.AssertSnapshotJSON(t, "test_name", actualResult)

// For text snapshots  
testutil.AssertSnapshotText(t, "test_name", actualText)

// To update snapshots
UPDATE_SNAPSHOTS=true go test -run TestName ./...
```

### Inline Grammar Tests

Use `testutil.ParseGrammar` for inline grammar content in tests:

```go
import "grammar/testutil"

content := `rule_a = "a";`
sources := testutil.ParseGrammar(t, content)
```

## Key Conventions

### Error Handling

- Parser errors: Use `ast.ErrorList` for multiple errors
- Return error types, not strings

### Logging

- Use structured logging from `grammar/log` package
- Levels: DEBUG, INFO, WARN, ERROR

### Git Workflow

- Commit related changes together
- Use meaningful commit messages
- Don't auto-commit (per user request)

## Implementation Plans

Implementation plans are in `implementation/` directory, numbered 01-42.
- Feature #39+: Snapshot tests use `testutil/snapshot.go`
- Check implementation plans for test requirements before implementing

## CLI Commands

| Command | Description |
|---------|-------------|
| `grammar check` | Validate grammar files |
| `grammar fmt` | Format grammar files |
| `grammar diff` | Compare files (--tokens, --ast) |
| `grammar print` | Print tokens/AST (--token, --ast) |
| `grammar lsp` | Start language server |

## Common Issues

### Snapshot Test Failures

If snapshot tests fail after intentional output changes:
```bash
UPDATE_SNAPSHOTS=true go test ./...
```

### Test Data Formatting

Don't commit whitespace-only changes to `.grammar` test files - run formatter first:
```bash
grammar fmt testdata/packages/basic/A.grammar
```
