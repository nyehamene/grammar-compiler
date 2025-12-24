# Diff tokens

Compare tokens from input files and display their differences.

- Implement `grammar diff --tokens PATH PATH`.

## Diff tokens table format update

- Diffing ../testdata/file1.grammar and ../testdata/file2.grammar
  should produce:

  ```
    Line:Col   KIND            LEXEME
    ----------------------------------
  - 1:1        IDENT           doc1
  + 1:1        IDENT           doc2
  ```

- Diffing ../testdata/file1.grammar and ../testdata/file1.grammar
  should produce:

  ```
  ```
