package server

func (s *Server) Load(path string) ([]byte, error) {
	uri, err := ParseURI(path)
	if err != nil {
		s.log.Printf("failed to parse uri: '%s'", path)
		return nil, err
	}

	content, ok := s.documents[uri]
	if !ok {
		pathAbsLocal := uri.String()
		if uri.Scheme == "file" {
			pathAbsLocal = uri.Path
		}
		s.log.Printf("file(fs): %s", path)
		return s.fsFileLoader.Load(pathAbsLocal)
	}
	s.log.Printf("file(mem): %s", uri.Path)
	return []byte(content), nil
}
