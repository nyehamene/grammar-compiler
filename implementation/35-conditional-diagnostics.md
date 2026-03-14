# Plan: Conditional Diagnostics Publishing

## Goal
To update the language server to respect client capabilities regarding diagnostics. If a client supports the "pull" diagnostic model (`textDocument/diagnostic`), the server should no longer "push" diagnostics using the legacy `textDocument/publishDiagnostics` notification.

## Background
The Language Server Protocol (LSP) 3.17 introduced a "pull diagnostics" model where the client requests diagnostics from the server. Older clients rely on the server pushing diagnostics. The server currently pushes diagnostics on events like `didOpen` and `didChange`. This behavior should be suppressed for modern clients that support the pull model.

## Implementation Plan

- [ ] **1. Track Client Capabilities (`server/server.go`)**
    - [ ] Add a new boolean field `clientHasDiagnosticSupport` to the `Server` struct. This field will track whether the connected client supports pull diagnostics.

- [ ] **2. Process Client Capabilities on Initialization (`server/initialize.go`)**
    - [ ] In the `handleInitializeRequest` function, inspect the `params.Capabilities.TextDocument.Diagnostic` from the client.
    - [ ] Safely check for the presence of this capability. If it exists, set `s.clientHasDiagnosticSupport = true`.

- [ ] **3. Conditionally Publish Diagnostics (`server/diagnostics.go`)**
    - [ ] In the `publishDiagnostics` function, add a check at the beginning. If `s.clientHasDiagnosticSupport` is true, return immediately without sending the `textDocument/publishDiagnostics` notification.

- [ ] **4. Update Integration Tests (`server/integration_server_test.go`)**
    - [ ] Add a new test to verify the conditional logic.
    - [ ] **Scenario 1 (With Pull Support):** Initialize the server with client capabilities that include pull diagnostics, open a file with an error, and assert that **no** `publishDiagnostics` notification is received.
    - [ ] **Scenario 2 (Without Pull Support):** Initialize the server with capabilities that lack pull diagnostics, open a file with an error, and assert that a `publishDiagnostics` notification **is** received.
