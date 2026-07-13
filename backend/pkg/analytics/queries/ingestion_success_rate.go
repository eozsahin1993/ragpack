package queries

import (
	"context"
	"time"
)

type MimeSuccessRate struct {
	MimeType    string  `json:"mime_type"`
	TotalCount  int64   `json:"total_count"`
	SuccessRate float64 `json:"success_rate"`
}

const ingestionSuccessRateSQL = `
SELECT mime_type, count(*) AS total_count,
       avg(CASE WHEN status = 'complete' THEN 1.0 ELSE 0.0 END) AS success_rate
FROM ingestion_events
WHERE occurred_at >= ?
GROUP BY mime_type ORDER BY success_rate ASC, total_count DESC`

// IngestionSuccessRate returns the success rate per mime_type over
// ingestion_events, worst-performing type first — surfaces whether a
// specific file type is systematically failing to ingest.
func IngestionSuccessRate(ctx context.Context, s *Store, cutoff time.Time) ([]MimeSuccessRate, error) {
	ok, err := s.ensureView(ctx, "ingestion_events")
	if err != nil || !ok {
		return []MimeSuccessRate{}, err
	}

	rows, err := s.db.QueryContext(ctx, ingestionSuccessRateSQL, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []MimeSuccessRate{}
	for rows.Next() {
		var m MimeSuccessRate
		if err := rows.Scan(&m.MimeType, &m.TotalCount, &m.SuccessRate); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
