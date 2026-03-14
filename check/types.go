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

// NamespaceType represents an imported grammar file (deprecated, use ModuleType)
type NamespaceType struct {
	Name string // The name of the namespace (the file path).
}

func (n *NamespaceType) String() string { return "namespace " + n.Name }
func (n *NamespaceType) isType()        {}

// ModuleType represents a module (a single .grammar file)
type ModuleType struct {
	Name string // The name of the module (the file path).
}

func (m *ModuleType) String() string { return "module " + m.Name }
func (m *ModuleType) isType()        {}

// PackageType represents a package (a directory containing .grammar files)
type PackageType struct {
	Name    string          // The package name (as declared by @package or inferred from directory)
	Path    string          // The package directory path
	Modules map[string]Type // Modules in the package (names to ModuleType)
}

func (p *PackageType) String() string { return "package " + p.Name }
func (p *PackageType) isType()        {}

var (
	String     = StringType{}
	Regexp     = RegexpType{}
	Production = ProductionType{}
	External   = ExternalType{}
)
