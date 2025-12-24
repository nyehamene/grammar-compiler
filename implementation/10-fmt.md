# Formatter

Formats input files inplace. If an error is found
the file containing the error is not modified, and
the program reports the errors and exit with a non-zero
exit code after formatting all the input files.

- Implement `grammar fmt PATH`


## Formating rules

1. Rules not separated by a blank line should be aligned
  so that the value separator `=` are on the same column.

  For example,

  Before formatting:
  ```grammar
  a = "a";
  bar = "bar";

  b = "a";
  name = "name";
  ```

  After formatting:
  ```grammar
  a   = "a";
  bar = "bar";

  b    = "a";
  name = "name";
  ```

2. Add a space around values inside grouping symbols
  like `(` and `)`, `[` and `]`, `{` and `}`, etc.

For example,

Before formatting:
```grammmar
one = (foo);
two = [bar];
xxx = {baz};
```

After formatting:
```grammar
one = ( foo );
two = [ bar ];
xxx = { baz }
```

3. Add a space around the alternative separator `|`.

  Before formatting:
  ```grammar
  foo = a|b;
  ```

  After formatting:
  ```grammar
  foo = a | b;
  ```

4. When arranging alternative vertically, align the
  value separator `=`, alternative separator `|`
  and rule separator `;` on the same column.

  Before formatting:
  ```grammar
  choice = one
     | two
     | three ;
  ```

  After formatting:
  ```grammar
  choice = one
         | two
         | three
         ;
  ```
