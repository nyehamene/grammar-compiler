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

func (m *MockFileLoader) LoadDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".grammar") {
			files = append(files, path+"/"+e.Name())
		}
	}
	return files, nil
}

func (m *MockFileLoader) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (m *MockFileLoader) NormalizePath(path string) (string, error) {
	return path, nil
}

func (m *MockFileLoader) SetWorkspaceRoot(path string) {}

func setupTestChecker(t *testing.T) *Checker {
	logger := log.NewConsoleLogger(os.Stderr, log.INFO)
	fileLoader := &MockFileLoader{}
	cu := NewCompilationUnit(fileLoader, logger)
	return NewChecker(cu, logger)
}

func TestCheckSuccess(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/success/a.grammar")
	if err := checker.CompilationUnit().Err("../testdata/check/success/a.grammar"); err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

func TestCheckNonExistentImport(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/nonexistent_import/a.grammar")
	errList, ok := checker.CompilationUnit().Errors["../testdata/check/nonexistent_import/a.grammar"]
	if !ok || len(errList) == 0 {
		t.Fatal("Expected errors, but got none.")
	}

	foundExpectedError := false
	for _, e := range errList {
		if strings.Contains(e.Message, "could not load imported namespace") && !e.Warning {
			foundExpectedError = true
			break
		}
	}

	if !foundExpectedError {
		t.Errorf("Expected error to contain 'could not load imported namespace', but got: %v", errList)
	}
}

func TestCheckImportCycle(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/cycle/a.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/cycle/a.grammar")
	if err == nil {
		t.Fatal("Expected an error for import cycle, but got none.")
	}

	if !strings.Contains(err.Error(), "import cycle detected") {
		t.Errorf("Expected error to contain 'import cycle detected', but got: %v", err)
	}
}

func TestCheckRedeclaration(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/redeclaration/a.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/redeclaration/a.grammar")
	if err == nil {
		t.Fatal("Expected an error for redeclaration, but got none.")
	}

	if !strings.Contains(err.Error(), "redeclared in this namespace") {
		t.Errorf("Expected error to contain 'redeclared in this namespace', but got: %v", err)
	}
}

func TestCheckUndefinedMember(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/undefined_member/a.grammar")
	errList, ok := checker.CompilationUnit().Errors["../testdata/check/undefined_member/a.grammar"]
	if !ok || len(errList) == 0 {
		t.Fatal("Expected errors, but got none.")
	}

	foundExpectedError := false
	for _, e := range errList {
		if strings.Contains(e.Message, "undefined member") && !e.Warning {
			foundExpectedError = true
			break
		}
	}

	if !foundExpectedError {
		t.Errorf("Expected error to contain 'undefined member', but got: %v", errList)
	}
}
