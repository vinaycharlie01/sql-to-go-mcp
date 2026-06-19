package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/generator"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/sqlparser"
)

var usersSQL = `CREATE TABLE users (
	id BIGSERIAL PRIMARY KEY,
	email VARCHAR(255) NOT NULL UNIQUE,
	name VARCHAR(100) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMP
)`

var defaultOpts = entity.GeneratorOptions{
	ModulePath: "github.com/example/app",
	Package:    "repository",
	ORM:        "gorm",
	Driver:     "postgres",
	WithTests:  true,
	WithBench:  true,
}

func parseTable(t *testing.T) *entity.TableDefinition {
	t.Helper()
	p := sqlparser.New()
	table, err := p.ParseCreateTable(usersSQL)
	require.NoError(t, err)
	return table
}

func TestEntityGenerator(t *testing.T) {
	g := generator.NewEntityGenerator()
	table := parseTable(t)

	file, err := g.Generate(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.NotEmpty(t, file.Path)
	assert.Contains(t, file.Path, "users")
	assert.Contains(t, file.Content, "type User struct")
	assert.Contains(t, file.Content, "TableName()")
	assert.Contains(t, file.Content, "gorm:")
	assert.Contains(t, file.Content, "json:")
}

func TestInterfaceGenerator(t *testing.T) {
	g := generator.NewInterfaceGenerator()
	table := parseTable(t)

	file, err := g.Generate(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "UserRepository interface")
	assert.Contains(t, file.Content, "Create(ctx context.Context")
	assert.Contains(t, file.Content, "GetByID(ctx context.Context")
	assert.Contains(t, file.Content, "Update(ctx context.Context")
	assert.Contains(t, file.Content, "Delete(ctx context.Context")
	assert.Contains(t, file.Content, "List(ctx context.Context")
}

func TestGORMGenerator(t *testing.T) {
	g := generator.NewGORMGenerator()
	table := parseTable(t)

	file, err := g.Generate(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "gorm.DB")
	assert.Contains(t, file.Content, "func (r *gormUserRepository) Create(")
	assert.Contains(t, file.Content, "func (r *gormUserRepository) GetByID(")
	assert.Contains(t, file.Content, "func (r *gormUserRepository) Update(")
	assert.Contains(t, file.Content, "func (r *gormUserRepository) Delete(")
	assert.Contains(t, file.Content, "func (r *gormUserRepository) List(")
	assert.Contains(t, file.Content, "r.log.InfoContext")
}

func TestSQLCGenerator_Config(t *testing.T) {
	g := generator.NewSQLCGenerator()
	table := parseTable(t)

	opts := defaultOpts
	opts.ORM = "sqlc"
	opts.Driver = "postgres"

	file, err := g.GenerateConfig(context.Background(), table, opts)
	require.NoError(t, err)

	assert.Equal(t, "sqlc.yaml", file.Path)
	assert.Contains(t, file.Content, "postgresql")
	assert.Contains(t, file.Content, "users")
}

func TestSQLCGenerator_Queries(t *testing.T) {
	g := generator.NewSQLCGenerator()
	table := parseTable(t)

	opts := defaultOpts
	opts.ORM = "sqlc"

	file, err := g.GenerateQueries(context.Background(), table, opts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "CreateUser")
	assert.Contains(t, file.Content, "GetUserByID")
	assert.Contains(t, file.Content, "ListUsers")
	assert.Contains(t, file.Content, "UpdateUser")
	assert.Contains(t, file.Content, "DeleteUser")
	assert.True(t, strings.HasSuffix(file.Path, ".sql"))
}

func TestTestGenerator_UnitTests(t *testing.T) {
	g := generator.NewTestGenerator()
	table := parseTable(t)

	file, err := g.GenerateUnitTests(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "MockUserRepository")
	assert.Contains(t, file.Content, "TestUserRepository_Create")
	assert.Contains(t, file.Content, "TestUserRepository_GetByID")
	assert.Contains(t, file.Content, "t.Run(tc.name")
}

func TestTestGenerator_IntegrationTests(t *testing.T) {
	g := generator.NewTestGenerator()
	table := parseTable(t)

	file, err := g.GenerateIntegrationTests(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "TestIntegration_UserRepository_CRUD")
	assert.Contains(t, file.Content, "testing.Short()")
}

func TestTestGenerator_Mocks(t *testing.T) {
	g := generator.NewTestGenerator()
	table := parseTable(t)

	file, err := g.GenerateMocks(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "package mocks")
	assert.Contains(t, file.Content, "mock.Mock")
}

func TestBenchmarkGenerator(t *testing.T) {
	g := generator.NewBenchmarkGenerator()
	table := parseTable(t)

	file, err := g.Generate(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "BenchmarkCreateUser")
	assert.Contains(t, file.Content, "BenchmarkGetByIDUser")
	assert.Contains(t, file.Content, "BenchmarkUpdateUser")
	assert.Contains(t, file.Content, "BenchmarkDeleteUser")
	assert.Contains(t, file.Content, "BenchmarkListUser")
	assert.Contains(t, file.Content, "b.ReportAllocs()")
	assert.Contains(t, file.Content, "b.ResetTimer()")
}

func TestMigrationGenerator(t *testing.T) {
	g := generator.NewMigrationGenerator()
	table := parseTable(t)

	file, err := g.Generate(context.Background(), table, defaultOpts)
	require.NoError(t, err)

	assert.Contains(t, file.Content, "CREATE TABLE IF NOT EXISTS users")
	assert.Contains(t, file.Content, "DROP TABLE IF EXISTS users")
	assert.Contains(t, file.Content, "+migrate Up")
	assert.Contains(t, file.Content, "+migrate Down")
}
