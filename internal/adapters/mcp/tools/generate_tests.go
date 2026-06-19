package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/application/service"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

// HandleGenerateTests returns the handler for the generate_tests MCP tool.
func HandleGenerateTests(svc *service.GeneratorService, log *slog.Logger) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tableDef, err := req.RequireString("table_definition")
		if err != nil {
			return mcp.NewToolResultError("table_definition is required"), nil
		}

		opts := entity.GeneratorOptions{
			ModulePath: optString(req, "module_path", "github.com/example/app"),
			Package:    optString(req, "package", "repository"),
			ORM:        "gorm",
			Driver:     "postgres",
			WithTests:  true,
			WithBench:  true,
		}

		result, err := svc.GenerateFromSchema(ctx, tableDef, opts)
		if err != nil {
			log.ErrorContext(ctx, "generate_tests failed", slog.String("error", err.Error()))
			return mcp.NewToolResultError(fmt.Sprintf("generation failed: %s", err)), nil
		}

		var sb strings.Builder
		sb.WriteString("# Generated Tests\n\n")

		for _, f := range result.Tests {
			writeFile(&sb, f)
		}
		for _, f := range result.Benchmarks {
			writeFile(&sb, f)
		}

		sb.WriteString("## How to Run\n\n")
		sb.WriteString("```bash\n")
		sb.WriteString("# Unit tests\n")
		sb.WriteString("go test ./internal/domain/repository/...\n\n")
		sb.WriteString("# Integration tests (requires a running database)\n")
		sb.WriteString("go test -tags=integration ./internal/adapters/db/repository/...\n\n")
		sb.WriteString("# Benchmarks\n")
		sb.WriteString("go test -bench=. -benchmem -count=3 ./pkg/benchmark/...\n")
		sb.WriteString("```\n")

		return mcp.NewToolResultText(sb.String()), nil
	}
}
