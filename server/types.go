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

// TextDocumentPositionParams represents parameters for requests that require a text document and a position inside it.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// HoverParams represents the parameters of a `textDocument/hover` request.
type HoverParams struct {
	TextDocumentPositionParams
}

// MarkupKind describes the format of a markup string.
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// MarkupContent represents a string that can be rendered as markup.
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// Hover represents the result of a `textDocument/hover` request.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// Location represents a location inside a resource, such as a line
// inside a text file.
type Location struct {
	URI   DocumentUri `json:"uri"`
	Range Range       `json:"range"`
}

// DefinitionParams represents the parameters of a `textDocument/definition` request.
type DefinitionParams = TextDocumentPositionParams

// ReferenceContext contains additional information about the context in which a reference request is made.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// ReferenceParams represents the parameters of a `textDocument/references` request.
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

