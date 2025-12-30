package check

import (
	"grammar/log"
	"os"
	"strings"
	"testing"
)

// MockFileLoader implements the FileLoader interface for testing purposes.
type MockFileLoader struct{}

func (m *MockFileLoader) Load(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func setupTestChecker(t *testing.T) *Checker {
	logger := log.NewStderrLogger()
	fileLoader := &MockFileLoader{}
	cu := NewCompilationUnit(fileLoader, logger)
	return NewChecker(cu, logger)
}

func TestCheckSuccess(t *testing.T) {
	checker := setupTestChecker(t)
	err := checker.Check("../testdata/check/success/a.grammar")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

func TestCheckNonExistentImport(t *testing.T) {
	checker := setupTestChecker(t)
	err := checker.Check("../testdata/check/nonexistent_import/a.grammar")
	if err == nil {
		t.Fatal("Expected an error, but got none.")
	}

	if !strings.Contains(err.Error(), "could not load imported namespace") {
		t.Errorf("Expected error to contain 'could not load imported namespace', but got: %v", err)
	}
}

func TestCheckImportCycle(t *testing.T) {
	checker := setupTestChecker(t)
	err := checker.Check("../testdata/check/cycle/a.grammar")
	if err == nil {
		t.Fatal("Expected an error for import cycle, but got none.")
	}

	if !strings.Contains(err.Error(), "import cycle detected") {
		t.Errorf("Expected error to contain 'import cycle detected', but got: %v", err)
	}
}

func TestCheckRedeclaration(t *testing.T) {
	checker := setupTestChecker(t)
	err := checker.Check("../testdata/check/redeclaration/a.grammar")
	if err == nil {
		t.Fatal("Expected an error for redeclaration, but got none.")
	}

	if !strings.Contains(err.Error(), "redeclared in this namespace") {
		t.Errorf("Expected error to contain 'redeclared in this namespace', but got: %v", err)
	}
}

func TestCheckUndefinedMember(t *testing.T) {
	checker := setupTestChecker(t)
	err := checker.Check("../testdata/check/undefined_member/a.grammar")
	if err == nil {
		t.Fatal("Expected an error for undefined member, but got none.")
	}

	if !strings.Contains(err.Error(), "undefined member") {
		t.Errorf("Expected error to contain 'undefined member', but got: %v", err)
	}
}
