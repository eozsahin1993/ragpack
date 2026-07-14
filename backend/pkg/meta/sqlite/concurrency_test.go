package sqlite

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"ragpack/pkg/meta"
)

// TestUpdatePromptConcurrentPartialUpdatesDontClobber guards the exact race
// UpdatePrompt used to have: two concurrent partial updates to the same row,
// each touching a different field. With more than one SQLite connection in
// flight (see sqlite.go's SetMaxOpenConns), a read-full-row-then-write-full-row
// implementation lets whichever write lands second silently revert the
// other's field to its stale pre-read value.
func TestUpdatePromptConcurrentPartialUpdatesDontClobber(t *testing.T) {
	ms := newTestStore(t)
	ctx := context.Background()

	// A fresh prompt per round, racing a Name-touching update against a
	// Content-touching update. The Name update reuses the prompt's own
	// current name (rather than an actual rename) so the slug — the lookup
	// key both goroutines share — never moves out from under either one;
	// that keeps the test isolated to the clobber bug (a stale-read field
	// overwriting a fresher concurrent write), not a "the row's key changed
	// mid-lookup" race, which is a separate concern.
	const rounds = 50
	for i := 0; i < rounds; i++ {
		p, err := ms.CreatePrompt(ctx, meta.CreatePromptInput{Name: fmt.Sprintf("prompt-%d", i), Content: "original"})
		if err != nil {
			t.Fatalf("round %d: create prompt: %v", i, err)
		}

		newName := p.Name
		newContent := "updated content"

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := ms.UpdatePrompt(ctx, p.Slug, meta.UpdatePromptInput{Content: &newContent}); err != nil {
				t.Errorf("round %d: update content: %v", i, err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := ms.UpdatePrompt(ctx, p.Slug, meta.UpdatePromptInput{Name: &newName}); err != nil {
				t.Errorf("round %d: update name: %v", i, err)
			}
		}()
		wg.Wait()

		final, err := ms.GetPromptBySlug(ctx, slugify(newName))
		if err != nil {
			t.Fatalf("round %d: get final prompt: %v", i, err)
		}
		if final.Content != newContent {
			t.Errorf("round %d: Content was clobbered: got %q, want %q", i, final.Content, newContent)
		}
		if final.Name != newName {
			t.Errorf("round %d: Name was clobbered: got %q, want %q", i, final.Name, newName)
		}
	}
}

// TestConcurrentWritesAcrossTablesDontError exercises WAL + busy_timeout +
// a real connection pool under sustained concurrent writes to different
// tables (jobs, documents, api keys) — the scenario SetMaxOpenConns(1) used
// to rule out by construction. No SQLITE_BUSY should surface with the
// busy_timeout pragma in place, and every write should succeed.
func TestConcurrentWritesAcrossTablesDontError(t *testing.T) {
	ms := newTestStore(t)
	ctx := context.Background()

	col, err := ms.CreateCollection(ctx, meta.CreateCollectionInput{
		Name: "concurrency-test", EmbedModel: "test-model", VectorDim: 8,
	})
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}

	const workers = 25
	var wg sync.WaitGroup
	var keyCounter atomic.Int64
	errs := make(chan error, workers*3)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			job, err := ms.CreateJob(ctx, col.ID, "file://test", "text/plain", meta.JobIntentIngest, false, nil, nil)
			if err != nil {
				errs <- err
				return
			}
			if err := ms.UpdateJobStatus(ctx, job.ID, meta.JobStatusComplete, nil); err != nil {
				errs <- err
				return
			}

			doc, err := ms.CreateDocument(ctx, col.ID, job.ID, "file://test", "text/plain", nil)
			if err != nil {
				errs <- err
				return
			}
			completeStatus := meta.DocumentStatusComplete
			three := 3
			if err := ms.UpdateDocument(ctx, doc.ID, meta.DocumentPatch{Status: &completeStatus, ChunkCount: &three}); err != nil {
				errs <- err
				return
			}

			plaintext := fmt.Sprintf("plaintext-%d", keyCounter.Add(1))
			if _, err := ms.CreateAPIKey(ctx, "concurrent-key", plaintext, []meta.GrantInput{{Permission: meta.PermissionRead}}, nil); err != nil {
				errs <- err
				return
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent write failed: %v", err)
	}
}
