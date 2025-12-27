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

// TextDocumentItem represents a text document.
type TextDocumentItem struct {
	URI        DocumentUri `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

// DiagnosticSeverity represents the severity of a diagnostic message.
type DiagnosticSeverity int

const (
	SeverityError   DiagnosticSeverity = 1
	SeverityWarning DiagnosticSeverity = 2
	SeverityInfo    DiagnosticSeverity = 3
	SeverityHint    DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// PublishDiagnosticsParams represents the parameters of a `textDocument/publishDiagnostics` notification.
type PublishDiagnosticsParams struct {
	URI         DocumentUri  `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}