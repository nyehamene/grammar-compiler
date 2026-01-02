package check

import "grammar/token"

// SymbolKind represents the kind of a symbol (e.g., rule, binding).
type SymbolKind int

const (
	RuleSymbol SymbolKind = iota
	BindingSymbol
)

// Symbol represents a declared symbol in a file.
type Symbol struct {
	Name     string
	Kind     SymbolKind
	IsPublic bool
	Pos      token.Pos
	IsUsed   bool
}

// SymbolTable holds the symbols for a single file.
type SymbolTable struct {
	Symbols     map[string]*Symbol
	PublicRules []*Symbol
	Filepath    string
}

// NewSymbolTable creates a new symbol table for a file.
func NewSymbolTable(filepath string) *SymbolTable {
	return &SymbolTable{
		Symbols:     make(map[string]*Symbol),
		PublicRules: []*Symbol{},
		Filepath:    filepath,
	}
}

// Add adds a symbol to the table.
func (st *SymbolTable) Add(symbol *Symbol) {
	st.Symbols[symbol.Name] = symbol
	if symbol.Kind == RuleSymbol && symbol.IsPublic {
		st.PublicRules = append(st.PublicRules, symbol)
	}
}

// Find returns the symbol with the given name.
func (st *SymbolTable) Find(name string) (*Symbol, bool) {
	s, found := st.Symbols[name]
	return s, found
}
