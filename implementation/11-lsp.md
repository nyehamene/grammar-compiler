# Language server

Language server implementation.

The language server protocol implementation code and API should be placed in 'server' directory.

- Create a `Message`, `RequestMessage`, `ResponseMessage`, and `NotificationMessage` types.
- Add the `ResponseError` and `ErrorCodes` types.
- Implement a JSON RPC encoder and decoder and save it to the files `server/rpc.go`.
- Add basic lsp server type and update the lsp command to call its start method.

- Add a `DocumentUri` parser type that parses a uri string as illustrated below
  (Implemented in `server/uri.go`)

## Gemini
Note: This feature implementation is not complete yet.
More steps will be added above until the lsp implementation is complete.
When this section is removed then the implement has been completed.
