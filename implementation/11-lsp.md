# Language server

Language server implementation.

The language server protocol implementation code and API should be placed in 'server' directory.

- Create a `Message`, `RequestMessage`, `ResponseMessage`, and `NotificationMessage` types.
- Add the `ResponseError` and `ErrorCodes` types.
- Implement a JSON RPC encoder and decoder and save it to the files `server/rpc.go`.
- Add basic lsp server type and update the lsp command to call its start method.

- Add a `DocumentUri` parser type that parses a uri string as illustrated below

   foo://example.com:8042/over/there?name=ferret#nose
  \_/   \______________/\_________/ \_________/ \__/
   |           |            |            |        |
scheme     authority       path        query   fragment
   |   _____________________|__
  / \ /                        \
  urn:example:animal:ferret:nose

  The type should have `scheme`, `authority`, `path`, `query` and `fragement` fields.
  Add a String method that formats the fields into a valid uri string.
  Save the code to `server/uri.go`.

## Gemini
Note: This feature implementation is not complete yet.
More steps will be added above until the lsp implementation is complete.
When this section is removed then the implement has been completed.
