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

## File Changes

| File                | Changes                                          |
|---------------------|---------------------------------------------------
| `log/logger.go`     | Add Level enum, StructuredLogger interface       |
| `log/structured.go` | New file: JSONLogger, ConsoleLogger, MultiLogger |
| `cmd/lsp/lsp.go`    | Add --log, --log-format, --log-level flags       |
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
