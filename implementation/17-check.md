# Type checker

See cli interface in ../command/check.txt.

If PATH argument is a file, the check shoud load the file and its dependencies
(imported namespaces) into a CompilationUnit and perform type checking and
name resolution. (see Loading namepsace section for details about how loading should be handled).

If PATH argument is a directory, the check should load all the grammar files in the directory
along with there dependencies into a CompilationUnit.
(see Loading namepsace section for details about how loading should be handled)

## Goal

- Implement a type checker.

## Types

- The language has the following builtin primitive types:

  1. String. The type of every string literal. Ex. "foo".
  2. Regexp. The type of every regular expression literal. Ex. /[a-z]/.

- A rule defines a variable of type `Production` whose value can be any valid expression.
- A binding defines a variable of type `Namespace`.

## Namespace loading

Namespace loading requires a compliation unit that maps the absolute path of a grammar
file to a `Namespace`. And is used for resoloving imported namespaces.

A Namespace is a structure that represents the semantic elements of a grammar. Specifically
its binding variables and rule fields. It map each identifier a Type.
Every declaration has an identifier. A declaration is bound to its identifier in a namespace.

The check should use the following steps to load a file.

1. create a compilation unit if not already created.
2. create a `Namespace` object to represent the file.
2. load each namespace imported in the file into the compilation unit.
3. bind each imported namespace to their identifier in the file to the Namespace created in step 2.
4. bind each rule in the file to their identifier in the Namespace created in step 2.

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


