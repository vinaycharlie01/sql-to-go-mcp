package generator

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

const sqlcConfigTmpl = `version: "2"
sql:
  - schema: "db/migrations/{{.TableName}}.sql"
    queries: "db/queries/{{.TableName}}.sql"
    engine: "{{.Engine}}"
    gen:
      go:
        package: "{{.Package}}"
        out: "internal/adapters/db/sqlc"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        overrides:
          - db_type: "uuid"
            go_type: "string"
          - db_type: "timestamptz"
            go_type: "time.Time"
          - db_type: "jsonb"
            go_type: "encoding/json.RawMessage"
`

const sqlcQueriesTmpl = `-- name: Create{{.GoName}} :one
INSERT INTO {{.TableName}} (
{{- range $i, $c := .InsertCols}}
    {{$c.Name}}{{if not (last $i $.InsertColsLen)}},{{end}}
{{- end}}
) VALUES (
{{- range $i, $c := .InsertCols}}
    ${{inc $i}}{{if not (last $i $.InsertColsLen)}},{{end}}
{{- end}}
)
RETURNING *;

-- name: Get{{.GoName}}ByID :one
SELECT * FROM {{.TableName}}
WHERE id = $1
LIMIT 1;

-- name: List{{.GoName}}s :many
SELECT * FROM {{.TableName}}
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: Update{{.GoName}} :one
UPDATE {{.TableName}}
SET
{{- range $i, $c := .UpdateCols}}
    {{$c.Name}} = ${{inc $i}}{{if not (last $i $.UpdateColsLen)}},{{end}}
{{- end}}
WHERE id = ${{.WhereArgN}}
RETURNING *;

-- name: Delete{{.GoName}} :exec
DELETE FROM {{.TableName}}
WHERE id = $1;
`

const sqlcAdapterTmpl = `package {{.Package}}

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"{{.ModulePath}}/internal/domain/entity"
	"{{.ModulePath}}/internal/adapters/db/sqlc"
)

type sqlc{{.GoName}}Repository struct {
	q   *sqlc.Queries
	log *slog.Logger
}

// NewSQLC{{.GoName}}Repository constructs a sqlc-backed {{.GoName}}Repository.
func NewSQLC{{.GoName}}Repository(q *sqlc.Queries, log *slog.Logger) {{.GoName}}Repository {
	return &sqlc{{.GoName}}Repository{q: q, log: log}
}

func (r *sqlc{{.GoName}}Repository) Create(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	start := time.Now()
	_, err := r.q.Create{{.GoName}}(ctx, sqlc.Create{{.GoName}}Params{
{{- range .Columns}}
		{{.GoName}}: {{.VarName}}.{{.GoName}},
{{- end}}
	})
	if err != nil {
		return fmt.Errorf("{{.TableName}} create: %w", err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} created",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (r *sqlc{{.GoName}}Repository) GetByID(ctx context.Context, id {{.IDType}}) (*entity.{{.GoName}}, error) {
	start := time.Now()
	row, err := r.q.Get{{.GoName}}ByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("{{.TableName}} get by id %v: %w", id, err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} fetched",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return mapSQLCTo{{.GoName}}(row), nil
}

func (r *sqlc{{.GoName}}Repository) Update(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	start := time.Now()
	_, err := r.q.Update{{.GoName}}(ctx, sqlc.Update{{.GoName}}Params{
{{- range .Columns}}
		{{.GoName}}: {{.VarName}}.{{.GoName}},
{{- end}}
	})
	if err != nil {
		return fmt.Errorf("{{.TableName}} update: %w", err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} updated",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (r *sqlc{{.GoName}}Repository) Delete(ctx context.Context, id {{.IDType}}) error {
	start := time.Now()
	if err := r.q.Delete{{.GoName}}(ctx, id); err != nil {
		return fmt.Errorf("{{.TableName}} delete id %v: %w", id, err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} deleted",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (r *sqlc{{.GoName}}Repository) List(ctx context.Context, limit, offset int) ([]*entity.{{.GoName}}, error) {
	start := time.Now()
	rows, err := r.q.List{{.GoName}}s(ctx, sqlc.List{{.GoName}}sParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("{{.TableName}} list: %w", err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} listed",
		slog.String("table", "{{.TableName}}"),
		slog.Int("count", len(rows)),
		slog.Duration("duration", time.Since(start)),
	)
	out := make([]*entity.{{.GoName}}, len(rows))
	for i, row := range rows {
		out[i] = mapSQLCTo{{.GoName}}(row)
	}
	return out, nil
}

func mapSQLCTo{{.GoName}}(row sqlc.{{.GoName}}) *entity.{{.GoName}} {
	return &entity.{{.GoName}}{
{{- range .Columns}}
		{{.GoName}}: row.{{.GoName}},
{{- end}}
	}
}
`

