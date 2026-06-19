package analyzer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vinaycharlie01/sql-to-go-mcp/pkg/analyzer"
)

func TestAnalyzer_Analyze(t *testing.T) {
	a := analyzer.New()
	ctx := context.Background()

	tests := []struct {
		name          string
		query         string
		wantCodes     []string
		wantNoIssues  bool
		wantErr       bool
	}{
		{
			name:         "empty query",
			query:        "",
			wantErr:      true,
		},
		{
			name:         "clean query",
			query:        "SELECT id, name FROM users WHERE id = $1 LIMIT 10",
			wantNoIssues: true,
		},
		{
			name:      "select star",
			query:     "SELECT * FROM users LIMIT 10",
			wantCodes: []string{"SELECT_STAR"},
		},
		{
			name:      "update without where",
			query:     "UPDATE users SET name = 'foo'",
			wantCodes: []string{"MISSING_WHERE"},
		},
		{
			name:      "delete without where",
			query:     "DELETE FROM users",
			wantCodes: []string{"MISSING_WHERE"},
		},
		{
			name:      "leading wildcard like",
			query:     "SELECT id FROM users WHERE name LIKE '%foo' LIMIT 10",
			wantCodes: []string{"LEADING_WILDCARD"},
		},
		{
			name:      "select without limit",
			query:     "SELECT id, name FROM users WHERE active = true",
			wantCodes: []string{"MISSING_LIMIT"},
		},
		{
			name:      "not in subquery",
			query:     "SELECT id FROM users WHERE id NOT IN (SELECT user_id FROM banned) LIMIT 10",
			wantCodes: []string{"NOT_IN_SUBQUERY"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := a.Analyze(ctx, tc.query)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.wantNoIssues {
				assert.Empty(t, got.Issues)
				return
			}

			codes := make([]string, len(got.Issues))
			for i, issue := range got.Issues {
				codes[i] = issue.Code
			}
			for _, wantCode := range tc.wantCodes {
				assert.Contains(t, codes, wantCode, "expected issue code %s", wantCode)
			}
		})
	}
}

func TestAnalyzer_BenchmarkCodeGenerated(t *testing.T) {
	a := analyzer.New()
	got, err := a.Analyze(context.Background(), "SELECT id, name FROM users WHERE id = $1 LIMIT 1")
	require.NoError(t, err)
	assert.NotEmpty(t, got.BenchmarkCode)
	assert.Contains(t, got.BenchmarkCode, "BenchmarkQuery")
}
