package sqlparser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/sqlparser"
)

func TestParser_ParseCreateTable(t *testing.T) {
	p := sqlparser.New()

	tests := []struct {
		name       string
		sql        string
		wantTable  string
		wantCols   int
		wantErr    bool
		checkFirst func(t *testing.T, got interface{})
	}{
		{
			name: "postgres users table",
			sql: `CREATE TABLE users (
				id BIGSERIAL PRIMARY KEY,
				email VARCHAR(255) NOT NULL UNIQUE,
				name VARCHAR(100) NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT NOW(),
				deleted_at TIMESTAMP
			)`,
			wantTable: "users",
			wantCols:  5,
			wantErr:   false,
		},
		{
			name: "simple table",
			sql: `CREATE TABLE IF NOT EXISTS products (
				id SERIAL PRIMARY KEY,
				title TEXT NOT NULL,
				price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
				in_stock BOOLEAN NOT NULL DEFAULT TRUE
			)`,
			wantTable: "products",
			wantCols:  4,
			wantErr:   false,
		},
		{
			name: "table with composite primary key constraint",
			sql: `CREATE TABLE order_items (
				order_id BIGINT NOT NULL,
				item_id BIGINT NOT NULL,
				quantity INTEGER NOT NULL DEFAULT 1,
				PRIMARY KEY (order_id, item_id)
			)`,
			wantTable: "order_items",
			wantCols:  3,
			wantErr:   false,
		},
		{
			name:    "invalid input",
			sql:     "SELECT 1",
			wantErr: true,
		},
		{
			name:    "empty input",
			sql:     "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := p.ParseCreateTable(tc.sql)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantTable, got.Name)
			assert.Len(t, got.Columns, tc.wantCols)
		})
	}
}

func TestParser_ColumnTypes(t *testing.T) {
	p := sqlparser.New()

	sql := `CREATE TABLE type_test (
		col_bigserial BIGSERIAL PRIMARY KEY,
		col_int INTEGER NOT NULL,
		col_varchar VARCHAR(100) NOT NULL,
		col_text TEXT,
		col_bool BOOLEAN NOT NULL DEFAULT FALSE,
		col_float FLOAT NOT NULL,
		col_decimal DECIMAL(10,2),
		col_timestamp TIMESTAMP NOT NULL,
		col_uuid UUID,
		col_jsonb JSONB
	)`

	got, err := p.ParseCreateTable(sql)
	require.NoError(t, err)
	require.Len(t, got.Columns, 10)

	typeMap := map[string]string{}
	for _, c := range got.Columns {
		typeMap[c.Name] = c.GoType
	}

	assert.Equal(t, "int64", typeMap["col_bigserial"])
	assert.Equal(t, "int32", typeMap["col_int"])
	assert.Equal(t, "string", typeMap["col_varchar"])
	assert.Equal(t, "string", typeMap["col_text"])
	assert.Equal(t, "bool", typeMap["col_bool"])
	assert.Equal(t, "float64", typeMap["col_float"])
	assert.Equal(t, "float64", typeMap["col_decimal"])
	assert.Equal(t, "time.Time", typeMap["col_timestamp"])
	assert.Equal(t, "string", typeMap["col_uuid"])
	assert.Equal(t, "json.RawMessage", typeMap["col_jsonb"])
}

func TestParser_PrimaryKeyFlags(t *testing.T) {
	p := sqlparser.New()

	sql := `CREATE TABLE events (
		id BIGSERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		payload JSONB
	)`

	got, err := p.ParseCreateTable(sql)
	require.NoError(t, err)

	require.Len(t, got.Columns, 3)
	assert.True(t, got.Columns[0].IsPrimary, "id should be primary key")
	assert.True(t, got.Columns[0].IsSerial, "id should be serial")
	assert.False(t, got.Columns[1].IsPrimary)
	assert.True(t, got.Columns[1].IsNotNull)
	assert.True(t, got.Columns[2].IsNullable, "payload should be nullable")
}
