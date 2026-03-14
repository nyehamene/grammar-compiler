# Plan to Simplify Tests and Reduce .grammar File Dependencies (Feature #41)

## Goal

Reduce the number of .grammar files needed in tests while maintaining test coverage. This plan depends on implementing snapshot tests (feature #39) to catch regressions.

## Dependencies

- **Prerequisite**: Feature #39 (Snapshot Tests) - Must be implemented first
- **Implementation order**:
  1. Feature #38 (JSON Encoding) - implemented first
  2. Feature #40 (Remove Basic Logger) - implemented second
  3. Feature #41 (Simplify Tests) - implemented third ← you are here
- May need updates after features #38 and #40 are implemented if they change test interfaces or logging behavior

## Current State

- **34 .grammar files** across testdata directory
- Many files are unique to specific tests
- Some tests use inline content via `newDidOpenNotification` (server tests)
- Some tests require actual files on disk (package resolution, imports)

## Problems

1. **Too many test files** - 34 .grammar files hard to manage
2. **Duplication** - Similar grammar content repeated across files
3. **Hard to discover** - Which files are used by which tests
4. **File-based tests slower** - File I/O adds overhead vs inline content

## Solution: Test Data Tiering

### Tier 1: Inline Content (Preferred)
- Grammar content defined directly in test code
- Used via `newDidOpenNotification` for server tests
- Used via `LoadSource([]byte(...), path)` for checker tests
- **Goal**: 80% of tests

### Tier 2: Minimal Shared Fixtures
- Small set of reusable .grammar files
- Shared across multiple tests
- Well-documented purpose

### Tier 3: Special-Purpose Files
- Only for features requiring actual files (package resolution, directory-as-package)
- Kept to minimum

## Implementation Plan

### Phase 1: Add Snapshot Tests (Prerequisite)

Implement feature #39 to capture outputs. This allows refactoring with confidence.

### Phase 2: Audit Current Tests

Categorize tests:

| Category | Current Approach | Target Approach |
|----------|------------------|------------------|
| Parser tests | File-based | Inline |
| Checker tests | File-based | Inline for simple cases |
| Server LSP tests | Mostly inline | Keep inline |
| Package tests | File-based | Reduce to essential |
| Import tests | File-based | Inline where possible |

### Phase 3: Create Inline Test Helpers

Add helper functions to reduce boilerplate:

```go
// In test helpers
func Grammar(content string) []byte {
    return []byte(content)
}

func OpenGrammar(h *Harness, uri, content string) {
    h.send(newDidOpenNotification(uri, content, 1))
}

// Usage in tests
OpenGrammar(h, "file:///test.grammar", `
rule_a = "hello";
rule_b = rule_a;
`)
```

### Phase 4: Consolidate Shared Fixtures

Create minimal set of shared files:

```
testdata/
  fixtures/
    simple.grammar       # Single rule
    multiple.grammar     # Multiple rules
    import.grammar      # For import tests
    package.grammar     # For package tests
```

### Phase 5: Refactor Tests

Move from file-based to inline where possible:

**Before:**
```go
checker.Check("../testdata/check/success/a.grammar")
```

**After:**
```go
source := `
a = "hello";
b = a;
`
cu.LoadSource([]byte(source), "test.grammar")
checker.Check("test.grammar")
```

### Phase 6: Remove Unused Files

After refactoring, remove unused .grammar files.

## Test Helper Functions to Add

```go
// testutil/grammar.go
package testutil

// InlineGrammar creates grammar content for tests
func InlineGrammar(rules string) []byte {
    return []byte(rules)
}

// WithImport creates grammar with import statement
func WithImport(pkg, content string) string {
    return pkg + ` = @import("` + content + `");`
}

// Rule creates a rule definition
func Rule(name, body string) string {
    return name + " = " + body + ";"
}
```

## File Reduction Targets

| Directory | Current Files | Target | Reduction |
|-----------|--------------|--------|-----------|
| testdata/check/ | 14 | 4 | 71% |
| testdata/parser/ | 6 | 2 | 67% |
| testdata/packages/ | 6 | 4 | 33% |
| testdata/lsp/ | 4 | 2 | 50% |
| testdata/ | 4 | 2 | 50% |

## Workflow

1. **Add snapshot tests** (feature #39)
2. **Run baseline tests** - ensure all pass
3. **Refactor one test at a time** - convert to inline
4. **Run tests after each change** - verify no regression
5. **Update snapshots** if output changes
6. **Remove unused files** after confirming no usage

## Risks & Mitigations

| Risk | Mitigation |
|------|-------------|
| Regression from inline changes | Snapshot tests (#39) catch output differences |
| Tests requiring real files | Keep essential files, document why |
| Breaking package resolution tests | Test with temp directories |
| Feature #38 changes test interfaces | Update tests after #38 is implemented |
| Feature #41 changes logging behavior | Verify logging output in snapshots after #41 |

## Success Criteria

- [ ] Reduce .grammar files from 34 to ~15
- [ ] 80% of tests use inline content
- [ ] All tests pass with snapshot verification
- [ ] Clear documentation of shared fixtures
