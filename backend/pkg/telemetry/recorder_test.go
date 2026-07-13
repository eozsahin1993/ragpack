package telemetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v17/parquet/file"

	"ragpack/pkg/telemetry/ingestion"
	"ragpack/pkg/telemetry/query"
	"ragpack/pkg/telemetry/trace"
)

func TestRecorderWritesParquetOnClose(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})

	prompt := "prompt"
	answer := "answer"
	r.Record(&ingestion.Event{JobID: "j1", DocumentID: "d1", CollectionSlug: "docs", Status: "complete", ChunkCount: 3})
	r.Record(&query.Event{EventID: "q1", CollectionSlug: "docs", Endpoint: "rag", Status: "complete",
		Results: []query.ResultStat{{SourceName: "a.pdf", Similarity: 91.5}}})
	r.Record(&query.Event{CollectionSlug: "docs", Endpoint: "query", Status: "failed"})
	r.Record(&trace.Event{EventID: "q1", Chunks: []trace.Chunk{{SourceName: "a.pdf", ChunkText: "text"}},
		FormattedPrompt: &prompt, Answer: &answer})
	r.Record(&trace.Event{Chunks: []trace.Chunk{{SourceName: "orphan"}}}) // no EventID — must be dropped
	r.Close()

	wantRows := map[string]int64{"ingestion_events": 1, "query_events": 2, "query_traces": 1}
	for table, want := range wantRows {
		files := parquetFiles(t, filepath.Join(dir, table))
		if len(files) != 1 {
			t.Fatalf("%s: want 1 parquet file, got %d", table, len(files))
		}
		rdr, err := file.OpenParquetFile(files[0], false)
		if err != nil {
			t.Fatalf("%s: open parquet: %v", table, err)
		}
		if got := rdr.NumRows(); got != want {
			t.Errorf("%s: want %d rows, got %d", table, want, got)
		}
		rdr.Close()
	}
}

func TestNilAndDisabledRecorderAreNoOps(t *testing.T) {
	var r *Recorder
	r.Record(&query.Event{})
	r.Record(&ingestion.Event{})
	r.Record(&trace.Event{})
	r.Close()

	if got := New(Config{Enabled: false}); got != nil {
		t.Fatalf("disabled config: want nil recorder, got %v", got)
	}
}

func TestRedactTextDropsQueryTextAndTraces(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500, RedactText: true})
	q := "sensitive question"
	r.Record(&query.Event{EventID: "q1", QueryText: &q, Status: "complete"})
	r.Record(&trace.Event{EventID: "q1", Chunks: []trace.Chunk{{ChunkText: "secret"}}})
	r.Close()

	if files := parquetFiles(t, filepath.Join(dir, "query_traces")); len(files) != 0 {
		t.Errorf("redact: query_traces should be empty, got %d files", len(files))
	}
	if files := parquetFiles(t, filepath.Join(dir, "query_events")); len(files) != 1 {
		t.Errorf("redact: query_events should still be written, got %d files", len(files))
	}
}

func parquetFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".parquet" {
			out = append(out, path)
		}
		return nil
	})
	return out
}
