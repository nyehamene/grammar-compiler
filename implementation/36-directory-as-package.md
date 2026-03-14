# Directory-as-Package System

## Goal

Change the package system from treating each file as a package to treating a
directory as a package, where each file within the directory is a module.

## Background

Currently, each `.grammar` file is treated as a namespace/package.
Imports use `@import("file.grammar")` to import an entire file's definitions.

In the new system:
- A **directory** is a **package**
- Each **file** in a directory is a **module**
- **Terminology change**: Replace "Namespace" with "Module" throughout
  - Previously: Namespace = file
  - Now: Module = file (the term Namespace should be deprecated and replaced with Module)
- A module uses `@package("name")` to:
  1. Declare the package name (must be consistent across all modules in the directory)
  2. Return a reference to the package for accessing other modules
- `@import` supports:
  1. File-based import (deprecated): `@import("file.grammar")`
  2. Package-based import: `@import("path/to/package")` (directory)
- If a module has no `@package` directive:
  - If other modules exist in the same package, infer name from them
  - If no other modules exist, use directory name as package name

## Syntax Changes

### New Grammar

```
declaration = binding | package_directive | Rule | comment;
binding     = ident "=" import_directive ";";
package_directive = "@package" "(" string ")";
import_directive = "@import" "(" string ")";
```

### @package Directive Semantics

The `@package("name")` directive serves two purposes:

1. **Declares the package name**: All modules in the same directory must declare the same package name. It's an error if two modules in the same directory declare different package names.

2. **Returns a package reference**: The value can be used to access other modules in the package.
   ```grammar
   pkg :: @package("foo");  // declares package name "foo" and returns package reference
   bb :: pkg.B.b;           // access module B's rule b
   ```

### @import Directive Semantics

The `@import("path")` directive supports two types:

1. **File-based import** (deprecated): `@import("file.grammar")` - imports a single file as a module
2. **Package-based import**: `@import("directory")` - imports a directory as a package

### Path Resolution Rules

1. **Relative paths are resolved from the directory containing the current file**
2. `@package("name")` - "name" is the package name (not a path), must match other modules in directory
3. `@import("subdir")` - refers to a subdirectory (package import)
4. `@import("file.grammar")` - refers to a file (file-based import, deprecated)
5. **Security**: Path traversal attacks should be prevented by:
   - Rejecting paths that escape the workspace root
   - Normalizing paths and checking prefix before resolution

### Module Naming Convention

- Module name is derived from the filename **without** the `.grammar` extension
- Case-sensitive: `A.grammar` creates module `A`, `a.grammar` creates module `a`
- Files without `.grammar` extension should be ignored
- If two files in the same package have the same name (ignoring extension), it's an error

### Package Name Inference

If a module has no `@package` directive:
- If other modules exist in the same package that have `@package` declarations, use the first declared name
- If no other modules have `@package` declarations, use the directory name as the package name
- A module can have zero or more `@package` directives (all must declare the same name)

### Example

Given the package structure:
```
project/
  foo/
    - A.grammar
    - B.grammar
  bar/
    - X.grammar
    - Y.grammar
  root.grammar
```

**Module A declares package and accesses Module B:**
```grammar
// foo/A.grammar
pkg :: @package("foo");  // declares package name "foo"

// access B.b (B.grammar defines a rule named b)
bb :: pkg.B.b;
```

**Module B also declares the same package:**
```grammar
// foo/B.grammar
pkg :: @package("foo");  // must declare same name as A.grammar

// access A.a
aa :: pkg.A.a;
```

**Module A accesses bar/X via package import:**
```grammar
// foo/A.grammar
ext :: @import("../bar");  // import package "bar"

// ext is a package, X is a module in that package, x is a rule in X.grammar
xx :: ext.X.x;
```

**File-based import (deprecated):**
```grammar
// root.grammar
ext :: @import("foo/A.grammar");  // deprecated: imports A.grammar as a module

// ext is a module, access rule 'rule_name' from A.grammar
x :: ext.rule_name;
```

**Module with inferred package name:**
```grammar
// foo/A.grammar (no @package directive)
// Since B.grammar declares @package("foo"), this module is also in package "foo"
bb :: @package("foo").B.b;  // can still use @package to get package reference
```

