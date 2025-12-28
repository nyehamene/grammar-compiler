# Completion

Implement `textDocument/completion` to provide context-aware completion suggestions.

## Completion Scenarios

1.  **Member Access Completion**:
    - When the user types a `.` after an identifier that is bound to a namespace, the server should suggest all the rules defined in that namespace.
    - Example:
      ```grammar
      foo = @import("foo.grammar");
      bar = foo. // <-- Trigger completion here
      ```
    - If `foo.grammar` contains `name = "Foo";` and `age = /[0-9]+/;`, the suggestions should include `name` and `age`.

2.  **Rule Body Completion**:
    - When the user is inside a rule's body, the server should suggest all available symbols in the current scope.
    - This includes:
        - Rules defined in the current file.
        - Bindings (namespaces) defined in the current file.
        - Pre-defined external values (e.g., `$STRING`).
    - Example:
      ```grammar
      name = "John";
      message = // <-- Trigger completion here
      ```
    - Suggestions should include `name`.

## Implementation Plan

### 1. Update LSP Types (`server/types.go`)
-   Add the necessary LSP struct definitions for code completion:
    -   `CompletionParams`
    -   `CompletionContext`
    -   `CompletionItem`
    -   `CompletionList`
    -   `CompletionItemKind`

### 2. Create Completion Handler (`server/completion.go`)
-   Create a new file `server/completion.go`.
-   Implement the main `handleCompletion(id int, msg map[string]any)` function. This function will:
    1.  Parse the `CompletionParams` from the request.
    2.  Get the document content and cursor position.
    3.  Find the AST node at the cursor position using `ast.FindNodeAt`. The parent of this node is also needed for context.

### 3. Implement Completion Logic
-   Create a central function `getCompletions(node, parent ast.Node, ns *check.Namespace, cu *check.CompilationUnit) []CompletionItem` that determines the completion context and generates suggestions.
-   **Member Access Logic**:
    -   Check if the node before the cursor is a `.` in a `MemberExpr`.
    -   If so, get the receiver of the expression and use the type checker to confirm it's a `NamespaceType`.
    -   Get the corresponding namespace from the `CompilationUnit` and create `CompletionItem`s for each of its exported rules.
-   **Rule Body / Top-Level Logic**:
    -   If not a member access, gather all symbols (rules and bindings) from the current namespace (`ns.Members`).
    -   Create a `CompletionItem` for each symbol.
    -   Add suggestions for any pre-defined external values (e.g., `$STRING`).

### 4. Construct `CompletionItem`
-   For each suggestion, create a `CompletionItem` with the following details:
    -   `Label`: The name of the identifier (e.g., `my_rule`).
    -   `Kind`: Use `CompletionItemKindFunction` for rules and `CompletionItemKindModule` for namespaces.
    -   `Detail`: A short description, like `(production)` or `(namespace)`.
    -   `Documentation`: A more detailed description. This can reuse the logic from the `textDocument/hover` implementation to show the rule's body or the namespace's import path.

### 5. Update Server Dispatcher (`server/server.go`)
-   In `handleRequest`, add a `case` for `textDocument/completion` that calls the new `handleCompletion` function.

### 6. Add Integration Tests (`server/integration_server_test.go`)
-   Add a new test function `TestCompletion`.
-   This test will include sub-tests for each completion scenario:
    1.  **Member Access**: Open two files, one importing the other. Trigger completion after a `.` on the namespace binding and assert that the correct rule names are returned.
    2.  **Rule Body**: In a single file, define a few rules. Trigger completion in the body of another rule and assert that the previously defined rule names are suggested.
