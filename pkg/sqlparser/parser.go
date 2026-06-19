package sqlparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/entity"
)

var (
	reCreateTable = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?[` + "`" + `"']?(\w+)[` + "`" + `"']?\s*\((.+)\)`)
	reColumn      = regexp.MustCompile(`(?i)^\s*[` + "`" + `"']?(\w+)[` + "`" + `"']?\s+([A-Z]+(?:\s*\([^)]*\))?)(.*)$`)
	reVarType     = regexp.MustCompile(`(?i)^(\w+)\s*(?:\((\d+)(?:\s*,\s*(\d+))?\))?`)
	rePrimaryKey  = regexp.MustCompile(`(?is)^\s*PRIMARY\s+KEY\s*\(([^)]+)\)`)
	reUniqueKey   = regexp.MustCompile(`(?is)^\s*UNIQUE\s+(?:KEY\s+\w+\s*)?\(([^)]+)\)`)
	reForeignKey  = regexp.MustCompile(`(?is)^\s*(?:CONSTRAINT\s+\w+\s+)?FOREIGN\s+KEY\s*\(([^)]+)\)\s*REFERENCES\s+(\w+)\s*\(([^)]+)\)(?:\s+ON\s+DELETE\s+(\w+(?:\s+\w+)?))?(?:\s+ON\s+UPDATE\s+(\w+(?:\s+\w+)?))?`)
)

// Parser parses SQL CREATE TABLE statements.
type Parser struct{}

// New returns a new Parser.
func New() *Parser { return &Parser{} }

// ParseCreateTable parses a single CREATE TABLE statement.
func (p *Parser) ParseCreateTable(sql string) (*entity.TableDefinition, error) {
	m := reCreateTable.FindStringSubmatch(sql)
	if m == nil {
		return nil, fmt.Errorf("no CREATE TABLE statement found")
	}

	table := &entity.TableDefinition{Name: m[1]}
	body := m[2]

	for _, rawLine := range splitColumnDefs(body) {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		upper := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(upper, "PRIMARY KEY"):
			if pk := rePrimaryKey.FindStringSubmatch(line); pk != nil {
				for _, col := range splitIdents(pk[1]) {
					table.PrimaryKeys = append(table.PrimaryKeys, col)
				}
			}

		case strings.HasPrefix(upper, "UNIQUE"):
			if uk := reUniqueKey.FindStringSubmatch(line); uk != nil {
				cols := splitIdents(uk[1])
				table.Indexes = append(table.Indexes, entity.Index{
					Columns: cols,
					Unique:  true,
				})
			}

		case strings.HasPrefix(upper, "FOREIGN KEY"),
			strings.HasPrefix(upper, "CONSTRAINT"):
			if fk := reForeignKey.FindStringSubmatch(line); fk != nil {
				table.ForeignKeys = append(table.ForeignKeys, entity.ForeignKey{
					Column:    strings.TrimSpace(fk[1]),
					RefTable:  fk[2],
					RefColumn: strings.TrimSpace(fk[3]),
					OnDelete:  fk[4],
					OnUpdate:  fk[5],
				})
			}

		case strings.HasPrefix(upper, "CHECK"),
			strings.HasPrefix(upper, "KEY"),
			strings.HasPrefix(upper, "INDEX"):
			// skip unsupported constraint types

		default:
			col, err := parseColumn(line)
			if err == nil {
				table.Columns = append(table.Columns, *col)
			}
		}
	}

	// Apply table-level primary keys
	for i := range table.Columns {
		for _, pk := range table.PrimaryKeys {
			if strings.EqualFold(table.Columns[i].Name, pk) {
				table.Columns[i].IsPrimary = true
			}
		}
	}

	if len(table.Columns) == 0 {
		return nil, fmt.Errorf("no columns parsed from CREATE TABLE")
	}

	return table, nil
}

func parseColumn(line string) (*entity.Column, error) {
	m := reColumn.FindStringSubmatch(line)
	if m == nil {
		return nil, fmt.Errorf("cannot parse column: %q", line)
	}

	name := m[1]
	rawType := strings.TrimSpace(m[2])
	constraints := strings.ToUpper(m[3])

	tm := reVarType.FindStringSubmatch(rawType)
	if tm == nil {
		return nil, fmt.Errorf("cannot parse type: %q", rawType)
	}

	baseType := strings.ToUpper(tm[1])
	length, _ := strconv.Atoi(tm[2])
	scale, _ := strconv.Atoi(tm[3])

	col := &entity.Column{
		Name:      name,
		GoName:    toPascalCase(name),
		SQLType:   rawType,
		GoType:    sqlTypeToGo(baseType, length, scale),
		MaxLength: length,
		Scale:     scale,
		IsSerial:  isSerial(baseType),
	}

	col.IsNotNull = strings.Contains(constraints, "NOT NULL")
	col.IsUnique = strings.Contains(constraints, "UNIQUE")
	col.IsPrimary = strings.Contains(constraints, "PRIMARY KEY")

	if idx := strings.Index(constraints, "DEFAULT"); idx >= 0 {
		col.HasDefault = true
		after := strings.TrimSpace(m[3][idx+7:])
		// grab until end or next keyword
		col.Default = extractDefault(after)
	}

	col.IsNullable = !col.IsNotNull && !col.IsPrimary

	return col, nil
}

func sqlTypeToGo(baseType string, length, scale int) string {
	switch baseType {
	case "BIGSERIAL", "BIGINT", "INT8":
		return "int64"
	case "SERIAL", "INTEGER", "INT", "INT4", "MEDIUMINT":
		return "int32"
	case "SMALLSERIAL", "SMALLINT", "INT2", "TINYINT":
		return "int16"
	case "BOOLEAN", "BOOL", "BIT":
		return "bool"
	case "REAL", "FLOAT4":
		return "float32"
	case "DOUBLE", "FLOAT8", "FLOAT":
		return "float64"
	case "DECIMAL", "NUMERIC", "MONEY":
		return "float64"
	case "TEXT", "TINYTEXT", "MEDIUMTEXT", "LONGTEXT",
		"VARCHAR", "CHARACTER", "CHAR", "NVARCHAR", "NCHAR":
		return "string"
	case "UUID":
		return "string"
	case "TIMESTAMP", "TIMESTAMPTZ", "DATETIME", "DATE", "TIME", "TIMETZ":
		return "time.Time"
	case "JSONB", "JSON":
		return "json.RawMessage"
	case "BYTEA", "BLOB", "VARBINARY", "BINARY":
		return "[]byte"
	default:
		return "string"
	}
}

func isSerial(t string) bool {
	switch t {
	case "SERIAL", "BIGSERIAL", "SMALLSERIAL":
		return true
	}
	return false
}

func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var b strings.Builder
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		b.WriteRune(unicode.ToUpper(rune(w[0])))
		b.WriteString(w[1:])
	}
	return b.String()
}

// splitColumnDefs splits the body of a CREATE TABLE into individual column/constraint lines,
// respecting nested parentheses.
func splitColumnDefs(body string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range body {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}
	if start < len(body) {
		parts = append(parts, body[start:])
	}
	return parts
}

func splitIdents(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.Trim(strings.TrimSpace(p), "`\"'")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func extractDefault(s string) string {
	keywords := []string{"NOT NULL", "NULL", "UNIQUE", "PRIMARY", "REFERENCES", "CHECK", "COMMENT"}
	upper := strings.ToUpper(s)
	end := len(s)
	for _, kw := range keywords {
		if idx := strings.Index(upper, kw); idx >= 0 && idx < end {
			end = idx
		}
	}
	return strings.TrimSpace(s[:end])
}
