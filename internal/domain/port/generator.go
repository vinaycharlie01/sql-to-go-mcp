package port

import (
	"context"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

// SQLParser parses SQL DDL/DML into domain entities.
type SQLParser interface {
	ParseCreateTable(sql string) (*entity.TableDefinition, error)
}

// RepositoryGenerator generates repository layer artifacts from a table definition.
type RepositoryGenerator interface {
	GenerateEntity(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateInterface(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateGORMImpl(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateSQLCImpl(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateRawImpl(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
}

// TestGenerator generates test files for a repository.
type TestGenerator interface {
	GenerateUnitTests(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateIntegrationTests(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateMocks(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
}

// BenchmarkGenerator generates benchmark files.
type BenchmarkGenerator interface {
	GenerateBenchmark(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
}

// SQLCGenerator generates sqlc configuration and queries.
type SQLCGenerator interface {
	GenerateConfig(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
	GenerateQueries(ctx context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error)
}

// QueryAnalyzer analyzes SQL queries and returns optimization recommendations.
type QueryAnalyzer interface {
	Analyze(ctx context.Context, query string) (*QueryAnalysis, error)
}

// QueryAnalysis holds the result of a SQL query analysis.
type QueryAnalysis struct {
	Issues          []Issue
	Recommendations []string
	BenchmarkCode   string
}

// Issue represents a detected problem in a SQL query.
type Issue struct {
	Severity    string // error | warning | info
	Code        string
	Description string
}
