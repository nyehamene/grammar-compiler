# Comprehensive Test Plan for Directory-as-Package System

**Objective:** To thoroughly validate the implementation of the new Directory-as-Package system, ensuring correct package and module resolution, accurate symbol completion, and appropriate diagnostic reporting for both errors and deprecated functionalities (specifically, file-based imports).

**Test Categories:**

1.  **Unit Tests (`check/package_test.go`, `check/fileloader_test.go`):**
    *   **Focus:** Isolated testing of core package/module functionality, including `@package` directive parsing, package name validation (consistent names, mismatch errors, inference), module and package type structures, and basic package loading (e.g., handling non-existent directories, empty directories).
    *   **Specifics:** Verify correct module name derivation from filenames, path normalization, and protection against path traversal.
    *   **Relevance to gaps:** Ensures foundational package and file system logic is sound before integration.

2.  **Integration Tests (`check/package_integration_test.go`):**
    *   **Focus:** Testing the combined behavior of package loading and symbol resolution across multiple modules within the same package and between different packages via `@import` directives.
    *   **Specifics:** Validate `pkg.Module.rule` member access, detect and report circular package dependencies, and handle error cases for non-existent packages, modules, or rules.
    *   **Relevance to gaps:** Verifies end-to-end package resolution.

3.  **LSP Server Tests (`server/package_lsp_test.go` and specific LSP feature tests like `server/completion_test.go`, `server/diagnostics_test.go`, etc.):**
    *   **Focus:** Crucially, these tests will validate the Language Server Protocol's understanding and interaction with the new package system.
    *   **Completion (`server/completion_test.go`):**
        *   **Key Gap Coverage:** Test suggestions for module names after a package reference (`pkg.`), rule names after a module reference (`pkg.module_a.`), and package directories within `@import("")`.
    *   **Diagnostics (`server/diagnostics_test.go`):**
        *   **Key Gap Coverage:** Validate the reporting of:
            *   Errors for package name mismatches between modules in the same directory.
            *   Errors for accessing non-existent packages, modules, or rules.
            *   **Warning diagnostics for deprecated file-based imports (`@import("file.grammar")`)**.
    *   **Definition, References, Hover, Rename, Document Symbol, Document Link (`server/definition_test.go`, `server/references_test.go`, `server/hover_test.go`, `server/rename_test.go`, `server/document_symbol_test.go`, `server/document_link_test.go`):**
        *   **Key Gap Coverage:** Ensure these core LSP features correctly navigate, resolve, and update symbols across the new package and module boundaries. This includes definition lookup across packages, finding references in importing files, displaying correct type information on hover, and successfully renaming rules or modules that update references in other files. Document symbols should correctly identify `@package` directives and use "Module" as the kind. Document links should correctly point to directories for package imports and files for deprecated file imports (with a deprecation tooltip).

**Test Data:**
The plan requires a dedicated set of `testdata` structures to cover various scenarios, including:
*   `testdata/packages/basic/`: Simple package with multiple modules.
*   `testdata/packages/mismatch/`: Package with conflicting `@package` names for error testing.
*   `testdata/packages/empty/`: An empty directory to test edge cases.
*   `testdata/packages/nested/`: Nested package structures.
*   `testdata/lsp/packages/mypackage/` and `testdata/lsp/packages/lib/`: Comprehensive LSP-specific test files to simulate package imports and inter-package rule access, specifically designed for completion, diagnostics, and other LSP feature validation.

**Execution:**
The proposed tests should be integrated into the project's testing suite, runnable via `make test` or more granular `go test` commands targeting specific packages and test functions as detailed in the `test-plan-36-directory-as-package.md`.

---

## Detailed Test Cases

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

### 4. Cycle Detection Tests (`check/package_test.go`)

#### 4.1 Import Cycles
- [ ] Test circular package dependencies are detected
- [ ] Test self-reference through package is handled

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

### 6. Integration Tests (`check/package_integration_test.go`)

#### 6.1 Full Package Workflow
- [ ] Test loading a package with multiple modules
- [ ] Test accessing rules across modules in same package
- [ ] Test accessing rules across packages via @import

#### 6.2 Error Cases
- [ ] Test error for non-existent package directory
- [ ] Test error for package name mismatch
- [ ] Test error for accessing non-existent module

### 7. LSP Server Tests (`server/package_lsp_test.go`)

#### 7.1 Package Diagnostics
- [ ] Test diagnostics show package name mismatch error
- [ ] Test diagnostics show deprecation warning for file-based import

#### 7.2 Completion Tests (`server/completion_test.go`)

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

**Test Cases:**
- [ ] **Rename rule in package**: Rename `rule_to_rename` in module_a.grammar - should update reference in module_b.grammar
- [ ] **Rename module**: Rename `module_a` to `module_c` - should update all package member accesses (`module_a.` becomes `module_c.`)
- [ ] **Rename across packages**: Rename rule in lib package - should update references in app package that imports it

#### 7.8 Document Symbol Tests (`server/document_symbol_test.go`)

**Test Cases:**
- [ ] **Document symbol shows @package**: @package directive appears in document outline
- [ ] **Document symbol shows module kind**: Shows "Module" instead of deprecated "Namespace"

#### 7.9 Document Link Tests (`server/document_link_test.go`)

**Test Cases:**
- [ ] **Document link for @package**: Hovering on `@package("pkg")` in module.grammar should show link to pkg/ directory
- [ ] **Document link for package import**: Hovering on `@import("pkg")` in main.grammar should show link to pkg/ directory
- [ ] **Document link for deprecated file import**: Hovering on `@import("old.grammar")` should show link to old.grammar with deprecation tooltip

### 8. Snapshot Tests

**Objective**: To prevent regressions in LSP responses and diagnostic outputs by capturing and comparing expected output.

**Areas to Cover**:
- [ ] **LSP Completion Responses**: Capture `CompletionItem` arrays for various completion scenarios.
- [ ] **LSP Diagnostics**: Capture `Diagnostic` arrays for different error and warning conditions (including package name mismatches and deprecated file import warnings).
- [ ] **LSP Hover Responses**: Capture formatted hover content.
- [ ] **LSP Definition Responses**: Capture `Location`/`LocationLink` for go-to-definition.
- [ ] **LSP References Responses**: Capture `Location` arrays for find-all-references.
- [ ] **LSP Document Symbol Responses**: Capture `DocumentSymbol` arrays for document outline.

**Implementation Notes**:
- The existing `testutil/snapshot.go` utility is suitable for this purpose, providing `AssertSnapshotJSON` and `AssertSnapshotText` for comparison and update modes.
- Snapshots should be updated when intended behavior changes.
- Ensure timestamps and other dynamic data are handled deterministically or ignored during comparison.
- This will be crucial for maintaining the quality of LSP features as the package system evolves.
