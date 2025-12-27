package check

import (
	"strings"
	"testing"
)

func TestCheckSuccess(t *testing.T) {
	checker := NewChecker()
	err := checker.Check("../testdata/check/success/a.grammar")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
}

func TestCheckNonExistentImport(t *testing.T) {
	checker := NewChecker()
	err := checker.Check("../testdata/check/nonexistent_import/a.grammar")
	if err == nil {
		t.Fatal("Expected an error, but got none.")
	}

	if !strings.Contains(err.Error(), "could not load imported namespace") {
		t.Errorf("Expected error to contain 'could not load imported namespace', but got: %v", err)
	}
}

func TestCheckImportCycle(t *testing.T) {
	checker := NewChecker()
	err := checker.Check("../testdata/check/cycle/a.grammar")
	if err == nil {
		t.Fatal("Expected an error for import cycle, but got none.")
	}

	if !strings.Contains(err.Error(), "import cycle detected") {
		t.Errorf("Expected error to contain 'import cycle detected', but got: %v", err)
	}
}