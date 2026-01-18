# Plan: Conditional Diagnostics Publishing

## Goal
To update the language server to respect client capabilities regarding diagnostics. If a client supports the "pull" diagnostic model (`textDocument/diagnostic`), the server should no longer "push" diagnostics using the legacy `textDocument/publishDiagnostics` notification.

## Background
The Language Server Protocol (LSP) 3.17 introduced a "pull diagnostics" model where the client requests diagnostics from the server. Older clients rely on the server pushing diagnostics. The server currently pushes diagnostics on events like `didOpen` and `didChange`. This behavior should be suppressed for modern clients that support the pull model.

## Implementation Plan

- [x] **1. Track Client Capabilities (`server/server.go`)**
    - [x] Add a new boolean field `clientHasDiagnosticSupport` to the `Server` struct. This field will track whether the connected client supports pull diagnostics.

- [x] **2. Process Client Capabilities on Initialization (`server/initialize.go`)**
    - [x] In the `handleInitializeRequest` function, inspect the `params.Capabilities.TextDocument.Diagnostic` from the client.
    - [x] Safely check for the presence of this capability. If it exists, set `s.clientHasDiagnosticSupport = true`.

- [x] **3. Conditionally Publish Diagnostics (`server/diagnostics.go`)**
    - [x] In the `publishDiagnostics` function, add a check at the beginning. If `s.clientHasDiagnosticSupport` is true, return immediately without sending the `textDocument/publishDiagnostics` notification.

- [x] **4. Update Integration Tests (`server/integration_server_test.go`)**
    - [x] Add a new test to verify the conditional logic.
    - [x] **Scenario 1 (With Pull Support):** Initialize the server with client capabilities that include pull diagnostics, open a file with an error, and assert that **no** `publishDiagnostics` notification is received.
    - [x] **Scenario 2 (Without Pull Support):** Initialize the server with capabilities that lack pull diagnostics, open a file with an error, and assert that a `publishDiagnostics` notification **is** received.
