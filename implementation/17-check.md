# Type checker

- Implement a type checker

## Implementation

- Every grammar is a namespace.
- The language has the following builtin primitive types:

  1. String. The type of every string literal. Ex. "foo".
  2. Regexp. The type of every regular expression literal. Ex. /[a-z]/.

- A rule defines a variable of type `Production` whose value can be any valid expression.
- A rule must have a value. It is an error if no value is provided.
- A binding defines a variable of type `Namespace`.
- Rules are exported and are accessible from external namespaces via binding variables.
- The `@import` directive is used to import a namespace and bind it to a variable.
- The import path is resolved relative to the directory containing the importing namespace.
- It is an error if the imported file does not exist.
- Rules defined in an imported namespace are accessible through the binding variable.
- It is an error to access a non-existent rule.
- It is an error if 2 or more declarations have the same name (identifier).
  A rule cannot have the same name as another rule in the same file.
  A rule cannot have the same name as a binding variable in the same file.

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


