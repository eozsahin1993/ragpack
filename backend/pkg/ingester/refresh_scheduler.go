package ingester

import (
	"context"
	"log"
	"net/http"
	"time"

	"ragpack/pkg/fetcher"
	"ragpack/pkg/meta"
)

// autoRefreshCheckInterval is the scheduler's own tick — independent of any
// collection's refresh_interval_seconds, which is checked against last_auto_refresh_at.
const autoRefreshCheckInterval = 5 * time.Minute

// refreshScheduler is a second ticker alongside loop()'s requeue poller.
func (wp *WorkerPool) refreshScheduler(ctx context.Context) {
	ticker := time.NewTicker(autoRefreshCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wp.checkDueCollections(ctx)
		}
	}
}

// CheckNow runs one auto-refresh check immediately instead of waiting for the next tick — for tests only (concrete *WorkerPool, not the Ingester interface).
func (wp *WorkerPool) CheckNow(ctx context.Context) {
	wp.checkDueCollections(ctx)
}

// CheckCollectionNow checks one collection directly, bypassing the "is it due yet" gate — for tests exercising change-detection without waiting out refresh_interval_seconds.
func (wp *WorkerPool) CheckCollectionNow(ctx context.Context, col meta.Collection) {
	wp.checkCollection(ctx, col)
}

func (wp *WorkerPool) checkDueCollections(ctx context.Context) {
	due, err := wp.metaStore.ListCollectionsDueForAutoRefresh(ctx, time.Now())
	if err != nil {
		log.Printf("ingester: auto-refresh: list due collections: %v", err)
		return
	}
	for _, col := range due {
		wp.checkCollection(ctx, col)
	}
}

// checkCollection stamps last_auto_refresh_at once per collection, not once per document (write contention).
func (wp *WorkerPool) checkCollection(ctx context.Context, col meta.Collection) {
	docs, err := wp.metaStore.ListDocuments(ctx, meta.DocumentFilter{CollectionID: &col.ID}, meta.DocumentSort{}, 10_000, 0)
	if err != nil {
		log.Printf("ingester: auto-refresh: list documents for %s: %v", col.ID, err)
		return
	}
	for _, doc := range docs {
		wp.checkAndMaybeRefresh(ctx, col, doc)
	}
	if err := wp.metaStore.TouchCollectionAutoRefreshed(ctx, col.ID, time.Now()); err != nil {
		log.Printf("ingester: auto-refresh: touch collection %s: %v", col.ID, err)
	}
}

// checkAndMaybeRefresh writes nothing on the common case (304 or unsupported source) — last_etag is stamped later, in the resulting job's success path.
func (wp *WorkerPool) checkAndMaybeRefresh(ctx context.Context, col meta.Collection, doc meta.Document) {
	src, err := fetcher.New(ctx, doc.FileUri)
	if err != nil {
		return // upload:// and other non-refreshable sources land here
	}
	condSrc, ok := src.(fetcher.ConditionalFetcher) // http(s):// and s3:// both implement this — no scheme branch here
	if !ok {
		return
	}

	var etag, lastMod string
	if doc.LastETag != nil {
		etag = *doc.LastETag
	}
	if col.LastAutoRefreshAt != nil {
		lastMod = col.LastAutoRefreshAt.UTC().Format(http.TimeFormat)
	}

	result, err := condSrc.FetchConditional(ctx, doc.FileUri, etag, lastMod)
	if err != nil || result.NotModified {
		return // nothing changed, or a transient check failure — retried next check
	}

	// Reuse the body already in hand rather than paying for a second GET inside the job.
	job, err := wp.metaStore.CreateJob(ctx, doc.CollectionID, doc.FileUri, doc.MimeType,
		meta.JobIntentAutoRefresh, false, doc.ExtraJSON, nil)
	if err != nil {
		result.Body.Close()
		log.Printf("ingester: auto-refresh: create job for %s: %v", doc.ID, err)
		return
	}

	var newETag *string
	if result.ETag != "" {
		newETag = &result.ETag
	}

	// Bypasses Submit() (no etag param there) — same queue a manual refresh uses. Non-blocking: a full queue leaves the job pending for the next requeue poll.
	select {
	case wp.queue <- queueItem{job: job, reader: result.Body, etag: newETag}:
	default:
		result.Body.Close()
	}
}
