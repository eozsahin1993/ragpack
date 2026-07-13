// Package queries holds the five named dashboard questions and the view
// management they share — kept separate from pkg/analytics itself, which
// only owns the DuckDB connection lifecycle (opening it, applying the
// memory/thread/timeout caps, closing it). Each function here takes a plain
// *sql.DB and the telemetry directory, so this package has no knowledge of
// Engine at all.
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
)

func tableGlob(dir, table string) string {
	return filepath.Join(dir, table, "*", "*.parquet")
}

// hasFiles reports whether table currently has at least one flushed Parquet
// file. DuckDB's read_parquet errors on a glob matching zero files (confirmed
// via spike: "IO Error: No files found that match the pattern..."), so every
// query function must check this first and short-circuit to a typed empty
// result instead of querying — a fresh install, or telemetry recently
// enabled, can legitimately have zero files under some or all tables.
func hasFiles(dir, table string) (bool, error) {
	matches, err := filepath.Glob(tableGlob(dir, table))
	return len(matches) > 0, err
}

// ensureView (re)creates table as a view over its current Parquet files,
// returning false (no error) if there are none yet. Called at the top of
// every query function rather than once at Engine startup: a low-traffic
// install may have zero files under a table today and some tomorrow.
// read_parquet's glob re-expands fresh on every SELECT against the view, so
// this only has to handle the initially-empty case — it's not needed again
// later just because the janitor rotated files after the view was created.
func ensureView(ctx context.Context, db *sql.DB, dir, table string) (bool, error) {
	ok, err := hasFiles(dir, table)
	if err != nil || !ok {
		return false, err
	}
	// dir comes from config.TelemetryConfig.Dir (trusted, not request
	// input) — quoting for the literal, not sanitizing against attack.
	glob := strings.ReplaceAll(tableGlob(dir, table), "'", "''")
	stmt := fmt.Sprintf(
		`CREATE OR REPLACE VIEW %s AS SELECT * FROM read_parquet('%s', union_by_name=true)`,
		table, glob,
	)
	_, err = db.ExecContext(ctx, stmt)
	return err == nil, err
}
