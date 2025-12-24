# Diff nodes

Compare the ast node from input files and visually highlight
their differences.

- Implement `grammar diff --ast PATH PATH`.


## Update diff output

Diffing testdata/node1.grammar and testdata/node2.grammar should
produce:

```
- 1:1   RuleDecl: rule1
+ 1:1   RuleDecl: rule2
```
