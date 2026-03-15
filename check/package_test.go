package check

import (
	"grammar/ast"
	"grammar/log"
	"os"
	"strings"
	"testing"
)

func TestPackageDirectiveParsing(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantDecl bool
	}{
		{
			name:     "package directive standalone",
			src:      `@package("foo");`,
			wantDecl: true,
		},
		{
			name:     "package directive with rule",
			src:      `@package("foo"); rule_a = "a";`,
			wantDecl: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cu := NewCompilationUnit(&MockFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
			mod := cu.LoadSource([]byte(tt.src), "test.grammar")

			foundPackage := false
			for _, decl := range mod.File.Decls {
				if _, ok := decl.(*ast.DirectiveExpr); ok {
					foundPackage = true
					break
				}
			}

			if foundPackage != tt.wantDecl {
				t.Errorf("LoadSource() package directive = %v, want %v", foundPackage, tt.wantDecl)
			}
		})
	}
}

func TestPackageNameValidation(t *testing.T) {
	t.Run("same package name in same directory", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		_, err := cu.LoadPackage("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		err = cu.Err("../testdata/packages/basic/A.grammar")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
	})

	t.Run("different package names in same directory", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		_, err := cu.LoadPackage("../testdata/packages/mismatch")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		foundMismatch := false
		for _, errList := range cu.Errors {
			for _, e := range errList {
				if strings.Contains(e.Message, "package name mismatch") {
					foundMismatch = true
					break
				}
			}
		}

		if !foundMismatch {
			t.Error("Expected package name mismatch error")
		}
	})

	t.Run("module with no @package directive, infer from other module", func(t *testing.T) {
		// testdata/packages/inferred_from_other:
		// A.grammar: @package("inferred");
		// B.grammar: (no @package directive)
		_ = os.MkdirAll("../testdata/packages/inferred_from_other", 0755)
		_ = os.WriteFile("../testdata/packages/inferred_from_other/A.grammar", []byte(`@package("inferred"); rule_a = "a";`), 0644)
		_ = os.WriteFile("../testdata/packages/inferred_from_other/B.grammar", []byte(`rule_b = "b";`), 0644)
		defer os.RemoveAll("../testdata/packages/inferred_from_other")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		pkg, err := cu.LoadPackage("../testdata/packages/inferred_from_other")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		if pkg.Name != "inferred" {
			t.Errorf("Expected package name 'inferred', got %q", pkg.Name)
		}

		modA := pkg.Modules["A"]
		if modA == nil {
			t.Fatal("Expected module A")
		}
		if modA.PackageName != "inferred" {
			t.Errorf("Expected module A PackageName 'inferred', got %q", modA.PackageName)
		}

		modB := pkg.Modules["B"]
		if modB == nil {
			t.Fatal("Expected module B")
		}
		if modB.PackageName != "inferred" {
			t.Errorf("Expected module B PackageName 'inferred', got %q", modB.PackageName)
		}
	})

	t.Run("module with no @package directive, infer from directory name", func(t *testing.T) {
		// testdata/packages/infer_from_dir:
		// A.grammar: (no @package directive)
		_ = os.MkdirAll("../testdata/packages/infer_from_dir", 0755)
		_ = os.WriteFile("../testdata/packages/infer_from_dir/A.grammar", []byte(`rule_a = "a";`), 0644)
		defer os.RemoveAll("../testdata/packages/infer_from_dir")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		pkg, err := cu.LoadPackage("../testdata/packages/infer_from_dir")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		if pkg.Name != "infer_from_dir" {
			t.Errorf("Expected package name 'infer_from_dir', got %q", pkg.Name)
		}

		modA := pkg.Modules["A"]
		if modA == nil {
			t.Fatal("Expected module A")
		}
		if modA.PackageName != "infer_from_dir" {
			t.Errorf("Expected module A PackageName 'infer_from_dir', got %q", modA.PackageName)
		}
	})
}

