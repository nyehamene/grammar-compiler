# LSP Pull Diagnostics

## Goal
Implement the LSP "pull" diagnostic model, as introduced in LSP 3.17. This allows the client to explicitly request diagnostics for a document, complementing the existing "push" model where the server sends diagnostics proactively on file changes.

## Background
Currently, the server uses `textDocument/publishDiagnostics` notifications to push errors to the client after a `didOpen` or `didChange` event. The pull model introduces a new request/response flow:
1.  The client sends a `textDocument/diagnostic` request.
2.  The server responds with a `DocumentDiagnosticReport` containing all diagnostics for that file.

This approach gives clients more control over when they receive diagnostic information.

## Implementation Plan

### 1. Update Server Capabilities
The server must announce its support for pull diagnostics during the `initialize` handshake.

-   **`server/types.go`**:
    -   Add `DiagnosticServerCapabilities` and `DiagnosticOptions` structs.
-   **`server/initialize.go`**:
    -   In `ServerCapabilities`, add a `DiagnosticProvider` field of type `*DiagnosticOptions`.
    -   In `handleInitializeRequest`, set the `DiagnosticProvider` to indicate support for pull diagnostics. For now, we will enable workspace diagnostics and report inter-file dependencies.
      ```go
      DiagnosticProvider: &DiagnosticOptions{
          WorkspaceDiagnostics: true,
          InterFileDependencies: true,
      },
      ```

### 2. Add Pull Diagnostic Types
New LSP types are required to handle the request/response cycle.

-   **`server/types.go`**:
    -   `DocumentDiagnosticParams`: Parameters for the `textDocument/diagnostic` request.
    -   `DocumentDiagnosticReport`: A container for the response, which can be one of several kinds.
    -   `RelatedFullDocumentDiagnosticReport`: The full list of diagnostics for a document.
    -   `RelatedUnchangedDocumentDiagnosticReport`: A response indicating diagnostics have not changed (for a later optimization with `resultId`).

### 3. Implement the Request Handler
A new handler will process `textDocument/diagnostic` requests.

-   **`server/diagnostics.go`**:
    -   Create a new function `handleDocumentDiagnostic(s *Server, id int, msg map[string]any)`.
    -   **Logic**:
        1.  Unmarshal the `DocumentDiagnosticParams`.
        2.  Run the checker on the specified document's content to get a list of `check.Error`.
        3.  Convert the errors into LSP `Diagnostic` objects. This logic can be extracted from the existing `publishDiagnostics` function into a shared helper.
        4.  Construct a `RelatedFullDocumentDiagnosticReport` containing the diagnostics.
        5.  Send the report back to the client using `s.sendResponse`.

### 4. Refactor Diagnostic Generation
To avoid code duplication, the core logic for generating diagnostics should be shared.

-   **`server/diagnostics.go`**:
    -   Create a helper function, e.g., `generateDiagnosticsForURI(s *Server, uri DocumentUri) []Diagnostic`.
    -   This function will contain the logic to get document content, run the checker, and convert errors to `Diagnostic`s.
    -   Update the existing `publishDiagnostics` function and the new `handleDocumentDiagnostic` handler to both use this new helper.

### 5. Update Server Dispatcher
The main request router needs to know about the new method.

-   **`server/server.go`**:
    -   In the `handleRequest` method's `switch` statement, add a new case for `"textDocument/diagnostic"`.
    -   This case should call the new `handleDocumentDiagnostic` handler.

### 6. Add Integration Tests
Verify the implementation with a new end-to-end test.

-   **`server/integration_server_test.go`**:
    -   Add a new test function, `TestDocumentDiagnosticRequest`.
    -   The test should:
        1.  Set up the server.
        2.  Send a `didOpen` notification with a file containing a known syntax error.
        3.  Consume the initial `publishDiagnostics` notification.
        4.  Send a `textDocument/diagnostic` request for that same file.
        5.  Read the server's response and assert that it is a `DocumentDiagnosticReport` containing the correct diagnostic information for the syntax error.
