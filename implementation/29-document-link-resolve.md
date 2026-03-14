# LSP Document Link Resolve

## Goal
Implement the `documentLink/resolve` LSP feature to lazily provide the full details of a document link, including a tooltip.

## Background
The `textDocument/documentLink` request is used to identify links within a document (like `@import` paths). For performance, the server can return unresolved links and the client can then request to "resolve" a specific link just before it's needed (e.g., when the user hovers over it).

This implementation will update the existing `documentLink` feature to support this lazy resolution. The `documentLink/resolve` request will populate the `target` URI and add a helpful `tooltip`.

## Implementation Plan

### 1. Update Server Capabilities
The server must announce its support for resolving document links.

- [ ] **`server/initialize.go`**:
    - In `handleInitializeRequest`, find the `DocumentLinkProvider` options.
    - Set `ResolveProvider: true`.

### 2. Update LSP Types
The `DocumentLink` struct needs to be updated to support the `resolve` flow.

- [ ] **`server/types.go`**:
    - Add a `Data any `json:"data,omitempty"` field to the `DocumentLink` struct. This field will be used to carry the necessary information from the `documentLink` request to the `documentLink/resolve` request.
    - Add a `Tooltip string `json:"tooltip,omitempty"` field to the `DocumentLink` struct. This will be populated by the resolve handler.

### 3. Update `textDocument/documentLink` Handler
The existing `documentLink` handler must be modified to return unresolved links.

- [ ] **`server/documentlink.go`**:
    - Create a new struct (e.g., `documentLinkData`) to hold the information needed for resolution. It should contain the URI of the document containing the link and the raw import path string.
    - In `handleDocumentLink`, instead of resolving the path and setting the `Target`, do the following for each import:
        1. Create an instance of `documentLinkData` with the current document's URI and the import path.
        2. Create a `DocumentLink` object.
        3. Set the `Range` as before.
        4. Leave the `Target` as `nil`.
        5. Set the `Data` field to the `documentLinkData` instance.
    - This ensures the initial response is fast and contains all the necessary context for the `resolve` request.

### 4. Implement `documentLink/resolve` Handler
A new handler is needed to process the `documentLink/resolve` request.

- [ ] **`server/documentlink.go`**:
    - Create a new function `handleDocumentLinkResolve(s *Server, id int, msg map[string]any)`.
    - **Logic**:
        1.  Parse the `DocumentLink` object from the request parameters. This is the unresolved link that was sent in the `documentLink` response.
        2.  Check if the `Data` field is present. If not, return the link as-is.
        3.  Unmarshal the `Data` field back into the `documentLinkData` struct.
        4.  Using the original document's URI and the import path from the unmarshalled data, resolve the import path to an absolute file URI.
        5.  Update the `Target` field of the `DocumentLink` with the resolved URI.
        6.  Set the `Tooltip` field to a user-friendly string, such as the absolute path of the target file.
        7.  Send the now fully-resolved `DocumentLink` object back as the response.

### 5. Update Server Dispatcher
The main request router needs to know about the new `documentLink/resolve` method.

- [ ] **`server/server.go`**:
    - In the `handleRequest` switch statement, add a new case for `"documentLink/resolve"` that calls the new `handleDocumentLinkResolve` handler.

### 6. Add/Update Integration Test
The tests must be updated to validate the new resolve-based workflow.

- [ ] **`server/integration_server_test.go`**:
    - Modify the existing `TestDocumentLinkRequest` to become `TestDocumentLinkResolve`.
    - **Test Steps**:
        1.  Send a `textDocument/documentLink` request as before.
        2.  Assert that the response is an array of `DocumentLink` objects where `Target` is `nil` and `Data` is not.
        3.  Take the first link from the response and create a `documentLink/resolve` request with it.
        4.  Send the `documentLink/resolve` request.
        5.  Read the response and assert that it is a single, complete `DocumentLink` object.
        6.  Verify that the `Target` is now correctly populated with the absolute URI.
        7.  Verify that the `Tooltip` field is present and contains the expected path.
