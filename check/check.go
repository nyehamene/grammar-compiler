package check

import (
	"fmt"
	"grammar/ast"
)

// Checker holds the state for the type-checking process.
type Checker struct {
	cu     *CompilationUnit
	Errors ast.ErrorList
}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{
		cu: NewCompilationUnit(),
	}
}

// Check initiates the checking process for a given path.
func (c *Checker) Check(path string) error {
	// Placeholder for file checking logic. In the next steps, this will
	// read the file, parse it, and perform semantic analysis.
	fmt.Printf("Checking file: %s\n", path) // Placeholder output
	return nil
}

// CheckSource initiates the checking process for a given source content.
func (c *Checker) CheckSource(content []byte, path string) error {
	// Placeholder for source content checking logic.
	fmt.Printf("Checking source from: %s\n", path) // Placeholder output
	return nil
}

// Err returns the collected errors, or nil if there are none.
func (c *Checker) Err() error {
	if len(c.Errors) == 0 {
		return nil
	}
	// In the future, this will return a formatted error list.
	return c.Errors
}

