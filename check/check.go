package check

import (
	"grammar/ast"
	"grammar/token"
)

// Checker holds the state for the type-checking process.
type Checker struct {
	cu *CompilationUnit
}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{
		cu: NewCompilationUnit(),
	}
}

// Check initiates the checking process for a given path.
func (c *Checker) Check(path string) error {
	_, err := c.cu.LoadFile(path)
	if err != nil {
		if _, isParserError := err.(ast.ErrorList); !isParserError {
			c.cu.AddError(token.NoPos, err.Error())
		}
	}
	return c.cu.Err()
}

// CheckSource initiates the checking process for a given source content.
func (c *Checker) CheckSource(content []byte, path string) error {
	_, err := c.cu.LoadSource(content, path)
	if err != nil {
		if _, isParserError := err.(ast.ErrorList); !isParserError {
			c.cu.AddError(token.NoPos, err.Error())
		}
	}
	return c.cu.Err()
}

// Err returns the collected errors, or nil if there are none.
func (c *Checker) Err() error {
	return c.cu.Err()
}
