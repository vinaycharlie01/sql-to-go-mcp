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

// HandleGenerateRepository returns the handler for the generate_repository MCP tool.
func HandleGenerateRepository(svc *service.GeneratorService, log *slog.Logger) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tableDef, err := req.RequireString("table_definition")
		if err != nil {
			return mcp.NewToolResultError("table_definition is required"), nil
		}

		opts := entity.GeneratorOptions{
			ModulePath: optString(req, "module_path", "github.com/example/app"),
			Package:    optString(req, "package", "repository"),
			ORM:        optString(req, "orm", "gorm"),
			Driver:     optString(req, "driver", "postgres"),
			WithTests:  true,
			WithBench:  true,
		}

		result, err := svc.GenerateFromSchema(ctx, tableDef, opts)
		if err != nil {
			log.ErrorContext(ctx, "generate_repository failed", slog.String("error", err.Error()))
			return mcp.NewToolResultError(fmt.Sprintf("generation failed: %s", err)), nil
		}

		return mcp.NewToolResultText(formatGenerationResult(result)), nil
	}
}

func formatGenerationResult(r *entity.GenerationResult) string {
	var sb strings.Builder

	sb.WriteString("# Generated Repository\n\n")

	writeFile(&sb, r.Entity)
	writeFile(&sb, r.Interface)
	writeFile(&sb, r.Implementation)
	writeFile(&sb, r.Migration)

	if r.SQLCConfig.Content != "" {
		writeFile(&sb, r.SQLCConfig)
	}
	if r.SQLCQuery.Content != "" {
		writeFile(&sb, r.SQLCQuery)
	}

	for _, f := range r.Tests {
		writeFile(&sb, f)
	}
	for _, f := range r.Benchmarks {
		writeFile(&sb, f)
	}

	return sb.String()
}

func writeFile(sb *strings.Builder, f entity.GeneratedFile) {
	if f.Content == "" {
		return
	}
	sb.WriteString(fmt.Sprintf("## %s\n\n```go\n%s\n```\n\n", f.Path, f.Content))
}

func optString(req mcp.CallToolRequest, key, defaultVal string) string {
	v, err := req.RequireString(key)
	if err != nil || strings.TrimSpace(v) == "" {
		return defaultVal
	}
	return v
}
