package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

const benchmarkTmpl = `package {{.Package}}_test

import (
	"context"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"{{.ModulePath}}/internal/adapters/db/repository"
	"{{.ModulePath}}/internal/domain/entity"
)

func openBenchDB(b *testing.B) *gorm.DB {
	b.Helper()
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=benchdb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Fatalf("open bench database: %v", err)
	}
	if err := db.AutoMigrate(&entity.{{.GoName}}{}); err != nil {
		b.Fatalf("auto migrate: %v", err)
	}
	return db
}

func BenchmarkCreate{{.GoName}}(b *testing.B) {
	db := openBenchDB(b)
	repo := repository.New{{.GoName}}Repository(db, nil)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := &entity.{{.GoName}}{}
		if err := repo.Create(ctx, rec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetByID{{.GoName}}(b *testing.B) {
	db := openBenchDB(b)
	repo := repository.New{{.GoName}}Repository(db, nil)
	ctx := context.Background()

	seed := &entity.{{.GoName}}{}
	if err := repo.Create(ctx, seed); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByID(ctx, {{.ZeroID}})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdate{{.GoName}}(b *testing.B) {
	db := openBenchDB(b)
	repo := repository.New{{.GoName}}Repository(db, nil)
	ctx := context.Background()

	seed := &entity.{{.GoName}}{}
	if err := repo.Create(ctx, seed); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := repo.Update(ctx, seed); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelete{{.GoName}}(b *testing.B) {
	db := openBenchDB(b)
	repo := repository.New{{.GoName}}Repository(db, nil)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		seed := &entity.{{.GoName}}{}
		if err := repo.Create(ctx, seed); err != nil {
			b.Fatalf("seed: %v", err)
		}
		b.StartTimer()
		if err := repo.Delete(ctx, {{.ZeroID}}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkList{{.GoName}}(b *testing.B) {
	db := openBenchDB(b)
	repo := repository.New{{.GoName}}Repository(db, nil)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, 100, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
`

type benchmarkTmplData struct {
	Package    string
	ModulePath string
	GoName     string
	TableName  string
	ZeroID     string
}

// BenchmarkGenerator generates benchmark files for a repository.
type BenchmarkGenerator struct{}

// NewBenchmarkGenerator returns a new BenchmarkGenerator.
func NewBenchmarkGenerator() *BenchmarkGenerator { return &BenchmarkGenerator{} }

// Generate renders the benchmark file for the given table.
func (g *BenchmarkGenerator) Generate(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	pkg := opts.Package
	if pkg == "" {
		pkg = "repository"
	}

	module := opts.ModulePath
	if module == "" {
		module = "github.com/example/app"
	}

	idType := pkIDType(table)
	zeroID := "0"
	if idType == "string" {
		zeroID = `""`
	}

	data := benchmarkTmplData{
		Package:    pkg,
		ModulePath: module,
		GoName:     toTitle(table.Name),
		TableName:  table.Name,
		ZeroID:     zeroID,
	}

	out, err := renderTemplate("benchmark", benchmarkTmpl, data)
	if err != nil {
		return entity.GeneratedFile{}, fmt.Errorf("render benchmark template: %w", err)
	}

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("pkg/benchmark/%s_bench_test.go", strings.ToLower(table.Name)),
		Content: out,
	}, nil
}
