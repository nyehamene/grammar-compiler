# Diff nodes

Compare ast nodes from input files display their differences.
Only different nodes should be displayed.

- Implement `grammar diff --ast PATH PATH`.

## Update diff output

- Diffing testdata/node1.grammar and testdata/node2.grammar should
  produce:

  ```
  - 1:1   RuleDecl: rule1
  + 1:1   RuleDecl: rule2
  - 1:9       StringLit: "hello"
  + 1:9       StringLit: "world"
  ```

- Diffing testdata/node1.grammar and testdata/node3.grammar should
  produce:

  ```
  - 1:1   RuleDecl: rule1
  + 1:1   RuleDecl: rule2
  ```

- Diffing testdata/node1.grammar and testdata/node1.grammar should
  produce:

  ```
  ```