## Implementation Plan

- [ ] **1. Grammar Update (`grammar.txt`)**
    - [ ] Add `package_directive` production (`@package("name")`)
    - [ ] Keep `binding` with `import_directive` for `@import` (both file and package based)
    - [ ] Update `declaration` to include `package_directive`
    - [ ] Add `import_directive` semantics for both file-based and package-based imports

- [ ] **2. AST Changes (`ast/ast.go`)**
    - [ ] **Terminology**: Rename `Namespace` to `Module` throughout (or add `Module` as alias)
    - [ ] Add `PackageName string` field to `File` to track declared package name
    - [ ] The existing `DirectiveExpr` can be reused for `@package` - just need to handle its semantics
    - [ ] Update `BindingDecl` to support both file-based and package-based imports:
        - Add `Kind BindingKind` (enum: `FileImport`, `PackageImport`)
    - [ ] Add `PackageType` to the type system

- [ ] **3. Tokenizer Changes (`token/tokenizer.go`)**
    - [ ] Add new token type for `@package` directive (or reuse directive token)
    - [ ] Ensure `@import` can distinguish between file path and directory path during semantic analysis

- [ ] **4. Parser Changes (`ast/parser.go`)**
    - [ ] `@package` is parsed as a `DirectiveExpr` (already supported)
    - [ ] Update `parseDecl()` to handle `package_directive` as declaration
    - [ ] The parser just captures the directive; semantic analysis handles package name validation
    - [ ] Update `parseBinding()` to track whether import is file-based or package-based

- [ ] **5. Formatter Changes (`ast/formatter.go`)**
    - [ ] Format `@package` directive like other directives
    - [ ] Update import formatting to indicate file vs package import (for deprecation warning)

- [ ] **6. Printer Changes (`ast/printer.go`)**
    - [ ] Print `@package` directive
    - [ ] Distinguish file-based vs package-based imports in output

- [ ] **7. Diff Changes (`cmd/diff/diff.go`)**
    - [ ] Handle `@package` directive diffs
    - [ ] Handle file-based vs package-based import diffs

- [ ] **8. Walk Changes (`ast/walk.go`)**
    - [ ] Walk `@package` directive (already handled by DirectiveExpr walk)

- [ ] **9. Type System Updates (`check/types.go`)**
    - [ ] Add `PackageType` struct
        - `Name string` - the package name (as declared by `@package`)
        - `Path string` - the package directory path
        - `Modules map[string]*Module` - modules in the package
    - [ ] Rename `NamespaceType` to `ModuleType` (or create ModuleType as alias)
    - [ ] **Terminology change**:
        - `Module` = represents a single `.grammar` file (formerly "Namespace")
        - `Package` = a directory containing one or more `.grammar` files

- [ ] **10. Module Updates (`check/namespace.go`)**
    - [ ] Rename `Namespace` struct to `Module` (or add Module alias)
    - [ ] Add `PackageName string` field to track declared package name
    - [ ] Add `PackagePath string` field to track which package the module belongs to
    - [ ] Create `Package` struct to represent a directory/package
        - `Name string` - package name (from `@package` directive or directory name)
        - `Path string`
        - `Modules map[string]*Module`

- [ ] **11. File Loader Updates (`check/fileloader.go`)**
    - [ ] Add `LoadDir(path string) ([]string, error)` method to list grammar files in a directory
    - [ ] Add `IsDir(path string) (bool, error)` method to check if path is a directory
    - [ ] Add `NormalizePath(path string) (string, error)` method to resolve and validate paths
    - [ ] Add `SetWorkspaceRoot(path string)` method to set the root for security checks
    - [ ] Implement path traversal protection: reject paths that resolve outside workspace root

