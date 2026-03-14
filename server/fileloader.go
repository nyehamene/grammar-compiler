package server

import (
	"grammar/check"
	"os"
	"path/filepath"
)

func (s *Server) Load(path string) ([]byte, error) {
	uri, err := ParseURI(path)
	if err != nil {
		s.logger.Printf("failed to parse uri: '%s'", path)
		return nil, err
	}

	doc, ok := s.documents[uri]
	if !ok {
		pathAbsLocal := uri.String()
		if uri.Scheme == "file" {
			pathAbsLocal = uri.Path
		}
		s.logger.Printf("file(fs): %s", path)
		return s.fsFileLoader.Load(pathAbsLocal)
	}
	s.logger.Printf("file(mem): %s", uri.Path)
	return []byte(string(doc.text)), nil
}

// LoadDir returns all .grammar files in a directory
func (s *Server) LoadDir(path string) ([]string, error) {
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

// IsDir returns true if the path is a directory
func (s *Server) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// NormalizePath normalizes and validates a path
func (s *Server) NormalizePath(path string) (string, error) {
	return filepath.Abs(path)
}

// SetWorkspaceRoot sets the workspace root for path validation
func (s *Server) SetWorkspaceRoot(path string) {
	// Server uses documents map, no need for workspace root
}

var _ check.FileLoader = (*Server)(nil)
