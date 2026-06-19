package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

const interfaceTmpl = `package {{.Package}}

import (
	"context"
)

// {{.GoName}}Repository defines the persistence contract for {{.GoName}}.
type {{.GoName}}Repository interface {
	Create(ctx context.Context, {{.VarName}} *{{.EntityPkg}}{{.GoName}}) error
	GetByID(ctx context.Context, id {{.IDType}}) (*{{.EntityPkg}}{{.GoName}}, error)
	Update(ctx context.Context, {{.VarName}} *{{.EntityPkg}}{{.GoName}}) error
	Delete(ctx context.Context, id {{.IDType}}) error
	List(ctx context.Context, limit, offset int) ([]*{{.EntityPkg}}{{.GoName}}, error)
}
`

type interfaceTmplData struct {
	Package   string
	GoName    string
	VarName   string
	IDType    string
	EntityPkg string
}

// InterfaceGenerator generates repository interface files.
type InterfaceGenerator struct{}

// NewInterfaceGenerator returns a new InterfaceGenerator.
func NewInterfaceGenerator() *InterfaceGenerator { return &InterfaceGenerator{} }

// Generate renders the repository interface for the given table.
func (g *InterfaceGenerator) Generate(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	pkg := opts.Package
	if pkg == "" {
		pkg = "repository"
	}

	goName := toTitle(table.Name)
	varName := strings.ToLower(goName[:1]) + goName[1:]
	idType := pkIDType(table)

	entityPkg := ""
	if pkg != "entity" {
		entityPkg = "entity."
	}

	data := interfaceTmplData{
		Package:   pkg,
		GoName:    goName,
		VarName:   varName,
		IDType:    idType,
		EntityPkg: entityPkg,
	}

	out, err := renderTemplate("interface", interfaceTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render interface template: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("internal/domain/repository/%s_repository.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}

// pkIDType returns the Go type of the primary key column.
func pkIDType(table *entity.TableDefinition) string {
	for _, c := range table.Columns {
		if c.IsPrimary {
			return c.GoType
		}
	}
	return "int64"
}