func TestLoadPackage(t *testing.T) {
	t.Run("load basic package", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		pkg, err := cu.LoadPackage("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		if pkg.Name != "basic" {
			t.Errorf("Package name = %q, want %q", pkg.Name, "basic")
		}

		if len(pkg.Modules) != 2 {
			t.Errorf("Module count = %d, want %d", len(pkg.Modules), 2)
		}

		if pkg.Modules["A"] == nil || pkg.Modules["B"] == nil {
			t.Error("Expected modules A and B to be loaded")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		_, err := cu.LoadPackage("../testdata/packages/nonexistent")
		if err == nil {
			t.Fatal("Expected error for non-existent directory")
		}
		if !strings.Contains(err.Error(), "no such file or directory") { // Update error message check
			t.Errorf("Expected 'no such file or directory' error, got: %v", err)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		_, err := cu.LoadPackage("../testdata/packages/empty")
		if err == nil {
			t.Fatal("Expected error for empty directory")
		}
		if !strings.Contains(err.Error(), "no .grammar files") {
			t.Errorf("Expected 'no .grammar files' error, got: %v", err)
		}
	})

	t.Run("package caching", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		pkg1, err := cu.LoadPackage("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		pkg2, err := cu.LoadPackage("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		if pkg1 != pkg2 {
			t.Error("Expected cached package to be returned")
		}
	})
}

func TestPackageNameResolution(t *testing.T) {
	t.Run("explicit package name", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		pkg, err := cu.LoadPackage("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		if pkg.Name != "basic" {
			t.Errorf("Package name = %q, want %q", pkg.Name, "basic")
		}
	})

	t.Run("inferred from directory", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))

		// First load the package which will also load the modules
		dirPath := "../testdata/packages/nested/sub"
		pkg, err := cu.LoadPackage(dirPath)
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		if pkg.Name != "sub" {
			t.Errorf("Package name = %q, want %q", pkg.Name, "sub")
		}

		// Get the module from the package
		mod := pkg.Modules["C"]
		if mod == nil {
			t.Fatal("Expected module C in package")
		}

		        if mod.PackageName != "sub" {
		            t.Errorf("Module package name = %q, want %q", mod.PackageName, "sub")
		        }
		    })
		}
		
		func TestModuleAndPackageStructFields(t *testing.T) {
			cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
			pkg, err := cu.LoadPackage("../testdata/packages/basic")
			if err != nil {
				t.Fatalf("LoadPackage() error = %v", err)
			}
		
			// Test Package struct fields
			if pkg.Name == "" {
				t.Error("Expected Package.Name to be set")
			}
			if pkg.Path == "" {
				t.Error("Expected Package.Path to be set")
			}
			if len(pkg.Modules) == 0 {
				t.Error("Expected Package.Modules to be populated")
			}
		
			// Test Module struct fields
			modA := pkg.Modules["A"]
			if modA == nil {
				t.Fatal("Expected module A")
			}
			if modA.PackageName == "" {
				t.Error("Expected Module.PackageName to be set")
			}
				if modA.Package.Path == "" {
					t.Error("Expected Module.Package.Path to be set")
				}		}
		
		func TestFileBasedImport(t *testing.T) {	t.Run("file import creates ModuleType", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		mod := cu.LoadSource([]byte(`a = @import("b.grammar");`), "../testdata/check/success/a.grammar")

		typ, ok := mod.Types["a"]
		if !ok {
			t.Fatal("Expected type for binding 'a'")
		}

		if _, ok := typ.(*ModuleType); !ok {
			t.Errorf("Expected ModuleType, got %T", typ)
		}
	})

	t.Run("deprecation warning for file import", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		cu.LoadSource([]byte(`a = @import("b.grammar");`), "../testdata/check/success/a.grammar")

		errList, ok := cu.Errors["../testdata/check/success/a.grammar"]
		if !ok || len(errList) == 0 {
			t.Fatal("Expected warnings, but got none.")
		}

		foundWarning := false
		for _, e := range errList {
			if e.Warning && strings.Contains(e.Message, "deprecated") {
				foundWarning = true
				break
			}
		}

		if !foundWarning {
			t.Errorf("Expected deprecation warning, got: %v", errList)
		}
	})

	t.Run("file-based import still works for backward compatibility", func(t *testing.T) {
		_ = os.MkdirAll("../testdata/compatibility", 0755)
		_ = os.WriteFile("../testdata/compatibility/imported.grammar", []byte(`publicRule = "value";`), 0644)
		_ = os.WriteFile("../testdata/compatibility/main.grammar", []byte(`imp = @import("imported.grammar"); myRule = imp.publicRule;`), 0644)
		defer os.RemoveAll("../testdata/compatibility")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		modMain := cu.LoadSource([]byte(`imp = @import("imported.grammar"); myRule = imp.publicRule;`), "../testdata/compatibility/main.grammar")
		
		err := cu.Err("../testdata/compatibility/main.grammar")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Check if the rule from the imported file is resolvable
		_, ok := modMain.Types["myRule"]
		if !ok {
			t.Error("Expected 'myRule' to be resolvable from imported file")
		}
	})
}

