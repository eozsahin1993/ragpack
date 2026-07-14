//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/config"
	"ragpack/pkg/ingester"
	"ragpack/test/integration/helpers"
)

// paragraphs returns n distinct, multi-sentence paragraphs (blank-line
// separated) — long enough for the paragraph chunker (test app: ChunkSize
// 500) to produce multiple chunks per document.
func paragraphs(n int, seed string) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "Paragraph %d about %s. It has several sentences so it takes up real space. "+
			"Here is another sentence to pad it out further. And one more for good measure.\n\n", i, seed)
	}
	return b.String()
}

func TestForceRefreshReusesUnchangedChunks(t *testing.T) {
	a, _, emb := helpers.NewFullTestApp(t)
	slug := helpers.CreateCollection(t, a.Admin, "Reuse Test")

	content := paragraphs(6, "gardening")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(content))
	}))
	defer srv.Close()
	fileURI := srv.URL + "/doc.txt"

	job := helpers.IngestURI(t, a.Admin, slug, fileURI, "text/plain", false, false)
	helpers.WaitForJob(t, a.Admin, job["id"].(string))
	doc := helpers.WaitForDocumentByURI(t, a.Admin, slug, fileURI)
	if doc["status"] != "complete" {
		t.Fatalf("initial ingest: expected complete, got %v (error=%v)", doc["status"], doc["error"])
	}
	firstChunkCount := doc["chunk_count"]
	callsAfterFirst := emb.CallCount()
	if callsAfterFirst == 0 {
		t.Fatalf("expected embed calls on first ingest, got 0")
	}

	// Force refresh with byte-identical content — every chunk's fingerprint
	// should match the stored one, so this must make zero new embed calls.
	job2 := helpers.IngestURI(t, a.Admin, slug, fileURI, "text/plain", true, true)
	helpers.WaitForJob(t, a.Admin, job2["id"].(string))
	doc2 := helpers.WaitForDocumentByURI(t, a.Admin, slug, fileURI)
	if doc2["status"] != "complete" {
		t.Fatalf("force refresh: expected complete, got %v (error=%v)", doc2["status"], doc2["error"])
	}
	if doc2["chunk_count"] != firstChunkCount {
		t.Errorf("expected chunk_count unchanged at %v, got %v", firstChunkCount, doc2["chunk_count"])
	}
	if got := emb.CallCount(); got != callsAfterFirst {
		t.Errorf("expected no new embed calls on unchanged-content refresh: calls went from %d to %d", callsAfterFirst, got)
	}
}

func TestForceRefreshEmbedsChangedContent(t *testing.T) {
	a, _, emb := helpers.NewFullTestApp(t)
	slug := helpers.CreateCollection(t, a.Admin, "Reembed Test")

	var mu sync.Mutex
	content := paragraphs(6, "gardening")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(content))
	}))
	defer srv.Close()
	fileURI := srv.URL + "/doc.txt"

	job := helpers.IngestURI(t, a.Admin, slug, fileURI, "text/plain", false, false)
	helpers.WaitForJob(t, a.Admin, job["id"].(string))
	doc := helpers.WaitForDocumentByURI(t, a.Admin, slug, fileURI)
	if doc["status"] != "complete" {
		t.Fatalf("initial ingest: expected complete, got %v (error=%v)", doc["status"], doc["error"])
	}
	callsAfterFirst := emb.CallCount()

	// Change every paragraph's content, then force refresh — this must
	// increase embed calls, guarding against a reuse check that's
	// accidentally always-true (e.g. matching on chunk_index instead of hash).
	mu.Lock()
	content = paragraphs(6, "woodworking")
	mu.Unlock()

	job2 := helpers.IngestURI(t, a.Admin, slug, fileURI, "text/plain", true, true)
	helpers.WaitForJob(t, a.Admin, job2["id"].(string))
	doc2 := helpers.WaitForDocumentByURI(t, a.Admin, slug, fileURI)
	if doc2["status"] != "complete" {
		t.Fatalf("force refresh: expected complete, got %v (error=%v)", doc2["status"], doc2["error"])
	}
	if got := emb.CallCount(); got <= callsAfterFirst {
		t.Errorf("expected new embed calls on changed-content refresh: calls stayed at %d", got)
	}
}

// etagServer serves content at /doc.txt honoring If-None-Match, so the
// auto-refresh scheduler's conditional fetch can be exercised against a
// real 200-vs-304 HTTP round trip.
type etagServer struct {
	mu      sync.Mutex
	content string
	etag    string
	hits    int
}

func newETagServer(content, etag string) (*httptest.Server, *etagServer) {
	s := &etagServer{content: content, etag: etag}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.hits++
		if r.Header.Get("If-None-Match") == s.etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", s.etag)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(s.content))
	}))
	return srv, s
}

func (s *etagServer) set(content, etag string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.content, s.etag = content, etag
}

func (s *etagServer) hitCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits
}

