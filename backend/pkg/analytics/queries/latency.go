package queries

import (
	"context"
	"database/sql"
	"time"
)

type LatencyBucket struct {
	Endpoint    string  `json:"endpoint"` // query | rag
	SampleCount int64   `json:"sample_count"`
	P50Ms       float64 `json:"p50_ms"`
	P95Ms       float64 `json:"p95_ms"`
}

const latencySQL = `
SELECT endpoint, count(*) AS sample_count,
       quantile_cont(total_ms, 0.5) AS p50_ms,
       quantile_cont(total_ms, 0.95) AS p95_ms
FROM query_events
WHERE occurred_at >= ? AND status = 'complete'
GROUP BY endpoint ORDER BY endpoint`

// Latency returns p50/p95 total_ms for completed query/RAG calls, split by
// endpoint. Failed calls are excluded — their total_ms reflects an
// early-return error path, not real query latency.
func Latency(ctx context.Context, db *sql.DB, dir string, cutoff time.Time) ([]LatencyBucket, error) {
	ok, err := ensureView(ctx, db, dir, "query_events")
	if err != nil || !ok {
		return []LatencyBucket{}, err
	}

	rows, err := db.QueryContext(ctx, latencySQL, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []LatencyBucket{}
	for rows.Next() {
		var b LatencyBucket
		if err := rows.Scan(&b.Endpoint, &b.SampleCount, &b.P50Ms, &b.P95Ms); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
