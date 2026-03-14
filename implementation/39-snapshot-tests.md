# Plan to Add Snapshot Tests (Feature #39)

## Goal

Add snapshot tests to prevent regressions in LSP responses, diagnostics, and other output.

## Dependencies

- This plan is a prerequisite for plan #41 (Simplify Tests)
- This plan can be implemented independently of plans #38 and #40

## Implementation Order

The key constraint is: **#39 must be implemented before #41**

The order relative to #38 and #40 is flexible:
- Option A: 38 → 39 → 40 → 41
- Option B: 38 → 40 → 39 → 41  
- Option C: 40 → 38 → 39 → 41

Any order works as long as #39 comes before #41.

## What are Snapshot Tests?

Snapshot tests capture the output of a function/test and save it to a file. Future test runs compare against the saved snapshot. If output changes, the test fails, alerting developers to unexpected changes.

## Snapshot Testing Library

Use Go's built-in testing approach or [github.com/hexdigest/gowrap](https://github.com/hexdigest/gowrap) for quick implementation. Alternatively, use [github.com/stretchr/testify/snapshot](https://github.com/stretchr/testify).

## Areas to Add Snapshot Tests

### 1. LSP Responses

| Test Area | Snapshot File | What to Capture | Current Status |
|-----------|---------------|------------------|----------------|
| Completion responses | `testdata/snapshots/completion/*.json` | CompletionItem arrays | ✅ COMPLETE |
| Hover responses | `testdata/snapshots/hover/*.json` | Hover content | ✅ COMPLETE |
| Definition responses | `testdata/snapshots/definition/*.json` | Location/LocationLink | ✅ COMPLETE |
| References responses | `testdata/snapshots/references/*.json` | Location arrays | ✅ COMPLETE |
| Diagnostics | `testdata/snapshots/diagnostics/*.json` | Diagnostic arrays | ✅ COMPLETE |
| Document symbols | `testdata/snapshots/documentSymbol/*.json` | SymbolInformation arrays | ✅ COMPLETE |

### 2. Log Output

| Test Area | Snapshot File | What to Capture | Current Status |
|-----------|---------------|------------------|----------------|
| JSON Logger output | `testdata/snapshots/log/json_*.jsonl` | JSON log lines | ✅ COMPLETE |
| Console Logger output | `testdata/snapshots/log/console_*.txt` | Formatted log text | ✅ COMPLETE |

### 3. Parsing & AST

| Test Area | Snapshot File | What to Capture | Current Status |
|-----------|---------------|------------------|----------------|
| Parser output | `testdata/snapshots/parser/*.txt` | AST representation | ✅ COMPLETE |
| Formatter output | `testdata/snapshots/formatter/*.grammar` | Formatted source | ✅ COMPLETE |

### 4. CLI Output

| Test Area | Snapshot File | What to Capture | Current Status |
|-----------|---------------|------------------|----------------|
| LSP help text | `testdata/snapshots/cli/help.txt` | Help output | ✅ COMPLETE |
| Error messages | `testdata/snapshots/cli/errors.txt` | Error output | ✅ COMPLETE |

## Implementation Plan

### Step 1: Create Snapshot Directory Structure

```
testdata/
  snapshots/
    completion/
    hover/
    definition/
    references/
    diagnostics/
    documentSymbol/
    log/
    parser/
    formatter/
    cli/
```

### Step 2: Create Snapshot Helper

Create `testutil/snapshot.go`:

```go
package testutil

import (
    "os"
    "path/filepath"
    "testing"
)

func AssertSnapshot(t *testing.T, name string, got any) {
    // Use testify/snapshot or implement custom logic
}
```

### Step 3: Add Snapshot Tests for LSP Handlers

Modify existing tests or create new snapshot tests:

```go
func TestCompletionSnapshot(t *testing.T) {
    // ... setup ...
    
    resp := h.send(completionRequest)
    testutil.AssertSnapshot(t, "completion/basic", resp)
}
```

### Step 4: Run Tests with Update Flag

```bash
# Update snapshots
go test -update ./...

# Run normally (will fail if output changed)
go test ./...
```

### Step 5: Add to CI

```yaml
# .github/workflows/test.yml
- name: Run snapshot tests
  run: go test ./...

# Ensure snapshots are up to date
- name: Verify snapshots
  run: go test -verify ./...
```

## Benefits

1. **Catch regressions early** - Any unexpected output change fails the test
2. **Document expected behavior** - Snapshots serve as living documentation
3. **Ease maintenance** - Update snapshots when behavior intentionally changes
4. **Fast feedback** - No need to manually verify output

## Testing Workflow

1. Write test that captures output
2. Run with `-update` flag to create initial snapshot
3. Review snapshot to ensure correctness
4. Commit snapshot alongside code
5. Future changes that affect output will fail the test

## Considerations

- **Timestamp handling**: Use fixed timestamps in tests or normalize in assertions
- **Dynamic data**: Use deterministic values (e.g., sequential IDs instead of random)
- **Large outputs**: Consider splitting into smaller snapshots
- **Binary data**: Skip or use checksums for binary outputs
