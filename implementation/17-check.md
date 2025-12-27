# Type checker

See cli interface in ../command/check.txt.

If PATH argument is a file, the check shoud load the file and its dependencies
(imported namespaces) into a CompilationUnit and perform type checking and
name resolution. (see Loading CompilationUnit section for details about how loading should be handled).

If PATH argument is a directory, the check should load all the grammar files in the directory
along with there dependencies into a CompilationUnit.
(see Loading CompilationUnit section for details about how loading should be handled)

## Goal

- Implement a type checker.

## Types

- The language has the following builtin primitive types:

  1. String. The type of every string literal. Ex. "foo".
  2. Regexp. The type of every regular expression literal. Ex. /[a-z]/.

- A rule defines a variable of type `Production` whose value can be any valid expression.
- A binding defines a variable of type `Namespace`.

## Loading CompilationUnit

Namespace loading requires a compliation unit that maps the absolute path of a grammar
file to a `Namespace`. And is used for resoloving imported namespaces.

A Namespace is a structure that represents the semantic elements of a grammar. Specifically
its binding variables and rule fields. It maps each identifier to a Type.
Every declaration has an identifier. A declaration is bound to its identifier
in the containing/enclosing namespace.

### Loading a file

The check follows these steps to load a file.

1. create a compilation unit if not already created.
2. create a `Namespace` object to represent the file.
2. load each namespace imported in the file into the compilation unit.
3. bind each imported namespace to their identifier in enclosing namespace.
4. bind each rule in the file to their identifier in the enclosing namespace.

### Loading a directory

The check follows these steps to load a directory.

1. create a compilation unit if not already created.
2. create a Namespace object for each file in the directory.
3. start with any of the files (e.g. the first one).
4. load each namespace imported in the chosen file into the compilation unit.
5. bind each imported namespace to their identifier in enclosing namespace.
6. bind each rule in the file to their identifier in the enclosing namespace.

## Symatic Analysis

### Member access

A member access expression has a receiver and an object: `receiver.object`.

- The receiver must resolve to a value of type `Namespace`.
- The object must resolve to a value of type `Production` in the receiver namespace.

## Implementation

- Every grammar is a namespace.
- A rule must have a value. It is an error if no value is provided.
- Rules are exported and are accessible from external namespaces via binding variables.
- The `@import` directive is used to import a namespace and bind it to a variable.
- The import path is resolved relative to the directory containing the importing namespace.
- It is an error if the imported file does not exist.
- Rules defined in an imported namespace are accessible through the binding variable.
- It is an error to access a non-existent rule.
- It is an error if 2 or more declarations have the same name (identifier).
  A rule cannot have the same name as another rule in the same file.
  A rule cannot have the same name as a binding variable in the same file.
- Name checker should detect and report import cycles as errors.

## Example
  For example: given a folder with the following file structure

  ```sh
  pkg/foo.grammar
  bar.grammar
  baz.grammar
  ```

  Where pkg/foo.grammar has the following content:
  ```
  // pkg/foo.grammar
  foo = "i am foo";
  ```

  And, baz.grammar has the following content:
  ```
  // baz.grammar
  baz = "i am baz";
  ```

  And, bar.grammar has content.
  ```
  // bar.grammar
  f = @import("pkg/foo.grammar");

  b = @import("baz.grammar");

  x = @import("xxx.grammar"); // error: "namespace xxx does not exist"

  s = @import("bar.grammar"); // error: "namespace bar cannot import itself"

  ident = f.foo b.baz;

  x1 = f.other; // error: "f.other is undefined. namespace foo has no field other"
  x2 = b.xxx; // error: "b.xxx is undefined. namespace baz has no field xxx"

  ident = "foo"; // error: "field ident redeclared in this namespace. (see other declaration at <INSERT LINE:COLUMN HERE>)"
  ```

## Todos

- [x] **1. Foundational Data Structures**
    - [x] Create `check/types.go` to define the type system (e.g., `Type` interface, `StringType`, `RegexpType`, `ProductionType`, `NamespaceType`).
    - [x] Create `check/namespace.go` to define the `Namespace` struct. It should hold mappings from identifiers to their declarations and types.
    - [x] Create `check/compilation.go` to define the `CompilationUnit` struct. It will manage a cache of loaded namespaces (mapping file paths to `Namespace` objects) and handle file loading logic.

- [x] **2. CLI and Entrypoint**
    - [x] Update `cmd/check/check.go` to properly handle `PATH` arguments (distinguishing between files and directories) and the `--stdin` flag.
    - [x] In `cmd/check/check.go`, create a new `checker.Checker` instance and invoke its main checking method.

- [x] **3. Namespace Loading and Resolution**
    - [x] Implement the `CompilationUnit.LoadFile(path)` method. This method should:
        - Check the cache for the file first.
        - If not cached, read and parse the file to get the AST.
        - Create a new `Namespace` for the file.
        - Add the new `Namespace` to the cache *before* processing imports to allow for cycle detection.
    - [x] In `LoadFile`, iterate through the AST declarations to populate the `Namespace`.
        - For `@import` declarations, recursively call `LoadFile` for the imported path. The import path must be resolved relative to the current file's directory.
        - Handle and report errors for non-existent files.

- [x] **4. Semantic Validation: Name Resolution**
    - [x] While populating the `Namespace` in `LoadFile`, detect and report errors for redeclared identifiers (rules or binding variables with the same name in the same file).
    - [x] Implement import cycle detection within the `CompilationUnit`. `LoadFile` can use a map of files currently in the loading stack to detect a cycle.

- [x] **5. Semantic Validation: Type Checking**
    - [x] Implement a `check(node ast.Node)` method on the checker that traverses the AST.
    - [x] Implement type checking for `MemberExpr` (`receiver.object`):
        - Verify the `receiver` resolves to a `Namespace` type.
        - Verify the `object` exists within that `Namespace` and is a `Production`.
        - Report errors for undefined member access.

- [ ] **6. Diagnostics and Error Reporting**
    - [ ] Ensure all semantic errors (redeclaration, import cycles, file not found, invalid member access) are collected.
    - [ ] In `cmd/check/check.go`, after the check is complete, iterate through the collected errors and print them to `stderr` in the format `path:line:col: message`.
    - [ ] Ensure the command exits with a non-zero status code if any errors are found.

- [ ] **7. Integration with Language Server**
    - [ ] Create a function in the `check` package that can be called from the language server. This function will take the document content and URI.
    - [ ] When a document is opened or changed (`didOpen`, `didChange`), call this checker function.
    - [ ] Convert the checker's errors into LSP `Diagnostic` messages and send them to the client via a `textDocument/publishDiagnostics` notification.



