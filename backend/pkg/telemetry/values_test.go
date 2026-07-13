package telemetry

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/apache/arrow/go/v17/parquet"
	"github.com/apache/arrow/go/v17/parquet/pqarrow"

	"ragpack/pkg/telemetry/ingestion"
	"ragpack/pkg/telemetry/query"
	"ragpack/pkg/telemetry/trace"
)

// readRecord reads a single-file parquet table back into one Arrow record,
// so column-level assertions can catch a hand-indexed b.Field(i) mismatch
// that a row-count check alone would miss.
func readRecord(t *testing.T, path string) arrow.Record {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	tbl, err := pqarrow.ReadTable(context.Background(), f, parquet.NewReaderProperties(nil), pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		t.Fatalf("read table %s: %v", path, err)
	}
	defer tbl.Release()

	tr := array.NewTableReader(tbl, tbl.NumRows())
	defer tr.Release()
	if !tr.Next() {
		t.Fatalf("%s: want at least one record", path)
	}
	rec := tr.Record()
	rec.Retain()
	return rec
}

func column(t *testing.T, rec arrow.Record, name string) arrow.Array {
	t.Helper()
	idx := rec.Schema().FieldIndices(name)
	if len(idx) == 0 {
		t.Fatalf("column %q not found in schema", name)
	}
	return rec.Column(idx[0])
}

func TestQueryEventsColumnValuesRoundTrip(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})

	apiKey := "key-123"
	filters := `{"foo":"bar"}`
	r.Record(&query.Event{
		EventID:        "q1",
		CollectionSlug: "docs",
		APIKeyID:       &apiKey,
		Origin:         "public",
		Endpoint:       "rag",
		TopK:           5,
		Hybrid:         query.HybridSettings{FullTextWeight: 0.3, SemanticWeight: 0.7, RRFK: 60, FullTextLimit: 20},
		FiltersJSON:    &filters,
		EmbedModel:     "text-embedding-3-small",
		Results: []query.ResultStat{
			{SourceName: "a.pdf", Similarity: 0.9, BM25Score: 1.2, RRFScore: 0.5},
			{SourceName: "b.pdf", Similarity: 0.8, BM25Score: 1.1, RRFScore: 0.4},
		},
		Status:  "complete",
		TotalMs: 123,
	})
	r.Close()

	files := parquetFiles(t, filepath.Join(dir, "query_events"))
	if len(files) != 1 {
		t.Fatalf("want 1 parquet file, got %d", len(files))
	}
	rec := readRecord(t, files[0])
	defer rec.Release()

	if got := column(t, rec, "event_id").(*array.String).Value(0); got != "q1" {
		t.Errorf("event_id: got %q, want q1", got)
	}
	if got := column(t, rec, "api_key_id").(*array.String).Value(0); got != apiKey {
		t.Errorf("api_key_id: got %q, want %q", got, apiKey)
	}
	if got := column(t, rec, "endpoint").(*array.String).Value(0); got != "rag" {
		t.Errorf("endpoint: got %q, want rag", got)
	}
	if got := column(t, rec, "embed_model").(*array.String).Value(0); got != "text-embedding-3-small" {
		t.Errorf("embed_model: got %q", got)
	}
	if got := column(t, rec, "top_k").(*array.Int32).Value(0); got != 5 {
		t.Errorf("top_k: got %d, want 5", got)
	}
	if got := column(t, rec, "total_ms").(*array.Int64).Value(0); got != 123 {
		t.Errorf("total_ms: got %d, want 123", got)
	}

	hybrid := column(t, rec, "hybrid_settings").(*array.Struct)
	if got := hybrid.Field(0).(*array.Float64).Value(0); got != 0.3 {
		t.Errorf("hybrid.full_text_weight: got %v, want 0.3", got)
	}
	if got := hybrid.Field(1).(*array.Float64).Value(0); got != 0.7 {
		t.Errorf("hybrid.semantic_weight: got %v, want 0.7", got)
	}
	if got := hybrid.Field(2).(*array.Float64).Value(0); got != 60 {
		t.Errorf("hybrid.rrf_k: got %v, want 60", got)
	}
	if got := hybrid.Field(3).(*array.Int32).Value(0); got != 20 {
		t.Errorf("hybrid.full_text_limit: got %v, want 20", got)
	}

	results := column(t, rec, "results").(*array.List)
	start, end := results.ValueOffsets(0)
	if end-start != 2 {
		t.Fatalf("results: want 2 entries, got %d", end-start)
	}
	resultVals := results.ListValues().(*array.Struct)
	sourceNames := resultVals.Field(0).(*array.String)
	if got := sourceNames.Value(int(start)); got != "a.pdf" {
		t.Errorf("results[0].source_name: got %q, want a.pdf", got)
	}
	if got := sourceNames.Value(int(start) + 1); got != "b.pdf" {
		t.Errorf("results[1].source_name: got %q, want b.pdf", got)
	}
	rrfScores := resultVals.Field(3).(*array.Float64)
	if got := rrfScores.Value(int(start)); got != 0.5 {
		t.Errorf("results[0].rrf_score: got %v, want 0.5", got)
	}

	// A field never set on this event must round-trip as null, not a zero value
	// that could be misread as "confirmed zero cost/tokens".
	if !column(t, rec, "llm_cost_usd").(*array.Float64).IsNull(0) {
		t.Error("llm_cost_usd: want null for a non-RAG-cost event, got a value")
	}
}

func TestIngestionEventsNullableFieldsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})

	tokens := int64(42)
	cost := 0.0021
	r.Record(&ingestion.Event{JobID: "j1", DocumentID: "d1", CollectionSlug: "docs", Status: "complete", EmbedTokens: &tokens, EmbedCostUSD: &cost})
	r.Record(&ingestion.Event{JobID: "j2", DocumentID: "d2", CollectionSlug: "docs", Status: "complete"}) // local model: both nil
	r.Close()

	files := parquetFiles(t, filepath.Join(dir, "ingestion_events"))
	if len(files) != 1 {
		t.Fatalf("want 1 parquet file, got %d", len(files))
	}
	rec := readRecord(t, files[0])
	defer rec.Release()

	jobIDs := column(t, rec, "job_id").(*array.String)
	embedTokens := column(t, rec, "embed_tokens").(*array.Int64)
	embedCost := column(t, rec, "embed_cost_usd").(*array.Float64)

	// Order isn't guaranteed by the buffer, so match rows by job_id.
	var row0, row1 int
	if jobIDs.Value(0) == "j1" {
		row0, row1 = 0, 1
	} else {
		row0, row1 = 1, 0
	}

	if embedTokens.IsNull(row0) || embedTokens.Value(row0) != 42 {
		t.Errorf("j1 embed_tokens: want 42, got null=%v value=%v", embedTokens.IsNull(row0), embedTokens.Value(row0))
	}
	if embedCost.IsNull(row0) || embedCost.Value(row0) != 0.0021 {
		t.Errorf("j1 embed_cost_usd: want 0.0021, got null=%v value=%v", embedCost.IsNull(row0), embedCost.Value(row0))
	}
	if !embedTokens.IsNull(row1) {
		t.Errorf("j2 embed_tokens: want null, got %v", embedTokens.Value(row1))
	}
	if !embedCost.IsNull(row1) {
		t.Errorf("j2 embed_cost_usd: want null, got %v", embedCost.Value(row1))
	}
}

func TestQueryTracesChunksRoundTrip(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Enabled: true, Dir: dir, RetentionDays: 14, MaxSizeMB: 500})

	header := "## Section 1"
	r.Record(&trace.Event{
		EventID: "q1",
		Chunks: []trace.Chunk{
			{SourceName: "a.pdf", ChunkHeader: &header, ChunkText: "first chunk", Similarity: 0.9, BM25Score: 1.5},
			{SourceName: "b.pdf", ChunkText: "second chunk, no header", Similarity: 0.8, BM25Score: 1.1},
		},
	})
	r.Close()

	files := parquetFiles(t, filepath.Join(dir, "query_traces"))
	if len(files) != 1 {
		t.Fatalf("want 1 parquet file, got %d", len(files))
	}
	rec := readRecord(t, files[0])
	defer rec.Release()

	chunks := column(t, rec, "chunks").(*array.List)
	start, end := chunks.ValueOffsets(0)
	if end-start != 2 {
		t.Fatalf("chunks: want 2 entries, got %d", end-start)
	}
	chunkVals := chunks.ListValues().(*array.Struct)
	texts := chunkVals.Field(2).(*array.String)
	headers := chunkVals.Field(1).(*array.String)

	if got := texts.Value(int(start)); got != "first chunk" {
		t.Errorf("chunks[0].chunk_text: got %q", got)
	}
	if headers.IsNull(int(start)) || headers.Value(int(start)) != "## Section 1" {
		t.Errorf("chunks[0].chunk_header: want %q, got null=%v", "## Section 1", headers.IsNull(int(start)))
	}
	if got := texts.Value(int(start) + 1); got != "second chunk, no header" {
		t.Errorf("chunks[1].chunk_text: got %q", got)
	}
	if !headers.IsNull(int(start) + 1) {
		t.Errorf("chunks[1].chunk_header: want null, got %q", headers.Value(int(start)+1))
	}
}
