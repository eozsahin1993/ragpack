// Package queries holds the five named dashboard questions and the view
// management they share — kept separate from pkg/analytics itself, which
// only owns the DuckDB connection lifecycle (opening it, applying the
// memory/thread/timeout caps, closing it).
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Store bundles the DuckDB connection with the telemetry directory and a
// cache of which tables already have a view defined. Every query function
// takes a *Store rather than a bare *sql.DB, so this package has no
// knowledge of Engine, but still shares one connection and one
// view-creation cache across all five dashboard questions.
type Store struct {
	db  *sql.DB
	dir string

	mu      sync.Mutex
	created map[string]bool
}

func NewStore(db *sql.DB, dir string) *Store {
	return &Store{db: db, dir: dir, created: make(map[string]bool)}
}

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

// ensureView creates table as a view over its current Parquet files the
// first time it's asked for, and never again after that — read_parquet's
// glob re-expands fresh on every SELECT against an existing view, so a
// view, once created, stays correct as files rotate or new ones land; there
// is nothing to redo. The mutex + created-table cache exist for a
// different reason: five dashboard panels query concurrently (see
// pkg/api/analytics), and without serializing the *first* creation, two
// goroutines racing CREATE OR REPLACE VIEW on the same table at the same
// time hit DuckDB's catalog write-write conflict — reproduced via the
// actual web-admin page, not caught by any test that only ever called one
// query method at a time.
func (s *Store) ensureView(ctx context.Context, table string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.created[table] {
		return true, nil
	}

	ok, err := hasFiles(s.dir, table)
	if err != nil || !ok {
		return false, err
	}
	// dir comes from config.TelemetryConfig.Dir (trusted, not request
	// input) — quoting for the literal, not sanitizing against attack.
	glob := strings.ReplaceAll(tableGlob(s.dir, table), "'", "''")
	stmt := fmt.Sprintf(
		`CREATE OR REPLACE VIEW %s AS SELECT * FROM read_parquet('%s', union_by_name=true)`,
		table, glob,
	)
	if _, err := s.db.ExecContext(ctx, stmt); err != nil {
		return false, err
	}
	s.created[table] = true
	return true, nil
}
