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

func TestCheckMultiplePublicRules(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/multiple_public_rules.grammar")
	errList, ok := checker.CompilationUnit().Errors["../testdata/check/unused/multiple_public_rules.grammar"]
	if !ok || len(errList) == 0 {
		t.Fatal("Expected warnings, but got none.")
	}

	warningCount := 0
	foundRule1 := false
	foundRule2 := false
	for _, e := range errList {
		if e.Warning && strings.Contains(e.Message, "more than one public rule in file:") {
			warningCount++
			if strings.Contains(e.Message, "PublicRule1") {
				foundRule1 = true
			}
			if strings.Contains(e.Message, "PublicRule2") {
				foundRule2 = true
			}
		}
	}

	if warningCount != 2 || !foundRule1 || !foundRule2 {
		t.Errorf("Expected two warnings for multiple public rules, but got %d. Errors: %v", warningCount, errList)
	}
}

func TestCheckSinglePublicRule(t *testing.T) {
	checker := setupTestChecker(t)
	checker.Check("../testdata/check/unused/single_public_rule.grammar")
	err := checker.CompilationUnit().Err("../testdata/check/unused/single_public_rule.grammar")
	if err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
}

