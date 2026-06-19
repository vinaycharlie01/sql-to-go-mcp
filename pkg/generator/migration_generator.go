package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

// MigrationGenerator generates SQL migration files.
type MigrationGenerator struct{}

// NewMigrationGenerator returns a new MigrationGenerator.
func NewMigrationGenerator() *MigrationGenerator { return &MigrationGenerator{} }

// Generate produces an up/down migration SQL file for the given table.
func (g *MigrationGenerator) Generate(_ context.Context, table *entity.TableDefinition, opts entity.GeneratorOptions) (entity.GeneratedFile, error) {
	var sb strings.Builder

	sb.WriteString("-- +migrate Up\n")
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", table.Name))

	for i, col := range table.Columns {
		sb.WriteString("    ")
		sb.WriteString(col.Name)
		sb.WriteString(" ")
		sb.WriteString(col.SQLType)

		if col.IsPrimary {
			sb.WriteString(" PRIMARY KEY")
		}
		if col.IsNotNull && !col.IsPrimary {
			sb.WriteString(" NOT NULL")
		}
		if col.IsUnique {
			sb.WriteString(" UNIQUE")
		}
		if col.HasDefault {
			sb.WriteString(" DEFAULT ")
			sb.WriteString(col.Default)
		}

		if i < len(table.Columns)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(");\n")

	for _, idx := range table.Indexes {
		unique := ""
		if idx.Unique {
			unique = "UNIQUE "
		}
		sb.WriteString(fmt.Sprintf(
			"\nCREATE %sINDEX IF NOT EXISTS idx_%s_%s ON %s (%s);\n",
			unique,
			table.Name,
			strings.Join(idx.Columns, "_"),
			table.Name,
			strings.Join(idx.Columns, ", "),
		))
	}

	sb.WriteString("\n-- +migrate Down\n")
	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", table.Name))

	return entity.GeneratedFile{
		Path:    fmt.Sprintf("db/migrations/001_create_%s.sql", strings.ToLower(table.Name)),
		Content: sb.String(),
	}, nil
}
