# Hover

Update Hover `textDocument/hover` provide the following output.

## Output format

(type)

value

## Rules

- For a string:
  When a client request hover information for a string literal (e.g. `"John"`),
  the LSP server should return:

  ```markdown
  (string)

  "John"
  ```

- For a regexp:
  When a client request hover information for a regexp literal (e.g. `/[a-z]/`),
  the LSP server should return:

  ```markdown
  (regexp)

  /[a-z]/
  ```

- For a external value:
  When a client request hover information for an external value (e.g. `$STRING`),
  the LSP server should return:

  ```markdown
  (external)

  $STRING
  ```

- For a production:
  When a client requests hover information for `message` on rule 3 in the grammar below:

  ```grammar
  name = "John";           // 1:
  message = "Mr" name;     // 2:
  mail = subject message;  // 3: client request hover information when cursor is on message
  ```

The LSP server should respond with:

  ```markdown
  (production)

  "Mr" name;
  ```

- For a binding:
  When a client requests hover information for `foo` on rule 2 in the grammar below:

  ```grammar
  foo = @import("foo.grammar");  // 1:
  bar = foo.name;                // 2:
  ```

The LSP server should respond with:

  ```markdown
  (namespace)

  foo.grammar;
  ```

  When a client requests hover information for `name` on rule 2 in the grammar above,
  assuming name is defined as `name = "Foo"` inside `foo.grammar`;
  Then LSP server should respond with:

  ```markdown
  (namespace)

  "foo";
  ```

## Todos

### Implementation Plan

#### 1. Update `server/hover.go`
*   **Modify `handleHover` Function**:
    1.  Use `ast.FindNodeAt` to get the specific AST node under the cursor.
    2.  Invoke the type checker to get the semantic type (`check.Type`) of the found node. This is crucial for distinguishing between different kinds of identifiers (e.g., a rule vs. a namespace).
    3.  Implement a `switch` statement on the node's type (`ast.Node`) to handle different hover scenarios.

#### 2. Implement Hover Content Logic
*   **For `ast.StringLit`, `ast.RegexLit`, `ast.ExternalValue`**:
    *   The type is determined directly from the AST node (`string`, `regexp`, `external`).
    *   The value is the node's literal text, extracted from the source file.
    *   Format the hover response as `(type)\n\nvalue`.

*   **For `ast.Ident` (as a rule or binding reference)**:
    *   Use `checker.TypeOf` to determine if it's a `check.ProductionType` or a `check.NamespaceType`.
    *   If **`ProductionType`**:
        1.  Locate the `*ast.RuleDecl` for the identifier within the current namespace.
        2.  Extract the source text of the rule's body.
        3.  Format the hover response as `(production)\n\n<body>;`.
    *   If **`NamespaceType`**:
        1.  Locate the `*ast.BindingDecl`.
        2.  Extract the import path from the declaration.
        3.  Format the hover response as `(namespace)\n\n<path>`.

*   **For `ast.MemberExpr` (member of an imported namespace)**:
    *   The `FindNodeAt` function will target the `Member` part of the expression (`ast.Ident`).
    *   Use `checker.TypeOf` on the `MemberExpr` to get the member's type.
    *   If it's a `ProductionType`, find the `*ast.RuleDecl` in the imported namespace.
    *   Extract the source text of that rule's body.
    *   Format the hover response as `(production)\n\n<body>;`.

#### 3. Create a Source Text Helper
*   Implement a helper function, e.g., `sourceOf(node ast.Node, src []rune) string`, to extract the source text for any given AST node using its `Pos()` and `End()` methods.
*   For rule bodies, which are slices of expressions (`[]ast.Expr`), the helper will calculate the range from the start of the first expression to the end of the last one.

#### 4. Update Tests
*   Modify `server/integration_server_test.go` to add comprehensive tests for the hover feature.
*   Ensure there are test cases for each scenario outlined in `implementation/23-hover.md`: strings, regexes, external values, local productions, imported productions (member access), and namespace bindings.
