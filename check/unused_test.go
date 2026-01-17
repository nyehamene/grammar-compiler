package check

import (
	"strings"
	"testing"
)

func TestCheckUnusedPrivateRule(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/unused_private_rule.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/unused/unused_private_rule.grammar")
	if err == nil {
		t.Fatal("Expected an error for unused private rule, but got none.")
	}

	if !strings.Contains(err.Error(), "unused symbol: privateRule") {
		t.Errorf("Expected error to contain 'unused symbol: privateRule', but got: %v", err)
	}
}

func TestCheckUnusedBinding(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/unused_binding.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/unused/unused_binding.grammar")
	if err == nil {
		t.Fatal("Expected an error for unused binding, but got none.")
	}

	if !strings.Contains(err.Error(), "unused symbol: b") {
		t.Errorf("Expected error to contain 'unused symbol: b', but got: %v", err)
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


