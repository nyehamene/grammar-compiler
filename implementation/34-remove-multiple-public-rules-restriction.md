# Plan: Remove Multiple Public Rules Restriction

## Goal
To remove the feature that reports a warning when a file declares more than one public rule. This makes the language less restrictive and allows developers to define multiple public symbols in a single file without generating warnings.

## Background
The check for multiple public rules was added as part of the unused symbol analysis (`implementation/32-unused-symbols.md`). This check resides in the `check` package and is surfaced as a warning diagnostic through the language server. This plan outlines the steps to remove this specific check and its related tests.

## Action Plan
- [x] **Update Checker**: Remove the logic for counting and reporting multiple public rules from `check/check.go`.
- [x] **Update Unit Tests**: Remove the corresponding unit tests from `check/unused_test.go` and associated test data.
- [x] **Update Integration Tests**: Remove any integration tests that assert the "multiple public rules" warning.
- [x] **Verify Changes**: Run all tests to ensure the removal did not introduce any regressions.