func TestPackageBasedImport(t *testing.T) {
	t.Run("directory import creates PackageType", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		// Create a mock main.grammar that imports the basic package
		mockMainGrammar := []byte(`basicPkg = @import("../testdata/packages/basic");`)
		mainModule := cu.LoadSource(mockMainGrammar, "main.grammar")

		typ, ok := mainModule.Types["basicPkg"]
		if !ok {
			t.Fatal("Expected type for binding 'basicPkg'")
		}

		packageType, ok := typ.(*PackageType)
		if !ok {
			t.Fatalf("Expected PackageType, got %T", typ)
		}

		if packageType.Name != "basic" {
			t.Errorf("Expected PackageType Name 'basic', got %q", packageType.Name)
		}
		if !strings.HasSuffix(packageType.Path, "testdata/packages/basic") {
			t.Errorf("Expected PackageType Path to end with 'testdata/packages/basic', got %q", packageType.Path)
		}
		if len(packageType.Modules) != 2 {
			t.Errorf("Expected 2 modules in PackageType, got %d", len(packageType.Modules))
		}
		if packageType.Modules["A"] == nil || packageType.Modules["B"] == nil {
			t.Error("Expected modules A and B in PackageType")
		}
	})

	t.Run("pkg.Module.rule resolves correctly", func(t *testing.T) {
		_ = os.MkdirAll("../testdata/access_rules", 0755)
		_ = os.WriteFile("../testdata/access_rules/mypackage/module_a.grammar", []byte(`@package("mypackage"); myRule = "value";`), 0644)
		_ = os.WriteFile("../testdata/access_rules/main.grammar", []byte(`pkg = @import("mypackage"); result = pkg.module_a.myRule;`), 0644)
		defer os.RemoveAll("../testdata/access_rules")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		modMain := cu.LoadSource([]byte(`pkg = @import("mypackage"); result = pkg.module_a.myRule;`), "../testdata/access_rules/main.grammar")

		err := cu.Err("../testdata/access_rules/main.grammar")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Verify that 'result' is resolvable and has a type
		_, ok := modMain.Types["result"]
		if !ok {
			t.Error("Expected 'result' rule to be resolvable")
		}
	})

	t.Run("accessing non-existent module in package produces error", func(t *testing.T) {
		_ = os.MkdirAll("../testdata/nonexistent_module", 0755)
		_ = os.WriteFile("../testdata/nonexistent_module/mypackage/module_a.grammar", []byte(`@package("mypackage"); myRule = "value";`), 0644)
		_ = os.WriteFile("../testdata/nonexistent_module/main.grammar", []byte(`pkg = @import("mypackage"); result = pkg.nonexistent_module.myRule;`), 0644)
		defer os.RemoveAll("../testdata/nonexistent_module")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		cu.LoadSource([]byte(`pkg = @import("mypackage"); result = pkg.nonexistent_module.myRule;`), "../testdata/nonexistent_module/main.grammar")

		err := cu.Err("../testdata/nonexistent_module/main.grammar")
		if err == nil {
			t.Fatal("Expected error for non-existent module, but got none.")
		}
		if !strings.Contains(err.Error(), "module 'nonexistent_module' not found in package 'mypackage'") {
			t.Errorf("Expected 'module not found' error, got: %v", err)
		}
	})

	t.Run("accessing non-existent rule in module produces error", func(t *testing.T) {
		_ = os.MkdirAll("../testdata/nonexistent_rule", 0755)
		_ = os.WriteFile("../testdata/nonexistent_rule/mypackage/module_a.grammar", []byte(`@package("mypackage"); myRule = "value";`), 0644)
		_ = os.WriteFile("../testdata/nonexistent_rule/main.grammar", []byte(`pkg = @import("mypackage"); result = pkg.module_a.nonexistent_rule;`), 0644)
		defer os.RemoveAll("../testdata/nonexistent_rule")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		cu.LoadSource([]byte(`pkg = @import("mypackage"); result = pkg.module_a.nonexistent_rule;`), "../testdata/nonexistent_rule/main.grammar")

		err := cu.Err("../testdata/nonexistent_rule/main.grammar")
		if err == nil {
			t.Fatal("Expected error for non-existent rule, but got none.")
		}
		if !strings.Contains(err.Error(), "rule 'nonexistent_rule' not found in module 'module_a'") {
			t.Errorf("Expected 'rule not found' error, got: %v", err)
		}
	})
}

