# Tokenizer String Escape Sequences

## Goal
To enhance the tokenizer to correctly handle backslash-escaped characters within string literals (`"...")`, and to flag unsupported escape sequences as errors.

## Background
The current string scanner (`scanString` in `token/tokenizer.go`) does not process escape sequences. This means a literal string like `"Hello\nWorld"` would be tokenized as a raw string containing a backslash followed by 'n', which is incorrect. This implementation will add support for a specific set of escape sequences.

## Supported Escape Sequences (for string literals)
The only escape sequences that will be considered valid are:
- `\"` (double quote)
- `\\` (backslash)
- `\n` (newline)
- `\t` (tab)
- `\r` (carriage return)
- `\c` (consistent with regex, meaning to be clarified)
- `\'` (single quote)

Any other character following a backslash (`\`) will be considered an invalid escape sequence.

## Action Plan

### 1. Update Tokenizer Logic in `token/tokenizer.go`
The `scanString` method needs to be rewritten to handle escape sequences.

- The method's loop should continue until it finds a `"` that is **not** preceded by an unescaped `\`.
- Inside the loop, if the current character is a backslash (`\\`), the tokenizer must consume it and then inspect the **next** character.
- The character following the backslash must be one of the allowed escaped characters listed above.
- If the escaped character is valid, the tokenizer should consume it and continue scanning. At this stage, the tokenizer will *not* unescape the character (e.g., `\n` will remain two characters `\` and `n` in the token's literal value). The unescaping will be handled by the parser/AST.
- If the character following a `\` is **not** in the supported list, the tokenizer should stop and return a `String` token with its `State` set to `token.Invalid`.
- If the end of the file is reached before a closing `"` is found, the token's state should be `token.Invalid` (this preserves existing unterminated string logic).

### 2. Update Parser Error Reporting in `ast/parser.go`
The parser was previously updated to report a generic error for invalid tokens.

- In the `parseTerminal` method, when an invalid `String` token is found, the error message should be changed to `"invalid string literal"`. This is a general message that covers both unterminated literals and the new case of invalid escape sequences.

### 3. Add Tokenizer Tests
New unit tests must be added to `token/tokenizer_test.go` to validate the updated `scanString` logic.

- **Valid Test Cases:**
  - A string containing an escaped double quote: `s = "foo\"bar";`
  - A string containing a mix of other valid escapes: `s = "hello\n\t\r\\world";`
  - A string ending with an escape: `s = "final\\";`
  - A string with an escaped single quote: `s = "It\'s fine";`

- **Invalid Test Cases:**
  - A string with an unsupported escape sequence: `s = "\a";`. The test must assert that the resulting `String` token has `State: token.Invalid`.
  - An unterminated string (already covered, but good to re-verify): `s = "abc`

### 4. Add Parser Tests
New integration tests should be added to `ast/parser_test.go`.

- Add a new test case to `TestParseInvalidToken` to parse a file containing a rule with an invalid escape sequence, e.g., `rule = "\z";`.
- The test should assert that `parser.ParseFile()` returns an error containing the message `"invalid string literal"`.

## Note on `\c` and Unescaping
- The escape sequence `\c` is not a standard regex or string escape. Before final implementation, its intended meaning should be clarified. For this plan, it is treated as a valid single-character escape.
- The tokenizer will only identify the boundaries and validity of the string literal, including escape sequences. It will *not* convert `\n` into an actual newline character. The conversion (unescaping) to the final string value should be handled in a later stage, likely when constructing the `StringLit.Value` in the AST.
