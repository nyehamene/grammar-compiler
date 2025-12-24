# Tokenizer

Scan input grammar files. If any error is found exit
with a non-zero exit code.

See ../grammar.txt for the valid tokens in the grammar.

- Implement `grammar check <PATH>`.
- Add tests that scans each file in ../example directory.

## Implementation Note

- A token should have a public `kind` property.

- A token should have a public `state` property
  that indicates if its valid or invalid.

  For example, when attempting to scan a string literal,
  if an error is found, the `state` property should indicate
  that the string is invalid, and the `kind` property should
  indicate that the token is a string.

- The tokenizer should a `scan` method that scans an input file
  and return a `[]Token`.

- The tokenizer should not report any errors. All error reporting
will be done in the parser that we will implement later.
