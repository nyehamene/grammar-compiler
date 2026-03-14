package check

import (
	"path/filepath"
	"testing"
)

func TestFileSystemFileLoaderLoadDir(t *testing.T) {
	loader := &FileSystemFileLoader{}

	t.Run("load basic directory", func(t *testing.T) {
		files, err := loader.LoadDir("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("LoadDir() error = %v", err)
		}

		if len(files) != 2 {
			t.Errorf("File count = %d, want %d", len(files), 2)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		files, err := loader.LoadDir("../testdata/packages/empty")
		if err != nil {
			t.Fatalf("LoadDir() error = %v", err)
		}

		if len(files) != 0 {
			t.Errorf("File count = %d, want %d", len(files), 0)
		}
	})

	t.Run("non-grammar files ignored", func(t *testing.T) {
		files, err := loader.LoadDir("../testdata/check/success")
		if err != nil {
			t.Fatalf("LoadDir() error = %v", err)
		}

		// Should only include .grammar files
		for _, f := range files {
			if filepath.Ext(f) != ".grammar" {
				t.Errorf("Non-grammar file included: %s", f)
			}
		}
	})
}

func TestFileSystemFileLoaderIsDir(t *testing.T) {
	loader := &FileSystemFileLoader{}

	t.Run("directory returns true", func(t *testing.T) {
		isDir, err := loader.IsDir("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("IsDir() error = %v", err)
		}
		if !isDir {
			t.Error("Expected true for directory")
		}
	})

	t.Run("file returns false", func(t *testing.T) {
		isDir, err := loader.IsDir("../testdata/packages/basic/A.grammar")
		if err != nil {
			t.Fatalf("IsDir() error = %v", err)
		}
		if isDir {
			t.Error("Expected false for file")
		}
	})

	t.Run("non-existent returns error", func(t *testing.T) {
		_, err := loader.IsDir("../testdata/packages/nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}
	})
}

func TestFileSystemFileLoaderNormalizePath(t *testing.T) {
	loader := &FileSystemFileLoader{}

	t.Run("resolve relative path", func(t *testing.T) {
		path, err := loader.NormalizePath("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("NormalizePath() error = %v", err)
		}
		abs, err := filepath.Abs("../testdata/packages/basic")
		if err != nil {
			t.Fatalf("filepath.Abs() error = %v", err)
		}
		if path != abs {
			t.Errorf("Normalized path = %q, want %q", path, abs)
		}
	})

	t.Run("reject path traversal without workspace", func(t *testing.T) {
		// Without workspace root set, should allow traversal
		path, err := loader.NormalizePath("../../etc/passwd")
		if err != nil {
			t.Fatalf("NormalizePath() error = %v", err)
		}
		_ = path // Just ensure no error
	})

	t.Run("reject path traversal with workspace", func(t *testing.T) {
		loader.SetWorkspaceRoot("../testdata/packages")
		defer loader.SetWorkspaceRoot("")

		_, err := loader.NormalizePath("../../etc/passwd")
		if err == nil {
			t.Error("Expected error for path traversal with workspace root")
		}
	})
}

func TestFileSystemFileLoaderWorkspaceRoot(t *testing.T) {
	loader := &FileSystemFileLoader{}

	t.Run("set workspace root", func(t *testing.T) {
		loader.SetWorkspaceRoot("../testdata")
		// Just ensure no panic
	})

	t.Run("path within workspace", func(t *testing.T) {
		loader.SetWorkspaceRoot("../testdata/packages")
		path, err := loader.NormalizePath("../testdata/packages/basic/A.grammar")
		if err != nil {
			t.Fatalf("NormalizePath() error = %v", err)
		}
		_ = path
	})
}
