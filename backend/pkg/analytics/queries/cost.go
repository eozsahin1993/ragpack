package queries

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type CollectionCost struct {
	CollectionSlug string  `json:"collection_slug"`
	TotalCostUSD   float64 `json:"total_cost_usd"`
}

// embed_cost_usd only exists on ingestion_events — query_events has no
// per-query embedding cost column (see pkg/telemetry/query/schema.go), so
// the query/RAG side's cost is llm_cost_usd alone.
const ingestionCostSQL = `
SELECT collection_slug, COALESCE(embed_cost_usd, 0) AS cost
FROM ingestion_events WHERE occurred_at >= ?`

const queryCostSQL = `
SELECT collection_slug, COALESCE(llm_cost_usd, 0) AS cost
FROM query_events WHERE occurred_at >= ?`

// CostByCollection returns SUM(embed_cost_usd) from ingestion plus
// SUM(llm_cost_usd) from query/RAG, grouped by collection_slug.
func CostByCollection(ctx context.Context, db *sql.DB, dir string, cutoff time.Time) ([]CollectionCost, error) {
	ingOK, err := ensureView(ctx, db, dir, "ingestion_events")
	if err != nil {
		return nil, err
	}
	qOK, err := ensureView(ctx, db, dir, "query_events")
	if err != nil {
		return nil, err
	}
	if !ingOK && !qOK {
		return []CollectionCost{}, nil
	}

	var parts []string
	var args []any
	if ingOK {
		parts = append(parts, ingestionCostSQL)
		args = append(args, cutoff)
	}
	if qOK {
		parts = append(parts, queryCostSQL)
		args = append(args, cutoff)
	}
	stmt := `SELECT collection_slug, SUM(cost) AS total_cost_usd FROM (` +
		strings.Join(parts, " UNION ALL ") + `) c GROUP BY collection_slug ORDER BY total_cost_usd DESC`

	rows, err := db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []CollectionCost{}
	for rows.Next() {
		var c CollectionCost
		if err := rows.Scan(&c.CollectionSlug, &c.TotalCostUSD); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
