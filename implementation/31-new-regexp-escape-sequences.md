# Add new regexp escape sequences

## Goal
To enhance the tokenizer to support additional common regex escape sequences: `*` (asterisk), `+` (plus), and `?` (question mark).

## Background
The current regex scanner, as detailed in `implementation/19-regex-escaping.md`, supports a set of escape sequences. To improve the expressiveness of the grammar's regex literals, we need to add support for escaping characters that have special meaning in regular expressions.

## Updated Supported Escape Sequences
The list of valid escape sequences will be extended to include:
- `*` (asterisk)
- `+` (plus)
- `?` (question mark)

This is in addition to the existing supported sequences: `\/`, `\n`, `\t`, `\r`, `\d`, `\s`, `\c`, `\\`, `\.`, `\(`, `\)`, `\{`, `\}`, `\[`, `\]`.

## Action Plan

- [ ] **1. Update Tokenizer Logic in `token/tokenizer.go`**
    The primary change is in the `scanRegex` method. The set of valid characters that can follow a backslash (`\`) must be updated to include `*`, `+`, and `?`. The logic will remain the same: if a character following a backslash is not in the supported set, the tokenizer will mark the regex token's `State` as `token.Invalid`.

- [ ] **2. Add Tokenizer Tests in `token/tokenizer_test.go`**
    Add new test cases to the existing regex tests to verify the correct handling of the new escape sequences.

    - **Valid Test Case:**
      - A regex containing the new escaped characters: `r = /a*b+c?/;`

    The tests should ensure that these sequences are accepted and that the tokenizer correctly identifies the end of the regex literal.

- [ ] **3. Parser and Error Handling**
    No changes are expected in the parser (`ast/parser.go`). It already correctly identifies tokens with an `Invalid` state and reports a generic `"invalid regex literal"` error. This behavior is sufficient. No new parser tests are required.