# Diff nodes

Compare ast nodes from input files display their differences.
Only different nodes should be displayed.

Use a preorder traversal to flatten each ast tree into a list of tokens
but presever the indentation level for each node. The indentation level
can be preserved to using a second list and store the indentation level
For example, node at index `n` in the node list will have its indentation
level stored in the indentation list at index `n`.

After flattening, compare the flatten list then print only the different
nodes apply the corresponding indentation level.

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
