# Parser

Parse input files and produces and AST. If an error is found
the program report the errors and exit with a non-zero exit code.

The implementation should provide proper error reporting.

See ../grammar.txt for for details.

- Implement a recursive descent parser that is invoked with
  `grammar check <PATH>`.

- The parser should get input tokens from the tokenizer implemented
  previously.

- The parser should return an ast node for each input file.

- Add tests to parse the grammar files in ../example
