package queries

import (
	"context"
	"database/sql"
	"time"
)

type MimeFailureRate struct {
	MimeType    string  `json:"mime_type"`
	TotalCount  int64   `json:"total_count"`
	FailureRate float64 `json:"failure_rate"`
}

const ingestionFailureRateSQL = `
SELECT mime_type, count(*) AS total_count,
       avg(CASE WHEN status = 'failed' THEN 1.0 ELSE 0.0 END) AS failure_rate
FROM ingestion_events
WHERE occurred_at >= ?
GROUP BY mime_type ORDER BY failure_rate DESC, total_count DESC`

// IngestionFailureRate returns the failure rate per mime_type over
// ingestion_events — surfaces whether a specific file type is systematically
// failing to ingest.
func IngestionFailureRate(ctx context.Context, db *sql.DB, dir string, cutoff time.Time) ([]MimeFailureRate, error) {
	ok, err := ensureView(ctx, db, dir, "ingestion_events")
	if err != nil || !ok {
		return []MimeFailureRate{}, err
	}

	rows, err := db.QueryContext(ctx, ingestionFailureRateSQL, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []MimeFailureRate{}
	for rows.Next() {
		var m MimeFailureRate
		if err := rows.Scan(&m.MimeType, &m.TotalCount, &m.FailureRate); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
