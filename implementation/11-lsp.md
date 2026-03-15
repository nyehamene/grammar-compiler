# Language server

Language server implementation.

The language server protocol implementation code and API should be placed in 'server' directory.

- Create a `Message`, `RequestMessage`, `ResponseMessage`, and `NotificationMessage` types.
- Add the `ResponseError` and `ErrorCodes` types.
- Implement a JSON RPC encoder and decoder and save it to the files `server/rpc.go`.
- Add basic lsp server type and update the lsp command to call its start method.

- Add a `DocumentUri` parser type that parses a uri string as illustrated below
  (Implemented in `server/uri.go`)

- Update the server to log its messages to the file ~/.cache/grammar/lsp.log
- The server now logs response messages.
- The server logs received messages and response messages in pretty-printed JSON format.
- The server logs received messages and response messages in a goroutine.
- The server now handles the `initialize` request and is successfully tested.
- Implement the `shutdown` request.
- Implement the `exit` notification.


## Implementation Plan

- [ ] **1. `textDocument/hover`**
    - [ ] **Types**: Add `HoverParams` and `Hover` response types to `server/types.go`.
    - [ ] **AST Position Logic**: Create a new helper function to find the `ast.Node` at a specific character offset. This will likely involve a new AST visitor.
    - [ ] **Handler**: Implement `handleHover` in `server/hover.go`, which will use the position logic to find the node under the cursor.
    - [ ] **Type Formatting**: Use the `checker.typeOf` method to get the type of the identified node and format it into a user-friendly string.
    - [ ] **Dispatcher**: Wire up the `textDocument/hover` method in `server.go`'s request handler.
    - [ ] **Test**: Add an integration test for the hover feature. (See `server/hover_test.go`)

- [ ] **2. `textDocument/definition` (Go to Definition)**
    - [ ] **Types**: Add `DefinitionParams` and `Location` types to `server/types.go`.
    - [ ] **Definition Logic**: Create a function that, given an `ast.Ident` node, finds its declaration (either in the current file or an imported one via the `CompilationUnit`).
    - [ ] **Handler**: Implement `handleDefinition` in `server/definition.go` to use the AST position logic and the new definition logic.
    - [ ] **Dispatcher**: Wire up the `textDocument/definition` method in `server.go`.
    - [ ] **Test**: Add an integration test for "Go to Definition." (See `server/definition_test.go`)

- [ ] **3. `textDocument/references` (Find All References)**
    - [ ] **Types**: Add `ReferenceParams` to `server/types.go`. The response is a `[]Location`.
    - [ ] **Reference Finding Logic**: Create a function that first finds a symbol's declaration, then traverses the AST of all documents in the `CompilationUnit` to find all usages.
    - [ ] **Handler**: Implement `handleReferences` in `server/references.go`.
    - [ ] **Dispatcher**: Wire up `textDocument/references` in `server.go`.
    - [ ] **Test**: Add an integration test for "Find All References." (See `server/references_test.go`)

- [ ] **4. `textDocument/documentSymbol`**
    - [ ] **Types**: Add `DocumentSymbolParams`, `DocumentSymbol`, and `SymbolKind` to `server/types.go`.
    - [ ] **Handler**: Implement `handleDocumentSymbol` in `server/documentsymbol.go`.
    - [ ] **Logic**: Traverse the top-level declarations of a document's AST to build a list of symbols.
    - [ ] **Dispatcher**: Wire up `textDocument/documentSymbol` in `server.go`.
    - [ ] **Test**: Add an integration test for document symbols. (See `server/document_symbol_test.go`)

- [ ] **5. `workspace/symbol`**
    - [ ] **Types**: Add `WorkspaceSymbolParams` and `SymbolInformation` to `server/types.go`.
    - [ ] **Handler**: Implement `handleWorkspaceSymbol` in `server/workspace_symbol.go`.
    - [ ] **Logic**: Iterate through all files in the `CompilationUnit`, find symbols matching the query, and return them as a list of `SymbolInformation`.
    - [ ] **Dispatcher**: Wire up `workspace/symbol` in `server.go`.
    - [ ] **Test**: Add an integration test for workspace symbols. (See `server/workspace_symbol_test.go`)

- [ ] **6. `textDocument/rename`**
    - [ ] **Types**: Add `RenameParams` and `WorkspaceEdit` to `server/types.go`.
    - [ ] **Handlers**: Implement `handlePrepareRename` and `handleRename` in `server/rename.go`.
    - [ ] **Logic**: Use the "Find All References" logic to get all locations of a symbol and create a `WorkspaceEdit` containing `TextEdit`s to perform the rename.
    - [ ] **Dispatcher**: Wire up `textDocument/prepareRename` and `textDocument/rename` in `server.go`.
    - [ ] **Test**: Add an integration test for the rename feature. (See `server/rename_test.go`)

- [ ] **7. `textDocument/completion`**
    - [ ] **Types**: Add `CompletionParams`, `CompletionItem`, `CompletionList` to `server/types.go`.
    - [ ] **Handler**: Implement `handleCompletion` in `server/completion.go`.
    - [ ] **Logic**: Provide completions for rule names, binding references, package members, and import paths.
    - [ ] **Dispatcher**: Wire up `textDocument/completion` in `server.go`.
    - [ ] **Test**: Add an integration test for completion. (See `server/completion_test.go`)

- [ ] **8. `textDocument/diagnostic` (Pull Diagnostics)**
    - [ ] **Types**: Add diagnostic-related types to `server/types.go`.
    - [ ] **Handler**: Implement `handleTextDocumentDiagnostic` in `server/diagnostics.go`.
    - [ ] **Logic**: Call the checker and return diagnostics.
    - [ ] **Dispatcher**: Wire up `textDocument/diagnostic` in `server.go`.
    - [ ] **Test**: Add an integration test for pull diagnostics. (See `server/diagnostics_test.go`)

- [ ] **9. `textDocument/documentHighlight`**
    - [ ] **Types**: Add `DocumentHighlightParams` and `DocumentHighlight` to `server/types.go`.
    - [ ] **Handler**: Implement `handleDocumentHighlight` in `server/document_highlight.go`.
    - [ ] **Logic**: Find all references to the symbol at the cursor position.
    - [ ] **Dispatcher**: Wire up `textDocument/documentHighlight` in `server.go`.
    - [ ] **Test**: Add an integration test for document highlight. (See `server/document_highlight_test.go`)

- [ ] **10. `textDocument/documentLink`**
    - [ ] **Types**: Add `DocumentLinkParams` and `DocumentLink` to `server/types.go`.
    - [ ] **Handler**: Implement `handleDocumentLink` in `server/documentlink.go`.
    - [ ] **Logic**: Generate links for `@import` and `@package` directives.
    - [ ] **Dispatcher**: Wire up `textDocument/documentLink` in `server.go`.
    - [ ] **Test**: Add an integration test for document links. (See `server/document_link_test.go`)

## LSP features

## Gemini
Note: This feature implementation is not complete yet.
More steps will be added above until the lsp implementation is complete.
When this section is removed then the implement has been completed.
