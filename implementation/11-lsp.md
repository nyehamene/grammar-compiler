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


## Todos

- [x] **1. `textDocument/hover`**
    - [x] **Types**: Add `HoverParams` and `Hover` response types to `server/types.go`.
    - [x] **AST Position Logic**: Create a new helper function to find the `ast.Node` at a specific character offset. This will likely involve a new AST visitor.
    - [x] **Handler**: Implement `handleHover` in `server/hover.go`, which will use the position logic to find the node under the cursor.
    - [x] **Type Formatting**: Use the `checker.typeOf` method to get the type of the identified node and format it into a user-friendly string.
    - [x] **Dispatcher**: Wire up the `textDocument/hover` method in `server.go`'s request handler.
    - [x] **Test**: Add an integration test for the hover feature.

- [x] **2. `textDocument/definition` (Go to Definition)**
    - [x] **Types**: Add `DefinitionParams` and `Location` types to `server/types.go`.
    - [x] **Definition Logic**: Create a function that, given an `ast.Ident` node, finds its declaration (either in the current file or an imported one via the `CompilationUnit`).
    - [x] **Handler**: Implement `handleDefinition` in `server/definition.go` to use the AST position logic and the new definition logic.
    - [x] **Dispatcher**: Wire up the `textDocument/definition` method in `server.go`.
    - [x] **Test**: Add an integration test for "Go to Definition."

- [ ] **3. `textDocument/references` (Find All References)**
    - [ ] **Types**: Add `ReferenceParams` to `server/types.go`. The response is a `[]Location`.
    - [ ] **Reference Finding Logic**: Create a function that first finds a symbol's declaration, then traverses the AST of all documents in the `CompilationUnit` to find all usages.
    - [ ] **Handler**: Implement `handleReferences` in `server/references.go`.
    - [ ] **Dispatcher**: Wire up `textDocument/references` in `server.go`.
    - [ ] **Test**: Add an integration test for "Find All References."

- [ ] **4. `textDocument/documentSymbol`**
    - [ ] **Types**: Add `DocumentSymbolParams`, `DocumentSymbol`, and `SymbolKind` to `server/types.go`.
    - [ ] **Handler**: Implement `handleDocumentSymbol` in `server/documentsymbol.go`.
    - [ ] **Logic**: Traverse the top-level declarations of a document's AST to build a list of symbols.
    - [ ] **Dispatcher**: Wire up `textDocument/documentSymbol` in `server.go`.
    - [ ] **Test**: Add an integration test for document symbols.

- [ ] **5. `workspace/symbol`**
    - [ ] **Types**: Add `WorkspaceSymbolParams` and `SymbolInformation` to `server/types.go`.
    - [ ] **Handler**: Implement `handleWorkspaceSymbol` in `server/workspace_symbol.go`.
    - [ ] **Logic**: Iterate through all files in the `CompilationUnit`, find symbols matching the query, and return them as a list of `SymbolInformation`.
    - [ ] **Dispatcher**: Wire up `workspace/symbol` in `server.go`.
    - [ ] **Test**: Add an integration test for workspace symbols.

- [ ] **6. `textDocument/rename`**
    - [ ] **Types**: Add `RenameParams` and `WorkspaceEdit` to `server/types.go`.
    - [ ] **Handlers**: Implement `handlePrepareRename` and `handleRename` in `server/rename.go`.
    - [ ] **Logic**: Use the "Find All References" logic to get all locations of a symbol and create a `WorkspaceEdit` containing `TextEdit`s to perform the rename.
    - [ ] **Dispatcher**: Wire up `textDocument/prepareRename` and `textDocument/rename` in `server.go`.
    - [ ] **Test**: Add an integration test for the rename feature.

## LSP features

## Gemini
Note: This feature implementation is not complete yet.
More steps will be added above until the lsp implementation is complete.
When this section is removed then the implement has been completed.
