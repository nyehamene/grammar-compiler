package check

import (
	"os"
	"path/filepath"
)

// FileLoader is an interface for loading file contents.
type FileLoader interface {
	Load(path string) ([]byte, error)
	LoadDir(path string) ([]string, error)
	IsDir(path string) (bool, error)
	NormalizePath(path string) (string, error)
	SetWorkspaceRoot(path string)
}

// FileSystemFileLoader implements FileLoader to read files from the OS filesystem.
type FileSystemFileLoader struct {
	workspaceRoot string
}

func (l *FileSystemFileLoader) Load(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (l *FileSystemFileLoader) LoadDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".grammar" {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files, nil
}

func (l *FileSystemFileLoader) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (l *FileSystemFileLoader) NormalizePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// If workspace root is set, ensure the path doesn't escape it
	if l.workspaceRoot != "" {
		absRoot, err := filepath.Abs(l.workspaceRoot)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(absRoot, absPath)
		if err != nil {
			return "", err
		}
		// Check for path traversal (starts with ..)
		if filepath.IsAbs(rel) || rel == ".." || len(rel) >= 3 && rel[0:3] == ".."+string(filepath.Separator) {
			return "", filepath.ErrBadPattern
		}
	}

	return absPath, nil
}

func (l *FileSystemFileLoader) SetWorkspaceRoot(path string) {
	l.workspaceRoot = path
}
