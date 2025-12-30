# New Logging System Features

## Goal
To implement a flexible, structured, and configurable logging system for the LSP server, checker, and compilation unit. The system should provide clear, contextual information, especially for LSP message flow, and be easily adaptable for testing and different deployment environments.

## Core Features Expected:

- Add a `Logger` interface to a new `log/` package. The interface should have `Print(any)` and `Printf(format string, v ...any)` methods.
- Provide implementations in a `/server/lsp_loggers.go` file.
- Avoid **Type Explosion**: No new structs should be added just to support or format logs.
- The `Print` method of the logger implementation will use a large type switch to access fields and format logs for specific LSP message types.
- Provide a `silentLogger` that does nothing.
- Provide a `lineLogger` that writes formatted output to an `io.Writer`.
- The `lineLogger` will handle client capabilities by printing them as an indented JSON object.
- Diagnostic messages will be logged one per line *after* they have been sent to the client.

### 1. **Structured LSP Message Logging**
-   **Incoming Requests**: Logged in the format `(->) Request <ID>-<Method>`.
-   **Outgoing Responses**: Logged in the format `(<-) Response <ID>-<Method>`.
    -   For error responses, the log should also clearly indicate the error (e.g., `(<-) Response <ID>-<Method> (Error: <error message>)`).
-   **Incoming Notifications**: Logged in the format `(->) Notification <Method>`.
-   **Outgoing Notifications**: Logged in the format `(<-) Notification <Method>`.

### 2. **Configurable Log Output**
-   Allow the logger to write to any `io.Writer` (e.g., `os.Stderr`, `os.Stdout`, `bytes.Buffer` for testing, or a file).
-   Default logging for the LSP server should still go to `~/.cache/grammar/lsp.log`.

### 3. **Default Silent Logger**
-   When no specific logger output is configured, the system should default to a silent logger that discards all log messages.

### 4. **Standard Internal Logging**
-   The `Printf` method will be used for general-purpose, formatted string logging within components like the checker or file loader.

### 5. **Error Logging**
-   All internal errors (e.g., JSON marshalling/unmarshalling failures) should be logged clearly.

### 6. **Simple and Consistent API**
-   The API should be straightforward: `logger.Print(someObject)` for structured logs and `logger.Printf("message: %s", value)` for simple formatted logs.

### 7. **Diagnostic Logging**
-   Log each diagnostic message individually *after* it has been sent to the client to minimize impact on response time.

## Implementation Notes:
-   The logger should be injected into components (Server, Checker, CompilationUnit) via their constructors or setter methods to maintain dependency inversion.
-   Consider how to differentiate between general internal logging and specific LSP message logging in the API design (e.g., `logger.LogRequest(id, method)` vs `logger.LogInfo("message")`).

## Todos

- [x] **Part 1: Define Logger Abstraction (`log/logger.go`)**
    - [x] Create `log/` directory.
    - [x] Define `Logger` interface with `Print(v any)` and `Printf(format string, v ...any)` methods.

- [x] **Part 2: Implement `Logger` Variations in `server/lsp_loggers.go`**
    - [x] Create `server/lsp_loggers.go` file.
    - [x] Implement `silentLogger` struct with empty `Print` and `Printf` methods.
    - [x] Implement `NewLineLogger` function returning `log.Logger` (an instance of `lineLogger`).
    - [x] Implement `lineLogger` struct.
    - [x] Implement `lineLogger.Printf` method (using `fmt.Fprintf`).
    - [x] Implement `lineLogger.Print(v any)` method with a type switch for:
        - [x] `string`
        - [x] `error`
        - [x] `*RequestMessage` (and handle its `ID` and `Method`)
        - [x] `*ResponseMessage` (and handle its `ID`, `Error`)
        - [x] `*NotificationMessage` (and handle its `Method`)
        - [x] `*InitializeParams` (for capabilities as indented JSON)
        - [x] `*Diagnostic` (for diagnostic logging)

- [x] **Part 3: Refactor `check` Package**
    - [x] Update `check/option.go`: Modify `SetLogger` to accept `grammar/log.Logger`.
    - [x] Update `check/compilation.go`:
        - [x] Update import: replace `stdlog "log"` with `"grammar/log"`.
        - [x] Change `log` field type from `*stdlog.Logger` to `grammar/log.Logger`.
        - [x] Update `NewCompilationUnit` function signature to accept `grammar/log.Logger`.
        - [x] Replace `cu.log.Printf(...)` calls (if any) with `cu.log.Print(fmt.Sprintf(...))`.
    - [x] Update `check/check.go`:
        - [x] Update imports: add `"grammar/log"`, `"io/ioutil"`.
        - [x] Change `log` field type from `*stdlog.Logger` to `grammar/log.Logger`.
        - [x] Update `NewChecker` to initialize logger defaults and apply options.
        - [x] Replace `c.log.Printf(...)` calls (if any) with `c.log.Print(fmt.Sprintf(...))`.

- [x] **Part 4: Refactor `server` Package**
    - [x] Update `server/server.go`:
        - [x] Update imports: `stdlog "log"` and `"grammar/log"`.
        - [x] Change `logger` field type from `*stdlog.Logger` to `grammar/log.Logger`.
        - [x] Update `NewServer` and `NewServerWithLogger` to instantiate and inject the new logger.
        - [x] Replace all existing `s.log.Printf` and `s.log.Println` calls with `s.logger.Printf` or `s.logger.Print`.
        - [x] Adjust `sendErrorResponse` and `sendResponse` arguments for structured logging.
        - [x] Adjust `notify` to use structured logging.
    - [x] Update all handler files (`server/*.go`):
        - [x] Update imports: add `"grammar/log"`, ensure `fmt` is present.
        - [x] Replace `s.log.Printf` calls with `s.logger.Printf(...)`.
        - [x] Update calls to `s.sendErrorResponse` to pass LSP method.
        - [x] Update calls to `s.sendResponse` to pass LSP method.

- [x] **Part 5: Final Verification**
    - [x] Run `make test` to ensure all tests pass.