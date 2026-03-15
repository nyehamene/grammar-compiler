# Test Plan: Directory-as-Package Feature

## Overview

This test plan covers testing for the Directory-as-Package feature which introduces:
- Package as a directory
- Module as a file within a package
- `@package("name")` directive
- Module type (replacing Namespace)
- Package type

## Test Organization

### Unit Tests: `check/package_test.go`

Test the core package/module functionality in isolation.

### Integration Tests: `check/package_integration_test.go`

Test package loading with file system interactions.

---

## Test Cases

### 1. Package Declaration Tests (`check/package_test.go`)

#### 1.1 @package Directive Parsing
- [ ] Test `@package("foo");` is parsed as a DirectiveExpr
- [ ] Test `@package` can be used as standalone declaration
- [ ] Test `@package` can be used in a binding: `pkg = @package("foo");`

#### 1.2 Package Name Validation
- [ ] Test two modules with same package name in same directory - should pass
- [ ] Test two modules with different package names in same directory - should fail with "package name mismatch" error
- [ ] Test module with no @package directive uses directory name as package name

#### 1.3 Module Type Tests
- [ ] Test Module struct has Package and PackageName fields
- [ ] Test Module can be created with NewModule
- [ ] Test Package struct has Name, Path, Modules fields

#### 1.4 Package Type Tests
- [ ] Test PackageType has Name, Path, Modules fields
- [ ] Test PackageType can be created with module references

---

### 2. Package Loading Tests (`check/package_test.go`)

#### 2.1 LoadPackage Function
- [ ] Test LoadPackage loads all .grammar files in a directory
- [ ] Test LoadPackage returns error for non-existent directory
- [ ] Test LoadPackage returns error for directory with no .grammar files
- [ ] Test LoadPackage caches loaded packages

#### 2.2 Package Name Resolution
- [ ] Test package name is collected from @package declarations
- [ ] Test package name defaults to directory name if no @package
- [ ] Test package name mismatch error is reported

#### 2.3 Module Collection
- [ ] Test each .grammar file becomes a Module
- [ ] Test Module name is filename without .grammar extension
- [ ] Test Modules are keyed by module name in Package.Modules

---

### 3. Import Resolution Tests (`check/package_test.go`)

#### 3.1 File-based Import (Deprecated)
- [ ] Test @import("file.grammar") creates ModuleType
- [ ] Test deprecation warning is emitted for file-based import
- [ ] Test file-based import still works for backward compatibility

#### 3.2 Package-based Import
- [ ] Test @import("directory") creates PackageType
- [ ] Test PackageType has correct Name and Path
- [ ] Test PackageType.Modules contains all modules in package

#### 3.3 Member Access
- [ ] Test pkg.Module.rule resolves correctly
- [ ] Test accessing non-existent module in package produces error
- [ ] Test accessing non-existent rule in module produces error

---

### 4. Cycle Detection Tests (`check/package_test.go`)

#### 4.1 Import Cycles
- [ ] Test circular package dependencies are detected
- [ ] Test self-reference through package is handled

---

### 5. File Loader Tests (`check/fileloader_test.go`)

#### 5.1 Directory Loading
- [ ] Test LoadDir returns all .grammar files in directory
- [ ] Test LoadDir ignores non-.grammar files
- [ ] Test LoadDir returns empty list for directory with no .grammar files

#### 5.2 Path Validation
- [ ] Test IsDir returns true for directories
- [ ] Test IsDir returns false for files
- [ ] Test NormalizePath resolves relative paths
- [ ] Test NormalizePath rejects path traversal (../../etc)

---

### 6. Integration Tests (`check/package_integration_test.go`)

#### 6.1 Full Package Workflow
- [ ] Test loading a package with multiple modules
- [ ] Test accessing rules across modules in same package
- [ ] Test accessing rules across packages via @import

#### 6.2 Error Cases
- [ ] Test error for non-existent package directory
- [ ] Test error for package name mismatch
- [ ] Test error for accessing non-existent module

---

### 7. LSP Server Tests (`server/package_lsp_test.go`)

#### 7.1 Package Diagnostics
- [ ] Test diagnostics show package name mismatch error
- [ ] Test diagnostics show deprecation warning for file-based import

#### 7.2 Completion Tests (`server/completion_test.go`)

**Test Setup:**
```
testdata/lsp/packages/mypackage/
  - module_a.grammar: @package("mypackage"); rule_a = "a";
  - module_b.grammar: @package("mypackage"); rule_b = "b";
testdata/lsp/main.grammar: pkg = @import("mypackage");
```

**Test Cases:**
- [ ] **Package module completion**: After `pkg.` (where pkg is a package), complete with module names from the package
- [ ] **Package member completion**: After `pkg.module_a.`, complete with rule names from module_a
- [ ] **Package directory completion**: After `@import("")`, suggest package directories
- [ ] **Same-package module completion**: After `@package("mypackage").`, suggest module names
- [ ] **Completion in rule body**: Complete with imported package members