func TestAutoRefreshSweepPicksUpETagChange(t *testing.T) {
	a, ms, emb := helpers.NewFullTestApp(t)
	wp, ok := a.Ingester.(*ingester.WorkerPool)
	if !ok {
		t.Fatalf("expected *ingester.WorkerPool, got %T", a.Ingester)
	}

	slug := helpers.CreateCollection(t, a.Admin, "Auto Refresh Test")
	_, colResp := helpers.DoJSON(t, a.Admin, http.MethodGet, "/admin/collections/"+slug, nil)
	colID := colResp["id"].(string)

	srv, backing := newETagServer(paragraphs(6, "gardening"), "v1")
	defer srv.Close()
	fileURI := srv.URL + "/doc.txt"

	// Enable auto-refresh at the test-configured minimum.
	resp, _ := helpers.DoJSON(t, a.Admin, http.MethodPatch, "/admin/collections/id/"+colID,
		map[string]any{"refresh_enabled": true, "refresh_interval_seconds": config.DefaultMinCollectionRefreshSeconds})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("enable auto-refresh: expected 200, got %d", resp.StatusCode)
	}

	job := helpers.IngestURI(t, a.Admin, slug, fileURI, "text/plain", false, false)
	helpers.WaitForJob(t, a.Admin, job["id"].(string))
	doc := helpers.WaitForDocumentByURI(t, a.Admin, slug, fileURI)
	if doc["status"] != "complete" {
		t.Fatalf("initial ingest: expected complete, got %v (error=%v)", doc["status"], doc["error"])
	}
	docID := doc["id"].(string)
	if doc["last_etag"] != nil {
		t.Fatalf("expected last_etag nil after a plain (non-conditional) ingest, got %v", doc["last_etag"])
	}
	callsAfterIngest := emb.CallCount()
	ctx := t.Context()

	// First check goes through the real due-gate (CheckNow -> checkDueCollections),
	// exercising ListCollectionsDueForAutoRefresh once for real. Doc has no
	// stored etag, so this is an unconditional fetch — content is unchanged
	// from the ingest, so every chunk reuses its vector, but last_etag still
	// gets stamped now that the source has confirmed one.
	wp.CheckNow(ctx)
	waitForDocField(t, a.Admin, docID, "last_etag", "v1")
	if got := emb.CallCount(); got != callsAfterIngest {
		t.Errorf("expected no new embed calls (content unchanged): calls went from %d to %d", callsAfterIngest, got)
	}

	// TouchCollectionAutoRefreshed just stamped last_auto_refresh_at, so this
	// collection won't show up as "due" again for another 900s — exactly the
	// rate-limiting the real scheduler is supposed to do. The remaining
	// checks are testing change *detection*, not the due-gate, so they go
	// through CheckCollectionNow directly (same detection code, bypassing
	// only the "is it time yet" gate) against a freshly-read collection.
	col, err := ms.GetCollectionByID(ctx, colID)
	if err != nil {
		t.Fatalf("get collection: %v", err)
	}

	// Change source content + ETag: next check must detect it, refresh, and
	// re-embed (content genuinely changed).
	backing.set(paragraphs(6, "woodworking"), "v2")
	wp.CheckCollectionNow(ctx, col)
	waitForDocField(t, a.Admin, docID, "last_etag", "v2")
	if got := emb.CallCount(); got <= callsAfterIngest {
		t.Errorf("expected new embed calls after content+etag change: calls stayed at %d", got)
	}
	callsAfterChange := emb.CallCount()

	col, err = ms.GetCollectionByID(ctx, colID)
	if err != nil {
		t.Fatalf("get collection: %v", err)
	}

	// Nothing changed: server should answer 304, no new job, no new embed calls.
	hitsBefore := backing.hitCount()
	wp.CheckCollectionNow(ctx, col)
	time.Sleep(300 * time.Millisecond) // let a wrongly-created job have a chance to run
	if got := emb.CallCount(); got != callsAfterChange {
		t.Errorf("expected no embed calls on an unchanged (304) check: calls went from %d to %d", callsAfterChange, got)
	}
	if backing.hitCount() <= hitsBefore {
		t.Errorf("expected the scheduler to have hit the server again (for the 304), hit count stayed at %d", hitsBefore)
	}

	_, colResp2 := helpers.DoJSON(t, a.Admin, http.MethodGet, "/admin/collections/id/"+colID, nil)
	if colResp2["last_auto_refresh_at"] == nil {
		t.Errorf("expected last_auto_refresh_at to be stamped after CheckNow")
	}
}

// waitForDocField polls a document by ID until field equals want, or times out.
func waitForDocField(t *testing.T, app *fiber.App, docID, field string, want any) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		_, doc := helpers.DoJSON(t, app, http.MethodGet, "/admin/documents/"+docID, nil)
		last = doc
		if doc[field] == want {
			return doc
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for document %s.%s == %v, last was %v", docID, field, want, last)
	return nil
}
