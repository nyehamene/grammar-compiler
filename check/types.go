package check

// Type represents a type in the grammar language.
type Type interface {
	String() string
	// A dummy method to make the interface unique.
	isType()
}

// Basic types
type (
	StringType     struct{}
	RegexpType     struct{}
	ProductionType struct{}
	ExternalType   struct{}
)

func (StringType) String() string     { return "string" }
func (RegexpType) String() string     { return "regexp" }
func (ProductionType) String() string { return "production" }
func (ExternalType) String() string   { return "external" }

func (StringType) isType()     {}
func (RegexpType) isType()     {}
func (ProductionType) isType() {}
func (ExternalType) isType()   {}

// NamespaceType represents an imported grammar file.
type NamespaceType struct {
	Name string // The name of the namespace (the file path).
}

func (n *NamespaceType) String() string { return "namespace " + n.Name }
func (n *NamespaceType) isType()        {}

var (
	String     = StringType{}
	Regexp     = RegexpType{}
	Production = ProductionType{}
	External   = ExternalType{}
)
