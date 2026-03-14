package check

import (
	"strings"
	"testing"
)

func TestCheckUnusedPrivateRule(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/unused_private_rule.grammar")
	errList, ok := checker.CompilationUnit().Errors["../testdata/check/unused/unused_private_rule.grammar"]
	if !ok || len(errList) == 0 {
		t.Fatal("Expected errors, but got none.")
	}

	foundExpectedError := false
	for _, e := range errList {
		if strings.Contains(e.Message, "unused symbol: privateRule") {
			foundExpectedError = true
			break
		}
	}

	if !foundExpectedError {
		t.Errorf("Expected error to contain 'unused symbol: privateRule', but got: %v", errList)
	}
}

func TestCheckUnusedBinding(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/unused_binding.grammar")
	errList, ok := checker.CompilationUnit().Errors["../testdata/check/unused/unused_binding.grammar"]
	if !ok || len(errList) == 0 {
		t.Fatal("Expected errors, but got none.")
	}

	foundExpectedError := false
	for _, e := range errList {
		if strings.Contains(e.Message, "unused symbol: b") {
			foundExpectedError = true
			break
		}
	}

	if !foundExpectedError {
		t.Errorf("Expected error to contain 'unused symbol: b', but got: %v", errList)
	}
}

func TestCheckUsedPrivateRule(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/used_private_rule.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/unused/used_private_rule.grammar")
	if err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
}

func TestCheckUnusedPublicRule(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/unused_public_rule.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/unused/unused_public_rule.grammar")
	if err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
}
