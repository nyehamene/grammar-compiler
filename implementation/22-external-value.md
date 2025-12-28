### Implementation Plan

1.  **Tokenizer (`token/`)**:
    *   **File**: `token/token.go`
    *   **Action**: Add a new token `Kind` named `External` to represent the `$ident` syntax.
    *   **File**: `token/tokenizer.go`
    *   **Action**: Update the tokenizer to recognize sequences starting with `$` followed by an identifier, and emit them as `External` tokens.

2.  **Abstract Syntax Tree (`ast/`)**:
    *   **File**: `ast/ast.go`
    *   **Action**: Define a new AST node struct named `ExternalValue` that implements the `Expr` interface. This node will represent the `$ident` value in the AST.
    *   **File**: `ast/walk.go`
    *   **Action**: Update the `Visitor` interface and `Walk` function to handle the new `ExternalValue` node type.

3.  **Parser (`ast/`)**:
    *   **File**: `ast/parser.go`
    *   **Action**: Modify the `parseBasic` method to handle the new `External` token, parsing it into an `ExternalValue` AST node.

4.  **Tree-sitter Grammar (`treesitter/`)**:
    *   **File**: `treesitter/grammar.js`
    *   **Action**: Add a new rule for `external_value` that matches a `$` followed by an `ident`. Update the `terminal` rule to include `external_value` as an alternative.

5.  **Formatter (`ast/`)**:
    *   **File**: `ast/formatter.go`
    *   **Action**: Update the formatter's AST traversal logic to recognize and correctly format `ExternalValue` nodes, printing them back as `$name`.

6.  **Token Printer (`cmd/print/`)**:
    *   **File**: `cmd/print/print.go`
    *   **Action**: Enhance the token printer to handle the new `External` token, ensuring it is displayed correctly, including any colorization or special formatting.

7.  **Checker (`check/`)**:
    *   **File**: `check/check.go`
    *   **Action**: Add logic to the semantic checker to handle `ExternalValue` nodes. This will likely involve adding a new case in the checker's AST traversal, completing the integration of the new type.
