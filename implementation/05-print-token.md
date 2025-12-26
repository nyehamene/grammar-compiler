# Print token

Print tokens in the tabular format below.

Line:Col     KIND     LEXEME
----------------------------
1:1         ident     foo
2:12        string    "bar"
3:10        invalid   -
10:1        newline   \n

The lexeme should be properly escaped. The newline escape
character should be escaped and printed as `"\n"`.

- Implement `grammar print --tokens <PATH>`.
- Preserve newlines and comments.
