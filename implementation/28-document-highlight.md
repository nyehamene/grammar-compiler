# LSP Document Highlight

## Goal
Implement the `textDocument/documentHighlight` LSP feature. When the user's cursor is on a symbol, this feature highlights all other occurrences of that same symbol within the current document.

## Background
This feature provides immediate visual feedback about where a particular variable or rule is used in a file. It is similar to "Find All References" but is scoped only to the currently open file.

For example, given the following code:
```grammar
message = "Hello";         // Declaration
greeting = message " world"; // Usage 1
 farewell = message " bye";   // Usage 2
```
If the user's cursor is on any of the `message` identifiers, the server should respond with the ranges for all three occurrences, allowing the editor to highlight them.

## Implementation Plan

### 1. Update Server Capabilities
The server must announce its support for this feature during the `initialize` handshake.

-   **`server/initialize.go`**:
    -   In the `ServerCapabilities` struct, add a `DocumentHighlightProvider` field.
      ```go
      DocumentHighlightProvider *DocumentHighlightOptions `json:"documentHighlightProvider,omitempty"`
      ```
    -   In `handleInitializeRequest`, set this capability to indicate that the server provides document highlights.
      ```go
      DocumentHighlightProvider: &DocumentHighlightOptions{},
      ```

### 2. Add LSP Types
The necessary types for the request and response must be defined.

-   **`server/types.go`**:
    -   Define `DocumentHighlightParams`, which extends `TextDocumentPositionParams`.
    -   Define `DocumentHighlight`, which contains a `Range` and an optional `Kind`.
    -   Define `DocumentHighlightKind` (an integer enum for `Text`, `Read`, `Write`).
    -   Define `DocumentHighlightOptions` for the server capabilities.

### 3. Implement the Request Handler
A new handler will process `textDocument/documentHighlight` requests. This logic will be very similar to finding references but will be filtered to a single document.

-   **Create `server/documenthighlight.go`**:
    -   Implement the main handler function `handleDocumentHighlight(s *Server, id int, msg map[string]any)`.
    -   **Logic**:
        1.  Parse the `DocumentHighlightParams`.
        2.  Retrieve the AST and checker information for the requested document.
        3.  Find the AST node at the cursor's position using `ast.FindNodeAt`.
        4.  Use the checker's object resolution logic (`check.ObjectOf`) to find the declaration of the symbol at the cursor.
        5.  Use the checker's reference-finding logic (`check.UsesOf`) to get a list of all identifiers that refer to that declaration.
        6.  Iterate through the list of usages. For each usage that is within the requested document's URI, create a `DocumentHighlight` object.
        7.  The `Kind` can be determined by the context of the parent node (e.g., if the identifier is on the left side of a `RuleDecl`, it's a `Write` highlight; otherwise, it's a `Read` highlight).
        8.  Collect all created `DocumentHighlight` objects into a slice and send it as the response.

### 4. Update Server Dispatcher
The main request router must be updated to direct `textDocument/documentHighlight` requests to the new handler.

-   **`server/server.go`**:
    -   In the `handleRequest` switch statement, add a new case for `"textDocument/documentHighlight"` that calls `handleDocumentHighlight`.

### 5. Add Integration Test
A new test will be added to ensure the feature works correctly end-to-end.

-   **`server/integration_server_test.go`**:
    -   Add a new test function, `TestDocumentHighlightRequest`.
    -   **Setup**: Create a grammar file with a rule declaration and several usages of that rule.
    -   **Test Steps**:
        1.  Send a `didOpen` notification for the file.
        2.  Send a `textDocument/documentHighlight` request with the cursor positioned on one of the symbol's occurrences.
        3.  Read the response and assert that it contains the correct number of `DocumentHighlight` objects (one for the declaration and one for each usage).
        4.  Verify that the `Range` of each highlight correctly corresponds to the position of the symbol in the source file.
