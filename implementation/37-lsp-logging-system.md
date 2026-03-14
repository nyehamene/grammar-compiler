# LSP Logging System Enhancement Plan

## Goal

Implement a robust logging system for the LSP server with:
1. Optional `--log <filename>` CLI option
2. Structured JSON logging format
3. Comprehensive request/response/notification logging

## Current State

- Logger interface in `log/logger.go` with `Print` and `Printf` methods
- Server uses `lineLogger` in `server/log.go` for human-readable logs
- LSP command writes to `~/.cache/grammar/lsp.log` (hardcoded)

## Implementation Plan

### 1. Update Logger Interface (`log/logger.go`)

- [ ] Add structured logging levels: `DEBUG`, `INFO`, `WARN`, `ERROR`
- [ ] Add JSON formatting support with structured log entries
- [ ] Add time stamping to log entries

```go
type Level int
const (
    DEBUG Level = iota
    INFO
    WARN
    ERROR
)

type StructuredLogger interface {
    Log(level Level, msg string, fields map[string]any)
    Debug(msg string, fields map[string]any)
    Info(msg string, fields map[string]any)
    Warn(msg string, fields map[string]any)
    Error(msg string, fields map[string]any)
}
```

### 2. Create Structured Logger Implementations (`log/structured.go`)

- [ ] **JSONLogger**: Outputs logs as JSON lines
- [ ] **ConsoleLogger**: Human-readable colored output
- [ ] **MultiLogger**: Writes to multiple outputs

### 3. Update LSP Command (`cmd/lsp/lsp.go`)

- [ ] Add `--log <file>` flag
- [ ] Add `--log-format <json|text>` flag (default: json when --log specified)
- [ ] Add `--log-level <debug|info|warn|error>` flag (default: debug)
- [ ] Handle flag errors gracefully

### 4. Update Server Logging (`server/server.go`)

- [ ] Pass structured logger to server
- [ ] Log every request with: method, params, correlation ID
- [ ] Log every response with: id, result/error, timing
- [ ] Log every notification
- [ ] Add correlation IDs for request/response matching

### 5. Update All LSP Handlers

- [ ] Add structured logging to all handle* methods
- [ ] Include document URI, position, and other relevant context
- [ ] Log timing information (optional)

### 6. JSON Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00.123Z",
  "level": "info",
  "message": "textDocument/definition request",
  "method": "textDocument/definition",
  "request_id": 1,
  "document_uri": "file:///project/foo.grammar",
  "position": {"line": 10, "character": 5},
  "response_time_ms": 15
}
```

### 7. JSON Pretty Printing

- [ ] Add `--pretty` flag for human-readable JSON output
- [ ] When enabled, JSON output is formatted with indentation
- [ ] Default: compact JSON (single line per entry)
- [ ] Useful for debugging with tools like `jq` or when reading logs directly

```bash
# Compact JSON (default)
grammar-lsp --log /tmp/lsp.log --pretty=false
cat /tmp/lsp.log
# {"timestamp":"...","level":"info","message":"..."}

# Pretty-printed JSON
grammar-lsp --log /tmp/lsp.log --pretty
cat /tmp/lsp.log
# {
#   "timestamp": "...",
#   "level": "info",
#   "message": "..."
# }
```

### 8. Combined Flag Options

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--log` | filepath | (none) | Enable file logging |
| `--log-format` | json, text | json* | Output format |
| `--log-level` | debug, info, warn, error | debug* | Minimum log level |
| `--pretty` | true, false | false | Pretty-print JSON |

*Defaults apply when `--log` is specified

## File Changes

| File                | Changes                                          |
|---------------------|---------------------------------------------------
| `log/logger.go`     | Add Level enum, StructuredLogger interface, pretty option |
| `log/structured.go` | New file: JSONLogger, ConsoleLogger, MultiLogger |
| `cmd/lsp/lsp.go`    | Add --log, --log-format, --log-level, --pretty flags       |
| `server/server.go`  | Update to use structured logger                  |
| `server/log.go`     | Update to use structured logging                 |
| `server/*.go`       | Update handlers to use structured logging        |
|---------------------|---------------------------------------------------

## Backward Compatibility

- Default to current behavior (text logging to stderr)
- If no --log flag, use stderr with text format and info level
- If --log specified:
  - Default format: JSON
  - Default level: DEBUG
