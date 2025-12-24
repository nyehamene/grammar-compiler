# Print tokens

Color outputs lines as follows:

Line:Col     KIND     LEXEME
----------------------------
1:1         ident     `foo`
2:12        string    `"bar"`
3:10        invalid   `-`       <-- color this line red to highlight invalid token
10:1        newline   `\n`

