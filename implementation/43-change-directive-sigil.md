# Plan: Change Directive Sigil from `@` to `#`

## Overview

Change the directive sigil from `@` to `#` throughout the codebase.

**Current syntax:**
```grammar
@import("file.grammar")
@package("mypackage")
```

**New syntax:**
```grammar
#import("file.grammar")
#package("mypackage")
```

## Files to Modify

### 1. Core Grammar Definition

| File | Change |
|------|--------|
| `grammar.txt` | Change `@import` → `#import`, `@package` → `#package` |

### 2. Tokenizer

| File | Change |
|------|--------|
| `token/token.go` | Update token kinds if named after sigil |
| `token/tokenizer.go` | Change `@` recognition to `#` for directives |

### 3. Parser

| File | Change |
|------|--------|
| `ast/ast.go` | Update DirectiveExpr node handling |
| `ast/parser.go` | Update directive parsing |
| `ast/formatter.go` | Update directive formatting |

### 4. Checker

| File | Change |
|------|--------|
| `check/compilation.go` | Update directive handling |
| `check/check.go` | Update directive checks |

### 5. Server (LSP)

| File | Change |
|------|--------|
| `server/completion.go` | Update completion for `#import` |
| `server/documentlink.go` | Update document link detection |
| `server/hover.go` | Update hover for `#package` |

### 6. Commands

| File | Change |
|------|--------|
| `cmd/fmt/fmt.go` | (if any directive handling) |

### 7. Documentation

| File | Change |
|------|--------|
| `README.md` | Update examples |
| `command/*.txt` | Update help files |

### 8. Tests

| File | Change |
|------|--------|
| All `*_test.go` files | Update `@import`/`@package` to `#import`/`#package` |
| `testdata/**/*.grammar` | Update grammar test files |

## Implementation Steps

### Step 1: Update Grammar Definition
- [ ] `grammar.txt` - Change `@import` → `#import`, `@package` → `#package`

### Step 2: Update Tokenizer
- [ ] `token/tokenizer.go` - Change `@` recognition to `#`
- [ ] `token/tokenizer_test.go` - Update tests

### Step 3: Update Parser
- [ ] `ast/ast.go` - Update directive handling
- [ ] `ast/parser.go` - Update parsing
- [ ] `ast/formatter.go` - Update formatting

### Step 4: Update Checker
- [ ] `check/compilation.go` - Update directive handling
- [ ] `check/check.go` - Update checks

### Step 5: Update Server
- [ ] Update all server handlers that reference directives

### Step 6: Update Tests
- [ ] Update all test files
- [ ] Update all testdata files

### Step 7: Update Documentation
- [ ] `README.md`
- [ ] `command/*.txt`

## Estimated Impact

- **Files to modify**: ~50+ files
- **Test files**: ~20+ test files
- **Testdata files**: ~30+ grammar files

## Backward Compatibility

This is a **breaking change**. Existing `.grammar` files using `@import` and `@package` will need to be updated to use `#import` and `#package`.

## Alternative Approach

Consider supporting both sigils during a transition period:
- Accept both `@` and `#` as valid directive sigils
- Deprecate `@` in favor of `#`
- Remove `@` support in a future version
