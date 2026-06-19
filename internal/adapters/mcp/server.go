package mcp

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/adapters/mcp/tools"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/application/service"
)

const (
	serverName    = "sql-repository-mcp"
	serverVersion = "1.0.0"
)

// NewServer builds and returns a configured MCP server.
func NewServer(svc *service.GeneratorService, log *slog.Logger) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer(
		serverName,
		serverVersion,
		mcpserver.WithToolCapabilities(true),
	)

	registerGenerateRepository(s, svc, log)
	registerSQLToGORM(s, svc, log)
	registerSQLToSQLC(s, svc, log)
	registerBenchmarkQuery(s, svc, log)
	registerGenerateTests(s, svc, log)

	return s
}

// registerGenerateRepository registers the generate_repository tool.
func registerGenerateRepository(s *mcpserver.MCPServer, svc *service.GeneratorService, log *slog.Logger) {
	tool := mcp.NewTool("generate_repository",
		mcp.WithDescription(
			"Generate a complete repository layer (entity, interface, GORM/sqlc implementation, "+
				"migration, unit tests, integration tests, mocks, benchmarks) from a SQL CREATE TABLE statement.",
		),
		mcp.WithString("table_definition",
			mcp.Required(),
			mcp.Description("SQL CREATE TABLE statement to generate code from."),
		),
		mcp.WithString("module_path",
			mcp.Description("Go module path (e.g. github.com/example/app). Defaults to github.com/example/app."),
		),
		mcp.WithString("orm",
			mcp.Description("ORM to use: gorm (default) or sqlc."),
		),
		mcp.WithString("driver",
			mcp.Description("Database driver: postgres (default), mysql, sqlite, sqlserver."),
		),
		mcp.WithString("package",
			mcp.Description("Go package name for generated files. Defaults to repository."),
		),
	)
	s.AddTool(tool, tools.HandleGenerateRepository(svc, log))
}

// registerSQLToGORM registers the sql_to_gorm tool.
func registerSQLToGORM(s *mcpserver.MCPServer, svc *service.GeneratorService, log *slog.Logger) {
	tool := mcp.NewTool("sql_to_gorm",
		mcp.WithDescription(
			"Convert a SQL CREATE TABLE schema into a GORM model, repository implementation, and tests.",
		),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("SQL CREATE TABLE statement."),
		),
		mcp.WithString("module_path",
			mcp.Description("Go module path."),
		),
		mcp.WithString("package",
			mcp.Description("Go package name for generated files."),
		),
	)
	s.AddTool(tool, tools.HandleSQLToGORM(svc, log))
}

// registerSQLToSQLC registers the sql_to_sqlc tool.
func registerSQLToSQLC(s *mcpserver.MCPServer, svc *service.GeneratorService, log *slog.Logger) {
	tool := mcp.NewTool("sql_to_sqlc",
		mcp.WithDescription(
			"Convert a SQL CREATE TABLE schema and optional query into a sqlc configuration, queries file, and repository adapter.",
		),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("SQL CREATE TABLE statement."),
		),
		mcp.WithString("query",
			mcp.Description("Optional SQL query to include in the sqlc queries file."),
		),
		mcp.WithString("driver",
			mcp.Description("Database driver: postgres (default), mysql, sqlite."),
		),
		mcp.WithString("module_path",
			mcp.Description("Go module path."),
		),
	)
	s.AddTool(tool, tools.HandleSQLToSQLC(svc, log))
}

// registerBenchmarkQuery registers the benchmark_query tool.
func registerBenchmarkQuery(s *mcpserver.MCPServer, svc *service.GeneratorService, log *slog.Logger) {
	tool := mcp.NewTool("benchmark_query",
		mcp.WithDescription(
			"Analyze a SQL query for performance issues and generate a Go benchmark test for it.",
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("SQL query to analyze and benchmark."),
		),
	)
	s.AddTool(tool, tools.HandleBenchmarkQuery(svc, log))
}

// registerGenerateTests registers the generate_tests tool.
func registerGenerateTests(s *mcpserver.MCPServer, svc *service.GeneratorService, log *slog.Logger) {
	tool := mcp.NewTool("generate_tests",
		mcp.WithDescription(
			"Generate unit tests, integration tests, and testify mocks for a repository derived from a SQL CREATE TABLE statement.",
		),
		mcp.WithString("table_definition",
			mcp.Required(),
			mcp.Description("SQL CREATE TABLE statement."),
		),
		mcp.WithString("module_path",
			mcp.Description("Go module path."),
		),
		mcp.WithString("package",
			mcp.Description("Go package name."),
		),
	)
	s.AddTool(tool, tools.HandleGenerateTests(svc, log))
}
