# Plan: Change External Value Syntax from `$name` to `<name>`

## Overview

Change the external value syntax from `$ident` to `<ident>` throughout the codebase.

**Current syntax:**
```grammar
external = $foo;
```

**New syntax:**
```grammar
external = <foo>;
```

## Files to Modify

### 1. Core Grammar Definition

| File | Change |
|------|--------|
| `grammar.txt` | Change `$ident` → `<ident>` for external_value |

### 2. Tokenizer

| File | Change |
|------|--------|
| `token/token.go` | Update token kinds if named after sigil |
| `token/tokenizer.go` | Change `$` recognition to `<` for external values |
| `token/tokenizer_test.go` | Update tests |

### 3. Parser

| File | Change |
|------|--------|
| `ast/ast.go` | Update ExternalValue node handling |
| `ast/parser.go` | Update external value parsing |
| `ast/formatter.go` | Update external value formatting |

### 4. Checker

| File | Change |
|------|--------|
| `check/check.go` | Update semantic checks for external values |

### 5. Server (LSP)

| File | Change |
|------|--------|
| `server/completion.go` | Update completion for `<external>` |
| `server/hover.go` | Update hover for external values |

### 6. Documentation

| File | Change |
|------|--------|
| `README.md` | Update examples |
| `implementation/22-external-value.md` | Update plan |

### 7. Tests

| File | Change |
|------|--------|
| `token/tokenizer_test.go` | Update `$foo` → `<foo>` |
| `server/hover_test.go` | Update external value tests |
| All testdata files | Update grammar files |

## Implementation Steps

### Step 1: Update Grammar Definition
- [ ] `grammar.txt` - Change `external_value = "$" ident;` → `external_value = "<" ident ">";`

### Step 2: Update Tokenizer
- [ ] `token/tokenizer.go` - Change `$` recognition to `<` and `>`
- [ ] `token/tokenizer_test.go` - Update tests

### Step 3: Update Parser
- [ ] `ast/ast.go` - Update ExternalValue node
- [ ] `ast/parser.go` - Update parsing
- [ ] `ast/formatter.go` - Update formatting

### Step 4: Update Checker
- [ ] `check/check.go` - Update checks

### Step 5: Update Server
- [ ] Update server handlers

### Step 6: Update Tests
- [ ] Update all test files

### Step 7: Update Documentation
- [ ] `README.md`
- [ ] `implementation/22-external-value.md`

## Estimated Impact

- **Files to modify**: ~15 files
- **Test files**: ~5 files
- **Breaking change**: Yes

## Backward Compatibility

This is a **breaking change**. Existing `.grammar` files using `$name` will need to be updated to use `<name>`.

## Alternative Approach

Consider supporting both sigils during a transition period:
- Accept both `$` and `<>` as valid external value sigils
- Deprecate `$` in favor of `<>`
- Remove `$` support in a future version
