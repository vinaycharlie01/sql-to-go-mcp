package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/application/service"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/analyzer"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/sqlparser"
)

var testSchema = `CREATE TABLE orders (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL,
	total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
)`

func newService() *service.GeneratorService {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return service.NewGeneratorService(sqlparser.New(), analyzer.New(), log)
}

func TestGeneratorService_GenerateFromSchema_GORM(t *testing.T) {
	svc := newService()

	opts := entity.GeneratorOptions{
		ModulePath: "github.com/example/app",
		Package:    "repository",
		ORM:        "gorm",
		Driver:     "postgres",
		WithTests:  true,
		WithBench:  true,
	}

	result, err := svc.GenerateFromSchema(context.Background(), testSchema, opts)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Entity.Content)
	assert.NotEmpty(t, result.Interface.Content)
	assert.NotEmpty(t, result.Implementation.Content)
	assert.NotEmpty(t, result.Migration.Content)
	assert.NotEmpty(t, result.Tests, "should have test files")
	assert.NotEmpty(t, result.Benchmarks, "should have benchmark files")

	assert.Contains(t, result.Entity.Content, "type Order struct")
	assert.Contains(t, result.Interface.Content, "OrderRepository interface")
	assert.Contains(t, result.Implementation.Content, "gorm.DB")
}

func TestGeneratorService_GenerateFromSchema_SQLC(t *testing.T) {
	svc := newService()

	opts := entity.GeneratorOptions{
		ModulePath: "github.com/example/app",
		Package:    "sqlcgen",
		ORM:        "sqlc",
		Driver:     "postgres",
		WithTests:  true,
		WithBench:  false,
	}

	result, err := svc.GenerateFromSchema(context.Background(), testSchema, opts)
	require.NoError(t, err)

	assert.NotEmpty(t, result.SQLCConfig.Content)
	assert.NotEmpty(t, result.SQLCQuery.Content)
	assert.NotEmpty(t, result.Implementation.Content)
}

func TestGeneratorService_GenerateFromSchema_InvalidSchema(t *testing.T) {
	svc := newService()

	_, err := svc.GenerateFromSchema(context.Background(), "not a create table", entity.GeneratorOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse schema")
}

func TestGeneratorService_AnalyzeQuery(t *testing.T) {
	svc := newService()

	tests := []struct {
		name        string
		query       string
		wantIssues  bool
		wantErr     bool
	}{
		{
			name:       "clean query",
			query:      "SELECT id, total FROM orders WHERE id = $1 LIMIT 1",
			wantIssues: false,
		},
		{
			name:       "select star",
			query:      "SELECT * FROM orders LIMIT 100",
			wantIssues: true,
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := svc.AnalyzeQuery(context.Background(), tc.query)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.wantIssues {
				assert.NotEmpty(t, result.Issues)
			} else {
				assert.Empty(t, result.Issues)
			}
			assert.NotEmpty(t, result.BenchmarkCode)
		})
	}
}