type sqlcConfigData struct {
	TableName string
	Engine    string
	Package   string
}

type sqlcQueriesData struct {
	GoName        string
	TableName     string
	InsertCols    []colMeta
	InsertColsLen int
	UpdateCols    []colMeta
	UpdateColsLen int
	WhereArgN     int
}

type sqlcAdapterData struct {
	Package    string
	ModulePath string
	GoName     string
	VarName    string
	TableName  string
	IDType     string
	Columns    []colMeta
}

type colMeta struct {
	Name    string
	GoName  string
	VarName string
}

// SQLCGenerator generates sqlc configuration, queries, and repository adapters.
type SQLCGenerator struct{}

// NewSQLCGenerator returns a new SQLCGenerator.
func NewSQLCGenerator() *SQLCGenerator { return &SQLCGenerator{} }

func sqlcEngine(driver string) string {
	switch driver {
	case "mysql", "mariadb":
		return "mysql"
	case "sqlite":
		return "sqlite"
	default:
		return "postgresql"
	}
}

// GenerateConfig renders the sqlc.yaml for the given table.
func (g *SQLCGenerator) GenerateConfig(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	pkg := opts.Package
	if pkg == "" {
		pkg = "sqlcgen"
	}

	data := sqlcConfigData{
		TableName: table.Name,
		Engine:    sqlcEngine(opts.Driver),
		Package:   pkg,
	}

	out, err := renderTemplate("sqlc_config", sqlcConfigTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render sqlc config: %w", err)
	}

	return entity.GeneratedFile{
		Path:    "sqlc.yaml",
		Content: out,
	}, nil
}

// GenerateQueries renders the queries.sql for the given table.
func (g *SQLCGenerator) GenerateQueries(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	goName := toTitle(table.Name)

	var insertCols, updateCols []colMeta
	for _, c := range table.Columns {
		if c.IsPrimary && c.IsSerial {
			continue
		}
		m := colMeta{Name: c.Name, GoName: c.GoName}
		insertCols = append(insertCols, m)
		if !c.IsPrimary {
			updateCols = append(updateCols, m)
		}
	}

	funcMap := template.FuncMap{
		"inc":  func(i int) int { return i + 1 },
		"last": func(i, n int) bool { return i == n-1 },
	}

	tmpl, err := parseTemplateWithFuncs("sqlc_queries", sqlcQueriesTmpl, funcMap)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("parse sqlc queries template: %w", err)
	}

	data := sqlcQueriesData{
		GoName:        goName,
		TableName:     table.Name,
		InsertCols:    insertCols,
		InsertColsLen: len(insertCols),
		UpdateCols:    updateCols,
		UpdateColsLen: len(updateCols),
		WhereArgN:     len(updateCols) + 1,
	}

	out, err := execTemplate(tmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render sqlc queries: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("db/queries/%s.sql", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}

// GenerateAdapter renders the sqlc adapter repository for the given table.
func (g *SQLCGenerator) GenerateAdapter(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	pkg := opts.Package
	if pkg == "" {
		pkg = "repository"
	}

	module := opts.ModulePath
	if module == "" {
		module = "github.com/example/app"
	}

	goName := toTitle(table.Name)
	varName := strings.ToLower(goName[:1]) + goName[1:]

	var cols []colMeta
	for _, c := range table.Columns {
		cols = append(cols, colMeta{
			Name:    c.Name,
			GoName:  c.GoName,
			VarName: varName,
		})
	}

	data := sqlcAdapterData{
		Package:    pkg,
		ModulePath: module,
		GoName:     goName,
		VarName:    varName,
		TableName:  table.Name,
		IDType:     pkIDType(table),
		Columns:    cols,
	}

	out, err := renderTemplate("sqlc_adapter", sqlcAdapterTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render sqlc adapter: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("internal/adapters/db/repository/sqlc_%s_repository.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}