- JSON format provides full request/response details for debugging

## Test Plan

### 1. Logger Interface Tests (`log/logger_test.go`)

#### 1.1 Level Tests
- [ ] Test Level enum values (DEBUG=0, INFO=1, WARN=2, ERROR=3)
- [ ] Test Level String() method returns correct string
- [ ] Test Level parsing from string

#### 1.2 JSONLogger Tests
- [ ] Test JSONLogger outputs valid JSON lines
- [ ] Test JSONLogger includes timestamp field
- [ ] Test JSONLogger includes level field
- [ ] Test JSONLogger includes message field
- [ ] Test JSONLogger includes custom fields
- [ ] Test JSONLogger handles special characters in fields

#### 1.3 ConsoleLogger Tests
- [ ] Test ConsoleLogger outputs human-readable format
- [ ] Test ConsoleLogger includes timestamp
- [ ] Test ConsoleLogger includes level
- [ ] Test ConsoleLogger includes message

#### 1.4 MultiLogger Tests
- [ ] Test MultiLogger writes to all outputs
- [ ] Test MultiLogger handles nil writers gracefully

---

### 2. CLI Flag Tests (`cmd/lsp/lsp_test.go`)

#### 2.1 Log Flag Tests
- [ ] Test `--log <file>` creates log file
- [ ] Test `--log` with valid file path
- [ ] Test `--log` with invalid file path (should error)
- [ ] Test `--log` with directory path (should error)

#### 2.2 Log Format Flag Tests
- [ ] Test `--log-format json` sets JSON format
- [ ] Test `--log-format text` sets text format
- [ ] Test `--log-format` default is json when --log specified
- [ ] Test `--log-format` with invalid value (should error)

#### 2.3 Log Level Flag Tests
- [ ] Test `--log-level debug` sets DEBUG level
- [ ] Test `--log-level info` sets INFO level
- [ ] Test `--log-level warn` sets WARN level
- [ ] Test `--log-level error` sets ERROR level
- [ ] Test `--log-level` with invalid value (should error)

#### 2.4 Pretty Flag Tests
- [ ] Test `--pretty` flag is recognized
- [ ] Test `--pretty=true` enables pretty printing
- [ ] Test `--pretty=false` disables pretty printing
- [ ] Test default is false (compact JSON)

#### 2.5 Flag Combinations
- [ ] Test `--log --log-format json --log-level debug`
- [ ] Test `--log` alone (defaults to json, debug)
- [ ] Test without --log (uses stderr, text, info)

---

### 3. Server Logging Tests (`server/log_test.go`)

#### 3.1 Request Logging Tests
- [ ] Test request logging includes method name
- [ ] Test request logging includes request ID
- [ ] Test request logging includes timestamp
- [ ] Test notification logging (no ID)

#### 3.2 Response Logging Tests
- [ ] Test response logging includes request ID
- [ ] Test response logging includes result
- [ ] Test response logging includes error (if any)
- [ ] Test response logging includes timing

#### 3.3 Structured Fields Tests
- [ ] Test document URI is logged for textDocument methods
- [ ] Test position is logged for position-related methods
- [ ] Test context fields are logged

---

### 4. Integration Tests (`server/logging_integration_test.go`)

#### 4.1 Full Request/Response Cycle
- [ ] Test initialize request/response logged
- [ ] Test textDocument/didOpen notification logged
- [ ] Test textDocument/definition request logged with timing

#### 4.2 Log File Tests
- [ ] Test log file contains JSON lines
- [ ] Test log file can be parsed
- [ ] Test log file grows with multiple requests

#### 4.3 Backward Compatibility
- [ ] Test default behavior (no flags) uses stderr
- [ ] Test default level is INFO
- [ ] Test --log flag enables DEBUG level by default
- [ ] Test --pretty defaults to false (compact JSON)

#### 4.4 Pretty Print Tests
- [ ] Test JSON output is compact by default
- [ ] Test --pretty produces indented JSON
- [ ] Test pretty output is valid JSON when parsed

---

## Running Tests

```bash
# Run logger tests
go test -v ./log/... -run TestLevel

# Run CLI tests
go test -v ./cmd/lsp/... -run TestLog

# Run server logging tests
go test -v ./server/... -run TestLogging

# Run all logging tests
go test -v ./... -run "Log"
```
