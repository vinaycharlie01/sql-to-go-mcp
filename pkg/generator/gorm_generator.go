package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

const gormImplTmpl = `package {{.Package}}

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"{{.ModulePath}}/internal/domain/entity"
)

type gorm{{.GoName}}Repository struct {
	db  *gorm.DB
	log *slog.Logger
}

// New{{.GoName}}Repository constructs a GORM-backed {{.GoName}}Repository.
func New{{.GoName}}Repository(db *gorm.DB, log *slog.Logger) {{.GoName}}Repository {
	return &gorm{{.GoName}}Repository{db: db, log: log}
}

func (r *gorm{{.GoName}}Repository) Create(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	start := time.Now()
	if err := r.db.WithContext(ctx).Create({{.VarName}}).Error; err != nil {
		return fmt.Errorf("{{.TableName}} create: %w", err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} created",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (r *gorm{{.GoName}}Repository) GetByID(ctx context.Context, id {{.IDType}}) (*entity.{{.GoName}}, error) {
	start := time.Now()
	var {{.VarName}} entity.{{.GoName}}
	if err := r.db.WithContext(ctx).First(&{{.VarName}}, id).Error; err != nil {
		return nil, fmt.Errorf("{{.TableName}} get by id %v: %w", id, err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} fetched",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return &{{.VarName}}, nil
}

func (r *gorm{{.GoName}}Repository) Update(ctx context.Context, {{.VarName}} *entity.{{.GoName}}) error {
	start := time.Now()
	if err := r.db.WithContext(ctx).Save({{.VarName}}).Error; err != nil {
		return fmt.Errorf("{{.TableName}} update: %w", err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} updated",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (r *gorm{{.GoName}}Repository) Delete(ctx context.Context, id {{.IDType}}) error {
	start := time.Now()
	if err := r.db.WithContext(ctx).Delete(&entity.{{.GoName}}{}, id).Error; err != nil {
		return fmt.Errorf("{{.TableName}} delete id %v: %w", id, err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} deleted",
		slog.String("table", "{{.TableName}}"),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (r *gorm{{.GoName}}Repository) List(ctx context.Context, limit, offset int) ([]*entity.{{.GoName}}, error) {
	start := time.Now()
	var records []*entity.{{.GoName}}
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("{{.TableName}} list: %w", err)
	}
	r.log.InfoContext(ctx, "{{.TableName}} listed",
		slog.String("table", "{{.TableName}}"),
		slog.Int("count", len(records)),
		slog.Duration("duration", time.Since(start)),
	)
	return records, nil
}
`

type gormImplData struct {
	Package    string
	ModulePath string
	GoName     string
	VarName    string
	TableName  string
	IDType     string
}

// GORMGenerator generates GORM repository implementations.
type GORMGenerator struct{}

// NewGORMGenerator returns a new GORMGenerator.
func NewGORMGenerator() *GORMGenerator { return &GORMGenerator{} }

// Generate renders the GORM implementation for the given table.
func (g *GORMGenerator) Generate(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
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

	data := gormImplData{
		Package:    pkg,
		ModulePath: module,
		GoName:     goName,
		VarName:    varName,
		TableName:  table.Name,
		IDType:     pkIDType(table),
	}

	out, err := renderTemplate("gorm_impl", gormImplTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render gorm impl template: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("internal/adapters/db/repository/gorm_%s_repository.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}
