package command

import _ "embed"

//go:embed grammar.txt
var GrammarUsage string

//go:embed check.txt
var CheckUsage string

//go:embed diff.txt
var DiffUsage string

//go:embed fmt.txt
var FmtUsage string

//go:embed lsp.txt
var LspUsage string

//go:embed print.txt
var PrintUsage string
