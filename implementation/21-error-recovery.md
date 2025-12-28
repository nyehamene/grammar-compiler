# Error Recovery

## 1. Goal
To make the parser more resilient to syntax errors. Instead of stopping at the first error, the parser will report the error, resynchronize to a known state, and continue parsing to find any subsequent errors in the file.

## 2. Strategy
The recovery strategy will be based on synchronization. When the parser encounters a syntax error, it will enter a "recovery mode". In this mode, it will discard tokens until it reaches a "safe point" that likely indicates the start of a new, valid declaration.

The user has specified `;` and `comment` as synchronization points.

- **Recovery Mode Trigger:** The `errorf` method will be updated to set a `recovering` flag in the parser state. This will be done only if the parser is not already recovering, preventing a cascade of errors from a single mistake.
- **Synchronization in `parseDecl`:** The main declaration parsing function, `parseDecl`, will check the `recovering` flag at its entry point. If true, it will call a `synchronize()` method to perform the recovery.
- **Synchronization Points:**
    - **Semicolon (`;`):** This token marks the end of a declaration. When a `;` is found during recovery, the parser will consume it and exit recovery mode. This positions the parser to correctly parse the next declaration.

## 3. Implementation Steps

### a. Update `ast/parser.go`

1.  **Add `recovering` flag to `Parser` struct:**
2.  **Modify `errorf` to set the recovery flag:**
3.  **Modify `parseDecl` to call `synchronize`:**
4.  **Add the `synchronize` method:**

### b. Testing Strategy

1.  Create a new test file in `testdata/parser/` named `recovery.grammar`.
2.  This file will contain multiple syntax errors, such as:
    -   A rule with a missing `=`.
    -   A declaration with invalid tokens in the body.
    -   A valid rule following several invalid ones.
3.  Create a new test in `ast/parser_test.go` that parses `recovery.grammar`.
4.  The test will assert that:
    -   The parser reports the correct number of errors (one for each mistake, not a cascade).
    -   The AST returned by the parser correctly contains the valid declarations that were parsed after recovering from errors.

This ensures the recovery logic not only reports errors correctly but also successfully resumes parsing.