- [ ] **12. Compilation Unit Updates (`check/compilation.go`)**
    - [ ] Rename `Namespace` field to `Module` (or add Module)
    - [ ] Add `Packages map[string]*Package` to cache loaded packages
    - [ ] Implement `LoadPackage(path string) (*Package, error)`:
        - List all `.grammar` files in the directory
        - Validate each file can be loaded (no duplicate module names)
        - Load each file as a module within the package
        - **Package name resolution**:
          1. Collect all `@package("name")` declarations from modules in the directory
          2. If all declared names are consistent, use that name
          3. If inconsistent declarations exist, report error
          4. If no declarations exist, use directory name as package name
        - Build module name to module mapping
    - [ ] Update `LoadFile()` to detect if a file is part of a package:
        - If file has `@package` directive, load its parent directory as a package
    - [ ] Handle package vs file-based import resolution
    - [ ] Update cycle detection to handle package cycles
    - [ ] **Self-reference handling**: When accessing `pkg.Module` where module is current file, exclude current file from package members

- [ ] **13. Semantic Analysis Updates (`check/check.go`)**
    - [ ] Add `PackageSymbol` to `SymbolKind`
    - [ ] Update symbol collection to handle `@package` directive (via DirectiveExpr)
    - [ ] **Package name validation**:
        - Collect `@package("name")` declarations from all modules in a directory
        - If different names declared, report error: "package name mismatch: 'foo' vs 'bar'"
    - [ ] Update type checking for `@package` directive:
        - Returns `PackageType` with module references
    - [ ] Update type checking for `@import`:
        - If path ends with `.grammar`: file-based import (deprecated)
        - If path is a directory: package-based import
        - For file-based import: emit deprecation warning suggesting module import via package
    - [ ] Handle package name inference: use directory name if no `@package` declared

- [ ] **14. Definition Resolution (`check/definition.go`)**
    - [ ] Update `Resolve()` to handle `PackageType` member access
    - [ ] For `pkg.Module.rule`, resolve:
        1. `pkg` to `PackageType`
        2. `Module` to a module in that package
        3. `rule` to a rule in that module

- [ ] **15. Reference Resolution (`check/references.go`)**
    - [ ] Update to track references through package accesses

- [ ] **16. CLI Updates (`cmd/check/check.go`)**
    - [ ] When given a directory, treat it as workspace root
    - [ ] When given a file, use its parent directory as workspace root
    - [ ] Handle both file and directory arguments
    - [ ] Pass workspace root to file loader for path validation

- [ ] **17. LSP Server Updates (`server/`)**
    - [ ] **Document Symbol (`server/documentsymbol.go`)**:
        - Include `@package` directives as symbols
        - Show "Module" instead of "Namespace" in symbol kind
    - [ ] **Hover (`server/hover.go`)**: Show package/module information on hover
    - [ ] **Definition (`server/definition.go`)**: Navigate to definitions in packages
    - [ ] **References (`server/references.go`)**: Find all references including through packages
    - [ ] **Rename (`server/rename.go`)**: Update references when renaming across packages
    - [ ] **Completion (`server/completion.go`)**:
        - Suggest package modules after package reference
        - Suggest `@package` directive
    - [ ] **Diagnostics (`server/diagnostics.go`)**:
        - Report errors for invalid packages
        - Report errors for package name mismatch

- [ ] **18. Test Updates**
    - [ ] Add test files for new package system in `testdata/`
    - [ ] Update existing tests for backward compatibility or document breaking changes

## Backward Compatibility

- Legacy `@import("file.grammar")` continues to work but shows deprecation warning
- Files without `@package` directive get package name inferred from directory
- The term "Namespace" in error messages should be changed to "Module" (or both shown)
- Existing code without any directives continues to work unchanged

## Migration Path

1. Add `@package("name")` directive support (backward compatible - old code works)
2. Add deprecation warning for file-based `@import("file.grammar")`
3. Add package name validation (error on mismatch)
4. Update tooling to understand packages
5. Update terminology from "Namespace" to "Module" in user-facing messages
6. Consider making file-based import a hard error in a future version

## Error Cases to Handle

- Package directory does not exist
- Package directory has no `.grammar` files
- Circular package dependencies
- Invalid package path (relative path traversal) - reject paths that escape workspace root
- Module name conflicts within a package (two files with same name)
- **Package name mismatch**: two modules in same directory declare different package names
- Accessing non-existent module in package
- Accessing non-existent rule in module
- Path contains `..` that escapes the project root
- Self-reference: module tries to access itself via package (should exclude self from package members)
