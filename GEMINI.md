## Project Overview

This project is a Go-based tool for a custom grammar language. The language is used to describe the syntax of other programming languages. The tool provides a checker, a formatter, and a language server for this grammar language.

The project is structured into several packages:
- `cmd`: Contains the implementation of the command-line tool and its subcommands.
- `token`: Contains the tokenizer for the grammar language.
- `ast`: Will contain the abstract syntax tree representation and the parser.
- `example`: Contains example grammar files.

## Building and Running

The project uses a `Makefile` for common tasks.

- **Build:** To build the `grammar` binary, run:
  ```sh
  make build
  ```

- **Test:** To run the tests, run:
  ```sh
  make test
  ```

- **Run:** To run the `grammar` tool, run:
  ```sh
  make run ARGS="<arguments>"
  ```
  For example, to print the help message:
  ```sh
  make run ARGS="--help"
  ```

- **Clean:** To clean the build artifacts, run:
  ```sh
  make clean
  ```

## Development Conventions

The project follows standard Go conventions. The implementation is guided by the markdown files in the `implementation` directory. Each file in the `implementation` directory describes a stage of the implementation.

## Project Status

### Completed Stages
- Skeleton
- Version
- Help
- Tokenizer
- Print token
- Diff token
- Parser

### Pending Stages
- Print ast
- Diff ast
- Formatter
- Language server
- Treesitter
- Print token (color)
- Print tokens (serial number)
