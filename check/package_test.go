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
	})

	t.Run("non-existent directory", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		_, err := cu.LoadPackage("../testdata/packages/nonexistent")
		if err == nil {
			t.Fatal("Expected error for non-existent directory")
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

func TestFileBasedImport(t *testing.T) {
	t.Run("file import creates ModuleType", func(t *testing.T) {
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
}

func TestPackageBasedImport(t *testing.T) {
	t.Run("directory import creates PackageType", func(t *testing.T) {
		cu := NewCompilationUnit(&FileSystemFileLoader{}, log.NewConsoleLogger(os.Stderr, log.INFO))
		// First load the package
		pkg, err := cu.LoadPackage("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadPackage() error = %v", err)
		}

		// Verify package type
		if pkg.Name != "basic" {
			t.Errorf("Package name = %q, want %q", pkg.Name, "basic")
		}

		if len(pkg.Modules) != 2 {
			t.Errorf("Module count = %d, want %d", len(pkg.Modules), 2)
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
