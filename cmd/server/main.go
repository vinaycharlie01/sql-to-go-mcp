package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	mcpserver "github.com/mark3labs/mcp-go/server"
	mcpadapter "github.com/vinaycharlie01/sql-to-go-mcp/internal/adapters/mcp"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/application/service"
	"github.com/vinaycharlie01/sql-to-go-mcp/internal/config"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/analyzer"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/logger"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/sqlparser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath string
		logLevel   string
		logFormat  string
		port       int
	)

	flag.StringVar(&configPath, "config", "configs/config.yaml", "Path to config file")
	flag.StringVar(&logLevel, "log-level", "", "Override log level (debug|info|warn|error)")
	flag.StringVar(&logFormat, "log-format", "", "Override log format (json|text)")
	flag.IntVar(&port, "port", 0, "Override server port")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// CLI overrides (highest priority)
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat != "" {
		cfg.Logging.Format = logFormat
	}
	if port != 0 {
		cfg.Server.Port = port
	}

	log := logger.New(cfg.Logging.Level, cfg.Logging.Format)

	log.Info("starting sql-repository-mcp",
		slog.String("log_level", cfg.Logging.Level),
		slog.String("log_format", cfg.Logging.Format),
		slog.Int("port", cfg.Server.Port),
	)

	parser := sqlparser.New()
	queryAnalyzer := analyzer.New()

	svc := service.NewGeneratorService(parser, queryAnalyzer, log)

	s := mcpadapter.NewServer(svc, log)

	log.Info("MCP server ready — listening on stdio")

	if err := mcpserver.ServeStdio(s); err != nil {
		return fmt.Errorf("serve stdio: %w", err)
	}

	return nil
}
