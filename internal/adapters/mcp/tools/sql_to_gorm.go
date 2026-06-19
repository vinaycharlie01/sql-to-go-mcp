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

// HandleSQLToGORM returns the handler for the sql_to_gorm MCP tool.
func HandleSQLToGORM(svc *service.GeneratorService, log *slog.Logger) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		schema, err := req.RequireString("schema")
		if err != nil {
			return mcp.NewToolResultError("schema is required"), nil
		}

		opts := entity.GeneratorOptions{
			ModulePath: optString(req, "module_path", "github.com/example/app"),
			Package:    optString(req, "package", "repository"),
			ORM:        "gorm",
			Driver:     "postgres",
			WithTests:  true,
			WithBench:  false,
		}

		result, err := svc.GenerateFromSchema(ctx, schema, opts)
		if err != nil {
			log.ErrorContext(ctx, "sql_to_gorm failed", slog.String("error", err.Error()))
			return mcp.NewToolResultError(fmt.Sprintf("generation failed: %s", err)), nil
		}

		var sb strings.Builder
		sb.WriteString("# GORM Model & Repository\n\n")
		writeFile(&sb, result.Entity)
		writeFile(&sb, result.Interface)
		writeFile(&sb, result.Implementation)
		for _, f := range result.Tests {
			writeFile(&sb, f)
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}
