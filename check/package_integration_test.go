package check

import (
	"grammar/log"
	"os"
	"path/filepath" // Added import
	"strings"
	"testing"
)

func TestIntegrationFullPackageWorkflow(t *testing.T) {
	// Test loading a package with multiple modules
	t.Run("loading package with multiple modules", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker := NewChecker(cu, log.NewConsoleLogger(os.Stderr, log.INFO))
		// Use the testdata/packages/basic setup
		checker.Check("../testdata/packages/basic/A.grammar") // Load one file, it should pull in the package
		
		err := cu.Err("../testdata/packages/basic/A.grammar")
		if err != nil {
			t.Fatalf("Checker.Check() failed: %v", err)
		}

		pkgPath := "../testdata/packages/basic"
		pkg := cu.Packages[NormalizePath(pkgPath)]
		if pkg == nil {
			t.Fatalf("Expected package %q to be loaded", pkgPath)
		}

		if len(pkg.Modules) != 2 {
			t.Errorf("Expected 2 modules in package %q, got %d", pkgPath, len(pkg.Modules))
		}
		if pkg.Modules["A"] == nil || pkg.Modules["B"] == nil {
			t.Error("Expected modules A and B in the package")
		}
	})

	// Test accessing rules across modules in same package
	t.Run("accessing rules across modules in same package", func(t *testing.T) {
		// testdata/packages/cross_module_access
		// 		mypackage/module_a.grammar: @package("mypackage"); rule_a = "a";
		// 		mypackage/module_b.grammar: @package("mypackage"); rule_b = module_a.rule_a;
		_ = os.MkdirAll("../testdata/packages/cross_module_access/mypackage", 0755)
		_ = os.WriteFile("../testdata/packages/cross_module_access/mypackage/module_a.grammar", []byte(`@package("mypackage"); rule_a = "a";`), 0644)
		_ = os.WriteFile("../testdata/packages/cross_module_access/mypackage/module_b.grammar", []byte(`@package("mypackage"); rule_b = module_a.rule_a;`), 0644)
		defer os.RemoveAll("../testdata/packages/cross_module_access")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker := NewChecker(cu, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker.Check("../testdata/packages/cross_module_access/mypackage/module_b.grammar")
		
		err := cu.Err(NormalizePath("../testdata/packages/cross_module_access/mypackage/module_b.grammar"))
		if err != nil {
			t.Fatalf("Checker.Check() failed: %v", err)
		}

		modB := cu.Modules[NormalizePath("../testdata/packages/cross_module_access/mypackage/module_b.grammar")]
		if modB == nil {
			t.Fatal("Expected module_b to be loaded")
		}

		// Check if rule_b's type was resolved correctly, indicating successful access to module_a.rule_a
		_, ok := modB.Types["rule_b"]
		if !ok {
			t.Error("Expected rule_b to be resolvable and have a type")
		}
	})

	// Test accessing rules across packages via @import
	t.Run("accessing rules across packages via @import", func(t *testing.T) {
		// testdata/packages/cross_package_import
		// 		lib/utils.grammar: @package("lib"); commonRule = "common";
		// 		app/main.grammar: @package("app"); libPkg = @import("../lib"); appRule = libPkg.utils.commonRule;
		_ = os.MkdirAll("../testdata/packages/cross_package_import/lib", 0755)
		_ = os.MkdirAll("../testdata/packages/cross_package_import/app", 0755)
		_ = os.WriteFile("../testdata/packages/cross_package_import/lib/utils.grammar", []byte(`@package("lib"); commonRule = "common";`), 0644)
		_ = os.WriteFile("../testdata/packages/cross_package_import/app/main.grammar", []byte(`@package("app"); libPkg = @import("../lib"); appRule = libPkg.utils.commonRule;`), 0644)
		defer os.RemoveAll("../testdata/packages/cross_package_import")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker := NewChecker(cu, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker.Check("../testdata/packages/cross_package_import/app/main.grammar")
		
		err := cu.Err(NormalizePath("../testdata/packages/cross_package_import/app/main.grammar"))
		if err != nil {
			t.Fatalf("Checker.Check() failed: %v", err)
		}

		modMain := cu.Modules[NormalizePath("../testdata/packages/cross_package_import/app/main.grammar")]
		if modMain == nil {
			t.Fatal("Expected main.grammar to be loaded")
		}

		// Check if appRule's type was resolved correctly
		_, ok := modMain.Types["appRule"]
		if !ok {
			t.Error("Expected appRule to be resolvable and have a type")
		}
	})
}

func TestIntegrationErrorCases(t *testing.T) {
	// Test error for non-existent package directory
	t.Run("non-existent package directory error", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker := NewChecker(cu, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker.Check("../testdata/packages/nonexistent_package_dir/main.grammar")
		
		err := cu.Err("../testdata/packages/nonexistent_package_dir/main.grammar")
		if err == nil {
			t.Fatal("Expected error for non-existent package directory, but got none.")
		}
		if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "cannot load package") {
			t.Errorf("Expected 'no such file or directory' or 'cannot load package' error, got: %v", err)
		}
	})

	// Test error for package name mismatch
	t.Run("package name mismatch error", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker := NewChecker(cu, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker.Check("../testdata/packages/mismatch/A.grammar") // Loading A.grammar should trigger mismatch with B.grammar
		
		err := cu.Err("../testdata/packages/mismatch/A.grammar")
		if err == nil {
			t.Fatal("Expected package name mismatch error, but got none.")
		}
		if !strings.Contains(err.Error(), "package name mismatch") {
			t.Errorf("Expected 'package name mismatch' error, got: %v", err)
		}
	})

	// Test error for accessing non-existent module
	t.Run("accessing non-existent module error", func(t *testing.T) {
		// testdata/packages/nonexistent_module_access
		// 		mypackage/module_a.grammar: @package("mypackage"); rule_a = "a";
		// 		main.grammar: @package("main"); pkg = @import("mypackage"); result = pkg.nonexistent_module.rule_a;
		_ = os.MkdirAll("../testdata/packages/nonexistent_module_access/mypackage", 0755)
		_ = os.WriteFile("../testdata/packages/nonexistent_module_access/mypackage/module_a.grammar", []byte(`@package("mypackage"); rule_a = "a";`), 0644)
		_ = os.WriteFile("../testdata/packages/nonexistent_module_access/main.grammar", []byte(`@package("main"); pkg = @import("./mypackage"); result = pkg.nonexistent_module.rule_a;`), 0644)
		defer os.RemoveAll("../testdata/packages/nonexistent_module_access")

		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker := NewChecker(cu, log.NewConsoleLogger(os.Stderr, log.INFO))
		checker.Check("../testdata/packages/nonexistent_module_access/main.grammar")
		
		err := cu.Err("../testdata/packages/nonexistent_module_access/main.grammar")
		if err == nil {
			t.Fatal("Expected error for non-existent module access, but got none.")
		}
		if !strings.Contains(err.Error(), "module 'nonexistent_module' not found in package 'mypackage'") {
			t.Errorf("Expected 'module not found' error, got: %v", err)
		}
	})
}

// Helper to normalize paths for OS compatibility
func NormalizePath(p string) string {
	return filepath.Clean(strings.ReplaceAll(p, "/", string(os.PathSeparator)))
}
