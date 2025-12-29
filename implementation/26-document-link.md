# LSP Document Link

## Goal
Implement the `textDocument/documentLink` LSP feature. This will identify all file paths within `@import()` directives and present them as clickable links to the user in the editor.

## Background
The `@import("...")` directive in this grammar language contains a string literal that represents a path to another file in the local filesystem. The `textDocument/documentLink` request allows the server to tell the client which parts of the document are such links and where they point to.

For example, given the following code:
```grammar
foo = @import("foo.grammar");
```
The server should identify the range for the string `"foo.grammar"` and report that it is a link targeting the `foo.grammar` file.

## Implementation Plan

### 1. Update Server Capabilities
First, the server must announce that it can handle document link requests.

-   **`server/initialize.go`**:
    -   In the `ServerCapabilities` struct, add a `DocumentLinkProvider` field.
      ```go
      DocumentLinkProvider *DocumentLinkOptions `json:"documentLinkProvider,omitempty"`
      ```
    -   In `handleInitializeRequest`, set this capability to indicate that the server provides document links (resolve is not needed for now).
      ```go
      DocumentLinkProvider: &DocumentLinkOptions{
          ResolveProvider: false,
      },
      ```

### 2. Add LSP Types
The necessary LSP types for the request and response must be defined.

-   **`server/types.go`**:
    -   Define `DocumentLinkParams`, which contains the `TextDocument` identifier.
    -   Define `DocumentLink`, which contains a `Range` and a `Target` URI.
    -   Define `DocumentLinkOptions` for the server capabilities.

### 3. Implement the Request Handler
A new handler will be created to process incoming document link requests. This will involve finding all import declarations and creating link objects for them.

-   **Create `server/documentlink.go`**:
    -   Implement the main handler function `handleDocumentLink(s *Server, id int, msg map[string]any)`.
    -   **Logic**:
        1.  Parse the `DocumentLinkParams` from the request.
        2.  Retrieve the AST and source content for the requested document from the checker's compilation unit.
        3.  Traverse the AST using `ast.Walk` to find all `*ast.BindingDecl` nodes.
        4.  For each `BindingDecl`, check if its value is an `@import`.
        5.  If it is, extract the string literal containing the path.
        6.  **Resolve the Path**: Convert the relative path from the string literal into an absolute file URI. This logic can be adapted from the `resolveImport` function currently in `check/compilation.go`.
        7.  **Create Link**: Construct a `DocumentLink` object containing:
            -   `Range`: The range of the string literal within the document.
            -   `Target`: The absolute URI of the resolved file path.
        8.  Collect all created `DocumentLink` objects into a slice and send it as the response.

### 4. Update Server Dispatcher
The main request router must be updated to direct `textDocument/documentLink` requests to the new handler.

-   **`server/server.go`**:
    -   In the `handleRequest` switch statement, add a new case for `"textDocument/documentLink"` that calls `handleDocumentLink`.

### 5. Add Integration Test
A new test will be added to ensure the feature works correctly end-to-end.

-   **`server/integration_server_test.go`**:
    -   Add a new test function, `TestDocumentLinkRequest`.
    -   **Setup**: Create a temporary directory with two files: `a.grammar` (which imports `b.grammar`) and `b.grammar`.
    -   **Test Steps**:
        1.  Send a `didOpen` notification to the server for `a.grammar`.
        2.  Send a `textDocument/documentLink` request for `a.grammar`.
        3.  Read the response and assert that it contains exactly one `DocumentLink`.
        4.  Verify that the `Range` of the returned link corresponds to the position of the string `"b.grammar"`.
        5.  Verify that the `Target` of the link is the correct, absolute file URI for `b.grammar`.
