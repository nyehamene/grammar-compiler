# Diff tokens

Compare the tokens from input files and visually highlight
their differences.

- Implement `grammar diff --tokens PATH PATH`.


## Diff tokens table format


For example, diffing ../testdata/file1.grammar and ../testdata/file2.grammar
should produce the following table


```txt
  Line:Col   KIND            LEXEME
  ----------------------------------
- 1:1        IDENT           doc1
+ 1:1        IDENT           doc2
  1:6        ASSIGN          =
  1:8        STRING          "test"
  1:14       SEMICOLON       ;
  1:15       EOF
```

using `-` to indicate lines in ../testdata/file1.grammar but not in the other.
And `+` lines in ../testdata/file2.grammar not in the other file.
