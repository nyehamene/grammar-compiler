package check

// CheckDocument checks a document for semantic errors and returns a list of them.
// This is the main entry point for the language server.
func CheckDocument(content []byte, path string) (ErrorList, error) {
	checker := NewChecker()
	err := checker.CheckSource(content, path)
	if err != nil {
		if list, ok := err.(ErrorList); ok {
			return list, nil
		}
		return nil, err
	}
	return nil, nil
}
