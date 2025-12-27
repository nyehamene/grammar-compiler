package check

import "os"

// FileLoader is an interface for loading file contents.
type FileLoader interface {
	Load(path string) ([]byte, error)
}

// FileSystemFileLoader implements FileLoader to read files from the OS filesystem.
type FileSystemFileLoader struct{}

func (l *FileSystemFileLoader) Load(path string) ([]byte, error) {
	return os.ReadFile(path)
}
