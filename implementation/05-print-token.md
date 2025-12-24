# Print token

Print tokens in the tabular format below.

Line:Col     KIND     LEXEME
----------------------------
1:1         ident     `foo`
2:12        string    `"bar"`
3:10        invalid   `-`
10:1        newline   `\n`

The lexeme should be properly escaped. For example,
The newline character `\n` should be printed as `"\n"`.
A string literal should be printed `"foo"`.

Use colored output to highlight invalid tokens.

- Implement `grammar print --tokens <PATH>`.
