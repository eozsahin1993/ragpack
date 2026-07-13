package queries

import (
	"context"
	"strings"
	"time"
)

// embed_cost_usd only exists on ingestion_events, so LLMCostUSD is query/RAG-only.
type CollectionCost struct {
	CollectionSlug   string  `json:"collection_slug"`
	IngestionCostUSD float64 `json:"ingestion_cost_usd"`
	LLMCostUSD       float64 `json:"llm_cost_usd"`
}

const ingestionCostSQL = `
SELECT collection_slug, COALESCE(embed_cost_usd, 0) AS ingestion_cost_usd, 0 AS llm_cost_usd
FROM ingestion_events WHERE occurred_at >= ?`

const queryCostSQL = `
SELECT collection_slug, 0 AS ingestion_cost_usd, COALESCE(llm_cost_usd, 0) AS llm_cost_usd
FROM query_events WHERE occurred_at >= ?`

// CostByCollection returns SUM(embed_cost_usd) from ingestion and
// SUM(llm_cost_usd) from query/RAG separately, grouped by collection_slug.
func CostByCollection(ctx context.Context, s *Store, cutoff time.Time) ([]CollectionCost, error) {
	ingOK, err := s.ensureView(ctx, "ingestion_events")
	if err != nil {
		return nil, err
	}
	qOK, err := s.ensureView(ctx, "query_events")
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
	stmt := `SELECT collection_slug,
			SUM(ingestion_cost_usd) AS ingestion_cost_usd,
			SUM(llm_cost_usd) AS llm_cost_usd
		FROM (` + strings.Join(parts, " UNION ALL ") + `) c
		GROUP BY collection_slug ORDER BY SUM(ingestion_cost_usd) + SUM(llm_cost_usd) DESC`

	rows, err := s.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []CollectionCost{}
	for rows.Next() {
		var c CollectionCost
		if err := rows.Scan(&c.CollectionSlug, &c.IngestionCostUSD, &c.LLMCostUSD); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
