## Implementation Plan: Unused Symbol Analysis

### Goal
The objective is to enhance the language server to provide diagnostics for unused private symbols and to enforce a single public rule per file. This will improve code quality by highlighting dead code and enforcing a clear module interface.

### Background
The analysis will be performed on a per-file basis. The definitions for symbol visibility are as follows:
- **Private Rule:** A rule whose name begins with a lowercase letter (e.g., `myRule`). It is only intended for use within the defining file.
- **Public Rule:** A rule whose name begins with an uppercase letter (e.g., `MyRule`). It can be referenced from other files.
- **Binding:** An imported grammar (e.g., `import "..." as B;`). Bindings are always considered private to the file they are imported in.

The new diagnostic rules are:
1.  A `Warning` should be generated for any private rule that is not referenced within the same file.
2.  A `Warning` should be generated for any binding that is not used within the same file.
3.  A `Warning` should be generated if a file declares more than one public rule.

### Todos

-   [x] **1. Enhance the Checker for Symbol Analysis**
    -   [x] In `check/check.go`, augment the `Check` function to track declarations and usages.
    -   [x] Create a data structure to store symbol information, including its name, whether it's public, its declaration location (for reporting), and a flag to track if it's been used.
    -   [x] Iterate through the AST to populate this structure with all rule and binding declarations. While doing so, maintain a count of public rules.
    -   [x] Perform a second pass over the AST to find all non-terminal references. For each reference found, mark the corresponding symbol in your data structure as "used".
    -   [x] After the passes are complete, iterate through the symbol data structure. For each private symbol not marked as "used", generate a new "unused symbol" warning diagnostic.
    -   [x] If the public rule count is greater than one, generate a new "multiple public rules" warning diagnostic.
    -   [x] Ensure the new diagnostics are added to the list of errors returned by the `Check` function.

-   [x] **2. Integrate New Diagnostics into the LSP Server**
    -   [x] In `server/diagnostics.go`, locate the handler for `textDocument/diagnostic` requests.
    -   [x] This handler already calls the `check.Check` function. Ensure the new warnings returned from the checker are processed.
    -   [x] Convert the new warning types into LSP `Diagnostic` objects with `Severity` set to `protocol.DiagnosticSeverityWarning`. The position of the warning should correspond to the location of the unused symbol's declaration.

-   [x] **3. Add Comprehensive Tests**
    -   [x] In `check/check_test.go`, add new unit tests to validate the checker logic:
        -   [x] A test case with an unused private rule.
        -   [x] A test case with an unused binding.
        -   [x] A test case where a private rule is correctly used.
        -   [x] A test case with an unused public rule (should produce no warning).
        -   [x] A test case with two public rules (should produce a warning).
        -   [x] A test case with one public rule (should produce no warning).
    -   [x] In `server/diagnostics_test.go`, add integration tests that simulate an LSP client requesting diagnostics for `.grammar` files containing the above scenarios and assert that the correct warnings are returned.