package entity

// TableDefinition holds the parsed representation of a SQL CREATE TABLE statement.
type TableDefinition struct {
	Name        string
	Columns     []Column
	PrimaryKeys []string
	Indexes     []Index
	ForeignKeys []ForeignKey
}

// Column represents a single column in a table.
type Column struct {
	Name       string
	SQLType    string
	GoType     string
	GoName     string // PascalCase
	IsPrimary  bool
	IsUnique   bool
	IsNotNull  bool
	HasDefault bool
	Default    string
	IsNullable bool
	IsSerial   bool
	MaxLength  int
	Precision  int
	Scale      int
}

// Index represents a table index.
type Index struct {
	Name    string
	Columns []string
	Unique  bool
}

// ForeignKey represents a foreign key constraint.
type ForeignKey struct {
	Column     string
	RefTable   string
	RefColumn  string
	OnDelete   string
	OnUpdate   string
}

// GeneratorOptions controls what code is generated.
type GeneratorOptions struct {
	ModulePath string
	Package    string
	ORM        string // gorm | sqlc | raw
	WithTests  bool
	WithBench  bool
	Driver     string // postgres | mysql | sqlite | sqlserver
}

// GeneratedFile represents a single generated source file.
type GeneratedFile struct {
	Path    string
	Content string
}

// GenerationResult groups all files produced for a table.
type GenerationResult struct {
	Entity         GeneratedFile
	Interface      GeneratedFile
	Implementation GeneratedFile
	Tests          []GeneratedFile
	Benchmarks     []GeneratedFile
	Migration      GeneratedFile
	SQLCConfig     GeneratedFile
	SQLCQuery      GeneratedFile
}
