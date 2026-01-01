package server

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
