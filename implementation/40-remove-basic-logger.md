# Plan to Remove Basic Logger Interface (Feature #40)

## Goal

Remove the basic `Logger` interface, keeping only `StructuredLogger` and its implementations.

## Dependencies

- Depends on: Feature #38 (JSON Encoding) - should be implemented first
- After implementation, plan #41 (Simplify Tests) may need updates if test interfaces change

## Implementation Order

This is the second feature to implement:
1. **First**: Feature #38 (JSON Encoding)
2. **Second**: Feature #40 (Remove Basic Logger) ← you are here
3. **Third**: Feature #41 (Simplify Tests)

## Current State

- `Logger` interface has `Print(v any)` and `Printf(format string, v ...any)`
- `StructuredLogger` has level-based methods: `Log`, `Debug`, `Info`, `Warn`, `Error`
- `StructuredToBasic` wraps `StructuredLogger` to satisfy `Logger`

## Usage Analysis

| File | Current Type | Notes |
|------|--------------|-------|
| `server/server.go` | `log.Logger` | Server accepts logger |
| `check/compilation.go` | `log.Logger` | CompilationUnit |
| `check/check.go` | `log.Logger` | Checker |
| `server/log.go` | `log.Logger` | Returns old loggers (lineLogger, etc.) |
| `cmd/lsp/lsp.go` | Uses wrapper | Wraps StructuredLogger |

## Implementation Plan

### Step 1: Update `server/server.go`

- Change `logger grammar_log.Logger` to `logger grammar_log.StructuredLogger`

### Step 2: Update `check/compilation.go`

- Change `log log.Logger` to `log log.StructuredLogger`

### Step 3: Update `check/check.go`

- Change `log log.Logger` to `log log.StructuredLogger`

### Step 4: Update `cmd/lsp/lsp.go`

- Remove `NewStructuredToBasic` wrapper, pass `StructuredLogger` directly

### Step 5: Remove deprecated code from `server/log.go`

- `NewSilentLogger` - returns old Logger interface
- `NewWriterLogger` - returns old Logger interface
- `NewTestLogger` - returns old Logger interface
- `NewLineLogger` - returns old Logger interface
- Either remove these functions or refactor to return `StructuredLogger`

### Step 6: Remove from `log/logger.go`

- Delete `Logger` interface
- Delete `stderrLogger` struct and methods
- Delete `silentLogger` struct and methods
- Delete `StructuredToBasic` wrapper and `NewStructuredToBasic` function

### Step 7: Update tests

- Remove tests for basic Logger interface
- Update server tests to use `StructuredLogger` directly
- Remove `server/log_test.go` if it tests deprecated functionality

**Note**: Line numbers in this plan may change. Refer to the type signatures (`grammar_log.Logger` vs `grammar_log.StructuredLogger`) rather than line numbers.

## Benefits

- Single unified logging interface
- All logging is structured with levels and fields
- Simpler codebase with no duplicate interfaces
- Consistent logging across the entire codebase
