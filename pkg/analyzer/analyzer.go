package analyzer

import (
	"context"
	"fmt"
	"strings"

	"github.com/vinaycharlie01/sql-to-go-mcp/internal/domain/port"
)

// Analyzer implements port.QueryAnalyzer.
type Analyzer struct{}

// New returns a new Analyzer.
func New() *Analyzer { return &Analyzer{} }

// Analyze inspects a SQL query for common anti-patterns and returns recommendations.
func (a *Analyzer) Analyze(_ context.Context, query string) (*port.QueryAnalysis, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query must not be empty")
	}

	upper := strings.ToUpper(query)
	result := &port.QueryAnalysis{}

	// SELECT *
	if strings.Contains(upper, "SELECT *") || strings.Contains(upper, "SELECT\t*") {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "warning",
			Code:        "SELECT_STAR",
			Description: "SELECT * fetches all columns including unused ones, increasing network overhead and preventing index-only scans.",
		})
		result.Recommendations = append(result.Recommendations,
			"Replace SELECT * with explicit column names to allow index-only scans and reduce data transfer.",
		)
	}

	// Missing WHERE on UPDATE/DELETE
	if (strings.Contains(upper, "UPDATE ") || strings.Contains(upper, "DELETE FROM")) &&
		!strings.Contains(upper, "WHERE") {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "error",
			Code:        "MISSING_WHERE",
			Description: "UPDATE or DELETE without a WHERE clause affects every row in the table.",
		})
		result.Recommendations = append(result.Recommendations,
			"Add a WHERE clause to restrict UPDATE/DELETE to the target rows.",
		)
	}

	// Cartesian JOIN (FROM a, b without JOIN condition)
	fromIdx := strings.Index(upper, "FROM ")
	whereIdx := strings.Index(upper, "WHERE")
	if fromIdx >= 0 {
		fromClause := upper[fromIdx+5:]
		if whereIdx > fromIdx {
			fromClause = upper[fromIdx+5 : whereIdx]
		}
		commaCount := strings.Count(fromClause, ",")
		joinCount := strings.Count(upper, " JOIN ")
		if commaCount >= 1 && joinCount == 0 && !strings.Contains(upper, "IN (") {
			result.Issues = append(result.Issues, port.Issue{
				Severity:    "error",
				Code:        "CARTESIAN_JOIN",
				Description: "Comma-separated tables in FROM without explicit JOIN conditions produce a cartesian product.",
			})
			result.Recommendations = append(result.Recommendations,
				"Replace implicit comma joins with explicit INNER JOIN ... ON clauses.",
			)
		}
	}

	// N+1 signal: SELECT inside a loop pattern (subquery in SELECT list)
	if strings.Contains(upper, "(SELECT ") && strings.Count(upper, "(SELECT ") > 1 {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "warning",
			Code:        "POTENTIAL_N_PLUS_1",
			Description: "Multiple correlated subqueries in the SELECT list may cause N+1 query patterns.",
		})
		result.Recommendations = append(result.Recommendations,
			"Rewrite correlated subqueries as a single JOIN or use a CTE to fetch related data in one round-trip.",
		)
	}

	// Full table scan: LIKE with leading wildcard
	if strings.Contains(upper, "LIKE '%") || strings.Contains(upper, `LIKE "%`) {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "warning",
			Code:        "LEADING_WILDCARD",
			Description: "LIKE with a leading wildcard prevents index usage and causes a full table scan.",
		})
		result.Recommendations = append(result.Recommendations,
			"Consider a full-text search index (GIN/tsvector on PostgreSQL, FULLTEXT on MySQL) for prefix/suffix searches.",
		)
	}

	// Missing LIMIT on SELECT
	if strings.HasPrefix(strings.TrimSpace(upper), "SELECT") &&
		!strings.Contains(upper, "LIMIT") &&
		!strings.Contains(upper, "TOP ") &&
		!strings.Contains(upper, "FETCH ") {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "info",
			Code:        "MISSING_LIMIT",
			Description: "SELECT without LIMIT may return an unbounded result set.",
		})
		result.Recommendations = append(result.Recommendations,
			"Add LIMIT / OFFSET (or keyset pagination) to bound the result set size.",
		)
	}

	// OR on indexed column — inhibits index usage in some databases
	if strings.Count(upper, " OR ") > 2 {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "info",
			Code:        "MULTIPLE_OR",
			Description: "Multiple OR conditions can prevent the optimizer from using a single index.",
		})
		result.Recommendations = append(result.Recommendations,
			"Consider rewriting multiple OR conditions as UNION ALL or using IN (...) for better index utilization.",
		)
	}

	// NOT IN with subquery — often poor performance
	if strings.Contains(upper, "NOT IN (SELECT") {
		result.Issues = append(result.Issues, port.Issue{
			Severity:    "warning",
			Code:        "NOT_IN_SUBQUERY",
			Description: "NOT IN with a subquery can be slow and behaves unexpectedly with NULLs.",
		})
		result.Recommendations = append(result.Recommendations,
			"Replace NOT IN (SELECT ...) with NOT EXISTS (...) for better NULL handling and potential index usage.",
		)
	}

	// Function on indexed column in WHERE
	fnPatterns := []string{"UPPER(", "LOWER(", "DATE(", "YEAR(", "MONTH(", "TO_CHAR(", "CAST("}
	for _, fn := range fnPatterns {
		if whereIdx >= 0 && strings.Contains(upper[whereIdx:], fn) {
			result.Issues = append(result.Issues, port.Issue{
				Severity:    "warning",
				Code:        "FUNCTION_ON_COLUMN",
				Description: fmt.Sprintf("Applying %s to a column in WHERE prevents index usage.", strings.TrimSuffix(fn, "(")),
			})
			result.Recommendations = append(result.Recommendations,
				"Create a functional index or rewrite the WHERE clause to avoid wrapping indexed columns in functions.",
			)
			break
		}
	}

	result.BenchmarkCode = buildBenchmarkStub(query)
	return result, nil
}

func buildBenchmarkStub(query string) string {
	escaped := strings.ReplaceAll(query, "`", "'")
	return fmt.Sprintf(`func BenchmarkQuery(b *testing.B) {
	db := testutil.MustOpenDB(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.QueryContext(context.Background(), %q)
		if err != nil {
			b.Fatal(err)
		}
		rows.Close()
	}
}`, escaped)
}
