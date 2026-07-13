package queries

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type VolumePoint struct {
	Day       string `json:"day"`        // YYYY-MM-DD, UTC
	EventType string `json:"event_type"` // ingestion | query | rag
	Count     int64  `json:"count"`
}

const ingestionVolumeSQL = `
SELECT strftime(date_trunc('day', occurred_at), '%Y-%m-%d') AS day,
       'ingestion' AS event_type, count(*) AS count
FROM ingestion_events WHERE occurred_at >= ? GROUP BY 1`

const queryVolumeSQL = `
SELECT strftime(date_trunc('day', occurred_at), '%Y-%m-%d') AS day,
       endpoint AS event_type, count(*) AS count
FROM query_events WHERE occurred_at >= ? GROUP BY 1, 2`

// VolumeOverTime returns ingestion + query/RAG event counts bucketed by UTC
// day, for occurred_at >= cutoff.
func VolumeOverTime(ctx context.Context, db *sql.DB, dir string, cutoff time.Time) ([]VolumePoint, error) {
	ingOK, err := ensureView(ctx, db, dir, "ingestion_events")
	if err != nil {
		return nil, err
	}
	qOK, err := ensureView(ctx, db, dir, "query_events")
	if err != nil {
		return nil, err
	}
	if !ingOK && !qOK {
		return []VolumePoint{}, nil
	}

	var parts []string
	var args []any
	if ingOK {
		parts = append(parts, ingestionVolumeSQL)
		args = append(args, cutoff)
	}
	if qOK {
		parts = append(parts, queryVolumeSQL)
		args = append(args, cutoff)
	}
	stmt := strings.Join(parts, " UNION ALL ") + " ORDER BY 1, 2"

	rows, err := db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []VolumePoint{}
	for rows.Next() {
		var p VolumePoint
		if err := rows.Scan(&p.Day, &p.EventType, &p.Count); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