#### 7.3 Diagnostics Tests (`server/diagnostics_test.go`)

**Test Cases:**
- [ ] **Package name mismatch**: Two modules in same directory with different @package names - expect error diagnostic
- [ ] **Deprecated file import**: Using @import("file.grammar") - expect warning diagnostic
- [ ] **Missing package**: @import("nonexistent") - expect error diagnostic
- [ ] **Invalid module access**: pkg.NonexistentModule - expect error diagnostic
- [ ] **Invalid rule access**: pkg.Module.nonexistent_rule - expect error diagnostic
- [ ] **Successful package load**: Valid @package and @import - no errors

#### 7.4 Definition Tests (`server/definition_test.go`)

**Test Cases:**
- [ ] **Definition in same package**: Go to definition of rule defined in another module of same package
- [ ] **Definition across packages**: Go to definition of rule in imported package
- [ ] **Definition of @package**: Go to @package directive
- [ ] **Definition of module access**: Go to module in package

#### 7.5 References Tests (`server/references_test.go`)

**Test Cases:**
- [ ] **Find references to rule in same package**: References from other modules in same package
- [ ] **Find references to rule in imported package**: References from importing file
- [ ] **Find references to @package directive**: References to package name declaration

#### 7.6 Hover Tests (`server/hover_test.go`)

**Test Cases:**
- [ ] **Hover on @package**: Shows "package <name>" type
- [ ] **Hover on package binding**: Shows package type with path
- [ ] **Hover on module access**: Shows "module <name>" type
- [ ] **Hover on rule in imported package**: Shows rule type

#### 7.7 Rename Tests (`server/rename_test.go`)

**Test Setup:**
```
pkg/
  - module_a.grammar: @package("pkg"); rule_to_rename = "a";
  - module_b.grammar: @package("pkg"); other_rule = module_a.rule_to_rename;
```

**Test Cases:**
- [ ] **Rename rule in package**: Rename `rule_to_rename` in module_a.grammar - should update reference in module_b.grammar
- [ ] **Rename module**: Rename `module_a` to `module_c` - should update all package member accesses (`module_a.` becomes `module_c.`)
- [ ] **Rename across packages**: Rename rule in lib package - should update references in app package that imports it

#### 7.8 Document Symbol Tests (`server/document_symbol_test.go`)

**Test Cases:**
- [ ] **Document symbol shows @package**: @package directive appears in document outline
- [ ] **Document symbol shows module kind**: Shows "Module" instead of deprecated "Namespace"

#### 7.9 Document Link Tests (`server/document_link_test.go`)

**Test Setup:**
```
pkg/
  - module.grammar: @package("pkg"); rule_m = "m";
main.grammar: pkg = @import("pkg");
old.grammar: old = @import("old.grammar");  // deprecated file import
```

**Test Cases:**
- [ ] **Document link for @package**: Hovering on `@package("pkg")` in module.grammar should show link to pkg/ directory
- [ ] **Document link for package import**: Hovering on `@import("pkg")` in main.grammar should show link to pkg/ directory
- [ ] **Document link for deprecated file import**: Hovering on `@import("old.grammar")` should show link to old.grammar with deprecation tooltip

---

## Test Data

### testdata/packages/basic/
```
basic/
  - A.grammar: @package("basic"); rule_a = "a";
  - B.grammar: @package("basic"); rule_b = "b";
```

### testdata/packages/mismatch/
```
mismatch/
  - A.grammar: @package("foo");
  - B.grammar: @package("bar");
```

### testdata/packages/empty/
```
empty/ (empty directory)
```

### testdata/packages/nested/
```
nested/
  - sub/
    - C.grammar: @package("sub"); rule_c = "c";
```

### testdata/lsp/packages/mypackage/ (for LSP tests)
```
testdata/lsp/packages/mypackage/
  - module_a.grammar: @package("mypackage"); rule_a = "a";
  - module_b.grammar: @package("mypackage"); rule_b = "b";
testdata/lsp/main.grammar: pkg = @import("mypackage"); result = pkg.module_a.rule_a;
```

### testdata/lsp/packages/lib/ (for LSP cross-package tests)
```
testdata/lsp/packages/lib/
  - utils.grammar: @package("lib"); helper = "helper";
testdata/lsp/packages/app/
  - app.grammar: @package("app"); lib = @import("../lib"); use_lib = lib.utils.helper;
```

---

## Running Tests

```bash
# Run all package tests
go test -v ./check/... -run TestPackage

# Run specific test
go test -v ./check/... -run TestLoadPackage

# Run integration tests
go test -v ./server/... -run TestPackage

# Run LSP completion tests
go test -v ./server/... -run TestCompletion

# Run LSP diagnostics tests
go test -v ./server/... -run TestDiagnostics

# Run LSP definition tests
go test -v ./server/... -run TestDefinition

# Run all tests
make test
```
