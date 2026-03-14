# Plan to Improve Structured Logger JSON Encoding (Feature #38)

## Goal

Ensure all values logged through the structured logger are properly JSON-encoded, especially custom types like `DocumentUri`, `Position`, `Range`, etc.

## Dependencies

- This plan can be implemented independently first
- After implementation, verify plan #41 (Simplify Tests) still works correctly

## Problem

The current logger uses `json.Marshal` which doesn't always produce the expected output for custom types:

### Example Issues

1. **DocumentUri** - Has no JSON tags, would marshal struct fields instead of string:
   ```go
   // Current behavior:
   logger.Info("request", log.Fields{"document_uri": DocumentUri{Path: "/test.grammar"}})
   // Output: {"document_uri": {"Scheme":"","Authority":"","Path":"/test.grammar",...}}
   
   // Expected:
   // Output: {"document_uri": "file:///test.grammar"}
   ```

2. **Custom types without json.Marshaler** - Would use reflection-based encoding

3. **Error types** - May not serialize meaningfully

## Solution: Option A

Add `MarshalJSON()` methods directly to each type we control. This is the standard Go pattern for custom JSON encoding.

## Implementation Plan

### Step 1: Add MarshalJSON to DocumentUri

This is the primary fix. Modify `server/uri.go`:

```go
func (u DocumentUri) MarshalJSON() ([]byte, error) {
    return json.Marshal(u.String())
}
```

### Step 2: Add error encoder helper

Create `log/encode.go` with a simple wrapper for encoding errors:

```go
package log

// EncodeError converts an error to a string for JSON logging.
func EncodeError(err error) string {
    if err == nil {
        return ""
    }
    return err.Error()
}
```

### Step 3: Update Logger to Use Custom Field Encoding

Modify `log/logger.go` to handle error values:

```go
func (l *JSONLogger) Log(level Level, msg string, fields Fields) {
    // Pre-process fields to ensure proper encoding
    encodedFields := make(Fields)
    for k, v := range fields {
        if err, ok := v.(error); ok {
            encodedFields[k] = EncodeError(err)
        } else {
            encodedFields[k] = v
        }
    }
    // ... rest of logging logic
}
```

### Step 4: Add Tests

Add tests for JSON encoding in respective test files:

```go
// server/uri_test.go
func TestDocumentUriMarshalJSON(t *testing.T) {
    uri, _ := ParseURI("file:///test.grammar")
    data, err := json.Marshal(uri)
    if err != nil {
        t.Fatalf("Failed to marshal: %v", err)
    }
    expected := `"file:///test.grammar"`
    if string(data) != expected {
        t.Errorf("Expected %s, got %s", expected, string(data))
    }
}
```

**Note**: Position, Range, TextEdit, TextDocumentIdentifier, and TextDocumentItem already have JSON tags and work correctly. No changes needed for these types.

## Types to Support

| Type | Location | Current Behavior | Target Output | Notes |
|------|----------|------------------|---------------|-------|
| `DocumentUri` | `server/uri.go` | Struct fields | `"file:///path.grammar"` | **NEEDS FIX** - No MarshalJSON |
| `Position` | `server/types.go` | Works | `{"line": 1, "character": 5}` | Already has JSON tags |
| `Range` | `server/types.go` | Works | `{"start": {...}, "end": {...}}` | Already has JSON tags |
| `TextDocumentIdentifier` | `server/types.go` | Partial | `{"uri": "..."}` | Works via embedded URI MarshalJSON |
| `TextDocumentItem` | `server/types.go` | Partial | `{"uri": "...", ...}` | Works via embedded URI MarshalJSON |
| `TextEdit` | `server/types.go` | Works | `{"range": {...}, "newText": "..."}` | Already has JSON tags |
| `error` | std | Error | `"error message"` | Needs helper function |

## Files to Modify

| File | Changes |
|------|---------|
| `server/uri.go` | Add MarshalJSON for DocumentUri (primary fix needed) |
| `log/logger.go` | Add error encoding in Log method |
| `log/encode.go` | New file with EncodeError helper |
| `server/uri_test.go` | Add tests for DocumentUri MarshalJSON |

**Note**: Position, Range, and TextEdit already work correctly due to existing JSON tags. TextDocumentIdentifier and TextDocumentItem will work automatically once DocumentUri has MarshalJSON.
