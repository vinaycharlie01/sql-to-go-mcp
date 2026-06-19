package generator

import (
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

// requiredImports computes the set of imports needed for the generated entity
// based on the column types.
func requiredImports(table *entity.TableDefinition) []string {
	set := map[string]struct{}{}
	for _, c := range table.Columns {
		switch c.GoType {
		case "time.Time":
			set["time"] = struct{}{}
		case "json.RawMessage":
			set["encoding/json"] = struct{}{}
		}
	}
	var out []string
	for k := range set {
		out = append(out, k)
	}
	return out
}

// gormTagForColumn builds the GORM struct tag for a column.
func gormTagForColumn(c entity.Column) string {
	var parts []string
	parts = append(parts, "column:"+c.Name)
	if c.IsPrimary {
		parts = append(parts, "primaryKey")
		if c.IsSerial {
			parts = append(parts, "autoIncrement")
		}
	}
	if c.IsNotNull && !c.IsPrimary {
		parts = append(parts, "not null")
	}
	if c.IsUnique {
		parts = append(parts, "uniqueIndex")
	}
	if c.HasDefault {
		parts = append(parts, "default:"+c.Default)
	}
	return strings.Join(parts, ";")
}

// jsonTagForColumn returns a json struct tag value.
func jsonTagForColumn(c entity.Column) string {
	if c.IsNullable {
		return c.Name + ",omitempty"
	}
	return c.Name
}
