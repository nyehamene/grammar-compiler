package server

// Position in a text document expressed as zero-based line and character offset.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// A range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// A textual edit applicable to a text document.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// TextDocumentIdentifier is a light-weight descriptor for a text document.
type TextDocumentIdentifier struct {
	URI DocumentUri `json:"uri"`
}
