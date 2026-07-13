package telemetry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFakeParquet(t *testing.T, path string, size int, modTime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

func TestJanitorRemovesFilesPastRetention(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "ingestion_events", "2020-01-01", "old.parquet")
	recent := filepath.Join(dir, "ingestion_events", "2020-01-01", "recent.parquet")
	writeFakeParquet(t, old, 10, time.Now().Add(-20*24*time.Hour))
	writeFakeParquet(t, recent, 10, time.Now())

	r := &Recorder{cfg: Config{Dir: dir, RetentionDays: 14, MaxSizeMB: 500}}
	r.enforce()

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("want file past the retention window removed")
	}
	if _, err := os.Stat(recent); err != nil {
		t.Errorf("want recent file kept, got %v", err)
	}
}

func TestJanitorEnforcesSizeCapOldestFirst(t *testing.T) {
	dir := t.TempDir()
	const fileSize = 400 * 1024 // 3 * 400KB > 1MB cap, forces eviction
	now := time.Now()
	oldest := filepath.Join(dir, "query_events", "d", "1.parquet")
	middle := filepath.Join(dir, "query_events", "d", "2.parquet")
	newest := filepath.Join(dir, "query_events", "d", "3.parquet")
	writeFakeParquet(t, oldest, fileSize, now.Add(-3*time.Hour))
	writeFakeParquet(t, middle, fileSize, now.Add(-2*time.Hour))
	writeFakeParquet(t, newest, fileSize, now.Add(-1*time.Hour))

	r := &Recorder{cfg: Config{Dir: dir, RetentionDays: 14, MaxSizeMB: 1}}
	r.enforce()

	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Error("want the oldest file evicted first to satisfy the size cap")
	}
	if _, err := os.Stat(middle); err != nil {
		t.Error("want the middle file kept")
	}
	if _, err := os.Stat(newest); err != nil {
		t.Error("want the newest file kept")
	}
}

func TestJanitorKeepsFilesWithinRetentionAndCap(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ingestion_events", "d", "1.parquet")
	writeFakeParquet(t, p, 10, time.Now())

	r := &Recorder{cfg: Config{Dir: dir, RetentionDays: 14, MaxSizeMB: 500}}
	r.enforce()

	if _, err := os.Stat(p); err != nil {
		t.Errorf("want file within retention/cap kept, got %v", err)
	}
}