func TestImportCycle(t *testing.T) {
	t.Run("circular package dependencies", func(t *testing.T) {
		checker := setupTestChecker(t)
		checker.Check("../testdata/check/cycle/a.grammar")

		err := checker.CompilationUnit().Err("../testdata/check/cycle/a.grammar")
		if err == nil {
			t.Fatal("Expected an error for import cycle, but got none.")
		}

		if !strings.Contains(err.Error(), "import cycle detected") {
			t.Errorf("Expected error to contain 'import cycle detected', but got: %v", err)
		}
	})

	t.Run("self-reference through package is handled", func(t *testing.T) {
		// testdata/self_ref_package/
		// 		mypackage/module_a.grammar: @package("mypackage"); rule_a = pkg.module_a.rule_b; rule_b = "b";
		_ = os.MkdirAll("../testdata/self_ref_package/mypackage", 0755)
		_ = os.WriteFile("../testdata/self_ref_package/mypackage/module_a.grammar", []byte(`@package("mypackage"); rule_a = pkg.module_a.rule_b; rule_b = "b";`), 0644)
		_ = os.WriteFile("../testdata/self_ref_package/mypackage/module_b.grammar", []byte(`@package("mypackage"); rule_c = "c";`), 0644)
		defer os.RemoveAll("../testdata/self_ref_package")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		modA := cu.LoadSource([]byte(`@package("mypackage"); rule_a = pkg.module_a.rule_b; rule_b = "b";`), "../testdata/self_ref_package/mypackage/module_a.grammar")

		err := cu.Err("../testdata/self_ref_package/mypackage/module_a.grammar")
		if err != nil {
			// This scenario should not produce an error if self-reference within the module is allowed.
			// However, if the type checker explicitly disallows pkg.current_module.member access,
			// then an error is expected. For now, assuming it should resolve if valid.
			// The original plan says "exclude current file from package members" in compilation unit updates.
			// This test should assert it either works or errors with a specific message.
			t.Fatalf("Expected no error for self-reference, but got: %v", err)
		}

		// Verify that rule_a can resolve rule_b from the same module via pkg.module_a
		_, ok := modA.Types["rule_a"]
		if !ok {
			t.Error("Expected 'rule_a' to be resolvable")
		}
	})
}

type MockFileLoaderWithDirCheck struct{}

func (m *MockFileLoaderWithDirCheck) Load(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (m *MockFileLoaderWithDirCheck) LoadDir(path string) ([]string, error) {
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

func (m *MockFileLoaderWithDirCheck) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (m *MockFileLoaderWithDirCheck) NormalizePath(path string) (string, error) {
	return path, nil
}

func (m *MockFileLoaderWithDirCheck) SetWorkspaceRoot(path string) {}
