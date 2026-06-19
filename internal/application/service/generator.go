package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/port"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/generator"
)

// GeneratorService orchestrates SQL schema parsing and code generation.
type GeneratorService struct {
	parser    port.SQLParser
	entity_   *generator.EntityGenerator
	iface     *generator.InterfaceGenerator
	gorm      *generator.GORMGenerator
	sqlc      *generator.SQLCGenerator
	tests     *generator.TestGenerator
	benchmark *generator.BenchmarkGenerator
	migration *generator.MigrationGenerator
	analyzer  port.QueryAnalyzer
	log       *slog.Logger
}

// NewGeneratorService constructs a GeneratorService with all dependencies injected.
func NewGeneratorService(
	parser port.SQLParser,
	analyzer port.QueryAnalyzer,
	log *slog.Logger,
) *GeneratorService {
	return &GeneratorService{
		parser:    parser,
		entity_:   generator.NewEntityGenerator(),
		iface:     generator.NewInterfaceGenerator(),
		gorm:      generator.NewGORMGenerator(),
		sqlc:      generator.NewSQLCGenerator(),
		tests:     generator.NewTestGenerator(),
		benchmark: generator.NewBenchmarkGenerator(),
		migration: generator.NewMigrationGenerator(),
		analyzer:  analyzer,
		log:       log,
	}
}

// GenerateFromSchema parses a CREATE TABLE statement and generates all artifacts.
func (s *GeneratorService) GenerateFromSchema(ctx context.Context, schema string, opts entity.GeneratorOptions) (*entity.GenerationResult, error) {
	start := time.Now()

	table, err := s.parser.ParseCreateTable(schema)
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	result := &entity.GenerationResult{}

	result.Entity, err = s.entity_.Generate(ctx, table, opts)
	if err != nil {
		return nil, fmt.Errorf("generate entity: %w", err)
	}

	result.Interface, err = s.iface.Generate(ctx, table, opts)
	if err != nil {
		return nil, fmt.Errorf("generate interface: %w", err)
	}

	switch opts.ORM {
	case "sqlc":
		result.Implementation, err = s.sqlc.GenerateAdapter(ctx, table, opts)
	default: // gorm
		result.Implementation, err = s.gorm.Generate(ctx, table, opts)
	}
	if err != nil {
		return nil, fmt.Errorf("generate implementation: %w", err)
	}

	result.Migration, err = s.migration.Generate(ctx, table, opts)
	if err != nil {
		return nil, fmt.Errorf("generate migration: %w", err)
	}

	if opts.ORM == "sqlc" {
		sqlcCfg, errCfg := s.sqlc.GenerateConfig(ctx, table, opts)
		if errCfg != nil {
			return nil, fmt.Errorf("generate sqlc config: %w", errCfg)
		}
		result.SQLCConfig = sqlcCfg

		sqlcQ, errQ := s.sqlc.GenerateQueries(ctx, table, opts)
		if errQ != nil {
			return nil, fmt.Errorf("generate sqlc queries: %w", errQ)
		}
		result.SQLCQuery = sqlcQ
	}

	if opts.WithTests {
		unitTest, errU := s.tests.GenerateUnitTests(ctx, table, opts)
		if errU != nil {
			return nil, fmt.Errorf("generate unit tests: %w", errU)
		}
		result.Tests = append(result.Tests, unitTest)

		intTest, errI := s.tests.GenerateIntegrationTests(ctx, table, opts)
		if errI != nil {
			return nil, fmt.Errorf("generate integration tests: %w", errI)
		}
		result.Tests = append(result.Tests, intTest)

		mocks, errM := s.tests.GenerateMocks(ctx, table, opts)
		if errM != nil {
			return nil, fmt.Errorf("generate mocks: %w", errM)
		}
		result.Tests = append(result.Tests, mocks)
	}

	if opts.WithBench {
		bench, errB := s.benchmark.Generate(ctx, table, opts)
		if errB != nil {
			return nil, fmt.Errorf("generate benchmark: %w", errB)
		}
		result.Benchmarks = append(result.Benchmarks, bench)
	}

	s.log.InfoContext(ctx, "repository generated",
		slog.String("table", table.Name),
		slog.String("orm", opts.ORM),
		slog.Duration("duration", time.Since(start)),
	)

	return result, nil
}

// AnalyzeQuery analyzes a SQL query and returns recommendations.
func (s *GeneratorService) AnalyzeQuery(ctx context.Context, query string) (*port.QueryAnalysis, error) {
	start := time.Now()
	result, err := s.analyzer.Analyze(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("analyze query: %w", err)
	}
	s.log.InfoContext(ctx, "query analyzed",
		slog.Int("issues", len(result.Issues)),
		slog.Duration("duration", time.Since(start)),
	)
	return result, nil
}
