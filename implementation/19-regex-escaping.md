# Tokenizer Regex Escape Sequences

## Goal
To enhance the tokenizer to correctly handle backslash-escaped characters within regex literals (`/.../`) and to flag unsupported escape sequences as errors.

## Background
The current regex scanner (`scanRegex` in `token/tokenizer.go`) is simple and terminates at the first `/` character it encounters. This prevents the use of an escaped forward slash (e.g., `\/`) within the regex itself. This implementation will add support for a specific set of escape sequences and handle errors gracefully.

## Supported Escape Sequences
The only escape sequences that will be considered valid are:
- `\/` (forward slash)
- `\n` (newline)
- `\t` (tab)
- `\r` (carriage return)
- `\d` (digit character class)
- `\s` (whitespace character class)
- `\c` (as specified, to be clarified)
- `\\` (backslash)
- `\.` (dot)
- `\(` (left parenthesis)
- `\)` (right parenthesis)
- `\{` (left brace)
- `\}` (right brace)
- `\[` (left bracket)
- `\]` (right bracket)

Any other character following a backslash (`\`) will be considered an invalid escape sequence.

## Action Plan

- [x] **1. Update Tokenizer Logic in `token/tokenizer.go`**
The `scanRegex` method needs to be rewritten to handle escape sequences.

- The method's loop should continue until it finds a `/` that is **not** preceded by an unescaped `\`.
- Inside the loop, if the current character is a backslash (`\\`), the tokenizer must consume it and then inspect the **next** character.
- The character following the backslash must be one of the allowed escaped characters listed above.
- If the escaped character is valid, the tokenizer should consume it and continue scanning.
- If the character following a backslash is **not** in the supported list, the tokenizer should stop and return a `Regex` token with its `State` set to `token.Invalid`.
- If the end of the file is reached before a closing `/` is found, the token's state should be `token.Invalid` (this preserves existing unterminated regex logic).

- [x] **2. Update Parser Error Reporting in `ast/parser.go`**
The parser was previously updated to report a generic error for invalid tokens. We can make the error message for invalid regexes more specific.

- In the `parseTerminal` method, when an invalid `Regex` token is found, the error message should be changed to `"invalid regex literal"`. This is a general message that covers both unterminated literals and the new case of invalid escape sequences.

- [x] **3. Add Tokenizer Tests**
New unit tests must be added to `token/tokenizer_test.go` to validate the updated `scanRegex` logic.

- [x] **Valid Test Cases:**
  - A regex containing an escaped forward slash: `r = /[^\/]/;`
  - A regex containing a mix of other valid escapes: `r = /\n\t\d\s\(\)\[\]\{\}\.\\/;`
  - A regex ending with an escaped backslash: `r = /abc\\/;`

- [x] **Invalid Test Cases:**
  - A regex with an unsupported escape sequence: `r = /\a/`. The test must assert that the resulting `Regex` token has `State: token.Invalid`.
  - A regex with a dangling escape at the very end of the file: `r = /abc\/`

- [x] **4. Add Parser Tests**
New integration tests should be added to `ast/parser_test.go`.

- Add a new test case to `TestParseInvalidToken` to parse a file containing a rule with an invalid escape sequence, e.g., `rule = /\z/;`.
- The test should assert that `parser.ParseFile()` returns an error containing the message `"invalid regex literal"`.

## Note on `\c`
The escape sequence `\c` is not a standard regex escape. Before implementation, its intended meaning should be clarified. For this plan, it is treated as a valid single-character escape, and the tokenizer will be implemented to accept it.
