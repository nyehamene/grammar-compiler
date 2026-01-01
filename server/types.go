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

// DocumentSymbolParams represents the parameters of a `textDocument/documentSymbol` request.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SymbolKind represents the kind of a symbol.
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// DocumentSymbol represents a symbol in a text document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Tags           []int            `json:"tags,omitempty"` // SymbolTag
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// WorkspaceSymbolParams represents the parameters of a `workspace/symbol` request.
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

// SymbolInformation represents a symbol in the workspace.
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Tags          []int      `json:"tags,omitempty"` // SymbolTag
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

// PrepareRenameParams represents the parameters of a `textDocument/prepareRename` request.
type PrepareRenameParams = TextDocumentPositionParams

// RenameParams represents the parameters of a `textDocument/rename` request.
type RenameParams struct {
	TextDocumentPositionParams
	NewName string `json:"newName"`
}

// WorkspaceEdit represents a collection of text edits to be applied to documents.
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes,omitempty"`
}

// Completion
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

type CompletionTriggerKind int

const (
	Invoked                         CompletionTriggerKind = 1
	TriggerCharacter                CompletionTriggerKind = 2
	TriggerForIncompleteCompletions CompletionTriggerKind = 3
)

type CompletionItem struct {
	Label            string             `json:"label"`
	Kind             CompletionItemKind `json:"kind,omitempty"`
	Detail           string             `json:"detail,omitempty"`
	Documentation    *MarkupContent     `json:"documentation,omitempty"`
	InsertText       string             `json:"insertText,omitempty"`
	InsertTextFormat InsertTextFormat   `json:"insertTextFormat,omitempty"`
}

type CompletionItemKind int

const (
	TextCompletion          CompletionItemKind = 1
	MethodCompletion        CompletionItemKind = 2
	FunctionCompletion      CompletionItemKind = 3
	ConstructorCompletion   CompletionItemKind = 4
	FieldCompletion         CompletionItemKind = 5
	VariableCompletion      CompletionItemKind = 6
	ClassCompletion         CompletionItemKind = 7
	InterfaceCompletion     CompletionItemKind = 8
	ModuleCompletion        CompletionItemKind = 9
	PropertyCompletion      CompletionItemKind = 10
	UnitCompletion          CompletionItemKind = 11
	ValueCompletion         CompletionItemKind = 12
	EnumCompletion          CompletionItemKind = 13
	KeywordCompletion       CompletionItemKind = 14
	SnippetCompletion       CompletionItemKind = 15
	ColorCompletion         CompletionItemKind = 16
	FileCompletion          CompletionItemKind = 17
	ReferenceCompletion     CompletionItemKind = 18
	FolderCompletion        CompletionItemKind = 19
	EnumMemberCompletion    CompletionItemKind = 20
	ConstantCompletion      CompletionItemKind = 21
	StructCompletion        CompletionItemKind = 22
	EventCompletion         CompletionItemKind = 23
	OperatorCompletion      CompletionItemKind = 24
	TypeParameterCompletion CompletionItemKind = 25
)

type InsertTextFormat int

const (
	PlainTextTextFormat InsertTextFormat = 1
	SnippetTextFormat   InsertTextFormat = 2
)

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionOptions represents the server's completion capabilities.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   *bool    `json:"resolveProvider,omitempty"`
}

// DiagnosticOptions options for providing diagnostics.
type DiagnosticOptions struct {
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics,omitempty"`
	InterFileDependencies bool `json:"interFileDependencies,omitempty"`
}

// DocumentDiagnosticParams parameters for `textDocument/diagnostic` request.
type DocumentDiagnosticParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentDiagnosticReportKind is the kind of diagnostic report.
type DocumentDiagnosticReportKind string

const (
	// DocumentDiagnosticReportKindFull is a full diagnostic report.
	DocumentDiagnosticReportKindFull DocumentDiagnosticReportKind = "full"
	// DocumentDiagnosticReportKindUnchanged is an unchanged diagnostic report.
	DocumentDiagnosticReportKindUnchanged DocumentDiagnosticReportKind = "unchanged"
)

// RelatedFullDocumentDiagnosticReport is a full diagnostic report for a document.
type RelatedFullDocumentDiagnosticReport struct {
	Kind  DocumentDiagnosticReportKind `json:"kind"` // "full"
	Items []Diagnostic                 `json:"items"`
}

// DocumentLinkOptions is the server capability for document link.
type DocumentLinkOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

// DocumentLinkParams is the parameter for `textDocument/documentLink`.
type DocumentLinkParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentLink is a link to a document.
type DocumentLink struct {
	Range   Range       `json:"range"`
	Target  DocumentUri `json:"target,omitempty"`
	Tooltip string      `json:"tooltip,omitempty"`
	Data    any         `json:"data,omitempty"`
}

// DocumentHighlightOptions is the server capability for document highlight.
type DocumentHighlightOptions struct{}

// DocumentHighlightParams is the parameter for `textDocument/documentHighlight`.
type DocumentHighlightParams = TextDocumentPositionParams

// DocumentHighlight is a highlight of a symbol in a document.
type DocumentHighlight struct {
	Range Range                  `json:"range"`
	Kind  *DocumentHighlightKind `json:"kind,omitempty"`
}

// DocumentHighlightKind is the kind of a document highlight.
type DocumentHighlightKind int

const (
	// Text is a textual occurrence.
	Text DocumentHighlightKind = 1
	// Read is a read-access.
	Read DocumentHighlightKind = 2
	// Write is a write-access.
	Write DocumentHighlightKind = 3
)
