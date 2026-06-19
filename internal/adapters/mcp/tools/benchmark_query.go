package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/application/service"
)

// HandleBenchmarkQuery returns the handler for the benchmark_query MCP tool.
func HandleBenchmarkQuery(svc *service.GeneratorService, log *slog.Logger) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query is required"), nil
		}

		analysis, err := svc.AnalyzeQuery(ctx, query)
		if err != nil {
			log.ErrorContext(ctx, "benchmark_query failed", slog.String("error", err.Error()))
			return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %s", err)), nil
		}

		var sb strings.Builder
		sb.WriteString("# SQL Query Analysis & Benchmark\n\n")

		if len(analysis.Issues) > 0 {
			sb.WriteString("## Issues\n\n")
			for _, issue := range analysis.Issues {
				sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", issue.Severity, issue.Code, issue.Description))
			}
			sb.WriteString("\n")
		} else {
			sb.WriteString("## Issues\n\nNo issues detected.\n\n")
		}

		if len(analysis.Recommendations) > 0 {
			sb.WriteString("## Recommendations\n\n")
			for i, rec := range analysis.Recommendations {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
			}
			sb.WriteString("\n")
		}

		if analysis.BenchmarkCode != "" {
			sb.WriteString("## Benchmark\n\n```go\n")
			sb.WriteString(analysis.BenchmarkCode)
			sb.WriteString("\n```\n\n")
		}

		sb.WriteString("## How to Run\n\n```bash\ngo test -bench=BenchmarkQuery -benchmem -count=3 ./...\n```\n")

		return mcp.NewToolResultText(sb.String()), nil
	}
}
