package generator

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

const entityTmpl = `package {{.Package}}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
)

// {{.GoName}} maps to the {{.TableName}} table.
type {{.GoName}} struct {
{{- range .Columns}}
	{{.GoName}} {{.GoType}} ` + "`" + `gorm:"{{.GORMTag}}" json:"{{.JSONTag}}"` + "`" + `
{{- end}}
}

// TableName implements the gorm.Tabler interface.
func ({{.GoName}}) TableName() string { return "{{.TableName}}" }
`

type entityTmplData struct {
	Package   string
	GoName    string
	TableName string
	Imports   []string
	Columns   []colData
}

type colData struct {
	GoName  string
	GoType  string
	GORMTag string
	JSONTag string
}

// EntityGenerator generates domain entity structs.
type EntityGenerator struct{}

// NewEntityGenerator returns a new EntityGenerator.
func NewEntityGenerator() *EntityGenerator { return &EntityGenerator{} }

// Generate renders the entity file for the given table.
func (g *EntityGenerator) Generate(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	pkg := opts.Package
	if pkg == "" {
		pkg = "entity"
	}

	imports := requiredImports(table)
	sort.Strings(imports)

	cols := make([]colData, len(table.Columns))
	for i, c := range table.Columns {
		cols[i] = colData{
			GoName:  c.GoName,
			GoType:  resolveGoType(c),
			GORMTag: gormTagForColumn(c),
			JSONTag: jsonTagForColumn(c),
		}
	}

	data := entityTmplData{
		Package:   pkg,
		GoName:    toTitle(table.Name),
		TableName: table.Name,
		Imports:   imports,
		Columns:   cols,
	}

	out, err := renderTemplate("entity", entityTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render entity template: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("internal/domain/entity/%s.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}

// resolveGoType returns the Go type string, converting nullable columns to pointer types.
func resolveGoType(c entity.Column) string {
	if c.IsNullable && !c.IsPrimary {
		switch c.GoType {
		case "string", "bool", "int16", "int32", "int64", "float32", "float64":
			return "*" + c.GoType
		case "time.Time":
			return "*time.Time"
		}
	}
	return c.GoType
}

// renderTemplate executes a named template with the given data.
func renderTemplate(name, tmpl string, data any) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// toTitle converts snake_case or any name to PascalCase, singularizing the last word
// so that plural table names (e.g. "users", "order_items") yield proper Go type names.
func toTitle(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	var b strings.Builder
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		if i == len(words)-1 {
			w = singularize(w)
		}
		b.WriteString(strings.ToUpper(w[:1]))
		b.WriteString(strings.ToLower(w[1:]))
	}
	return b.String()
}

// singularize applies basic English singularization rules to a word.
func singularize(s string) string {
	lower := strings.ToLower(s)
	switch {
	case strings.HasSuffix(lower, "ies"):
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(lower, "ses"),
		strings.HasSuffix(lower, "xes"),
		strings.HasSuffix(lower, "ches"),
		strings.HasSuffix(lower, "shes"):
		return s[:len(s)-2]
	case strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss"):
		return s[:len(s)-1]
	}
	return s
}
