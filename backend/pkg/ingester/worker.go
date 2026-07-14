package ingester

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"strings"
	"time"

	chunkerpkg "ragpack/pkg/chunker"
	"ragpack/pkg/fetcher"
	"ragpack/pkg/meta"
	parserpkg "ragpack/pkg/parser"
	"ragpack/pkg/util"
)

// failJob marks the job as failed and returns the original error.
func (wp *WorkerPool) failJob(ctx context.Context, jobID string, err error) error {
	errStr := err.Error()
	if logErr := wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusFailed, &errStr); logErr != nil {
		log.Printf("ingester: job %s: failed to persist failure status: %v", jobID, logErr)
	}
	return err
}

func (wp *WorkerPool) processJob(ctx context.Context, item queueItem) error {
	job := item.job
	jobID := job.ID

	if err := wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusProcessing, nil); err != nil {
		return err
	}

	collection, err := wp.metaStore.GetCollectionByID(ctx, job.CollectionID)
	if err != nil {
		return wp.failJob(ctx, jobID, err)
	}

	existing, err := wp.metaStore.FindDocumentByFileUri(ctx, job.CollectionID, job.FileUri)
	if err != nil {
		return wp.failJob(ctx, jobID, err)
	}

	var docID string
	if existing != nil {
		if existing.Status == meta.DocumentStatusIngesting {
			return wp.failJob(ctx, jobID, meta.ErrDocumentAlreadyIngesting) // fast path — ResetDocument enforces this atomically too
		}
		if job.Intent == meta.JobIntentIngest && !job.Force && existing.Status == meta.DocumentStatusComplete {
			return wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusComplete, nil) // already ingested — skip
		}
		// Reuse existing document — ResetDocument's conditional UPDATE is what actually closes a race against another job.
		doc, err := wp.metaStore.ResetDocument(ctx, existing.ID, jobID)
		if err != nil {
			return wp.failJob(ctx, jobID, err)
		}
		docID = doc.ID
	} else {
		doc, err := wp.metaStore.CreateDocument(ctx, job.CollectionID, jobID, job.FileUri, job.MimeType, job.ExtraJSON)
		if err != nil {
			return wp.failJob(ctx, jobID, err)
		}
		docID = doc.ID
	}

	processStart := time.Now()
	var stats ingestStats
	chunkCount, processErr := wp.safeProcess(ctx, item, docID, collection, &stats)
	wp.recordIngestion(job, docID, collection, chunkCount, processErr, time.Since(processStart).Milliseconds(), stats)
	if processErr != nil {
		errStr := processErr.Error()
		failedStatus := meta.DocumentStatusFailed
		zero := 0
		failPatch := meta.DocumentPatch{Status: &failedStatus, ChunkCount: &zero, Error: &errStr}
		if err := wp.metaStore.UpdateDocument(ctx, docID, failPatch); err != nil {
			log.Printf("ingester: job %s: failed to mark document failed: %v", jobID, err)
		}
		return wp.failJob(ctx, jobID, processErr)
	}

	// item.etag is non-nil only for an auto-refresh job that found a new ETag.
	completeStatus := meta.DocumentStatusComplete
	completePatch := meta.DocumentPatch{Status: &completeStatus, ChunkCount: &chunkCount, ClearError: true, LastETag: item.etag}
	if err := wp.metaStore.UpdateDocument(ctx, docID, completePatch); err != nil {
		log.Printf("ingester: job %s: failed to mark document complete: %v", jobID, err)
		return wp.failJob(ctx, jobID, err)
	}

	return wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusComplete, nil)
}

// safeProcess guards process against panics — a bad document must fail that document, not take down the server.
func (wp *WorkerPool) safeProcess(ctx context.Context, item queueItem, documentID string, collection meta.Collection, stats *ingestStats) (count int, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ingester: job %s: panic during processing: %v\n%s", item.job.ID, r, debug.Stack())
			err = fmt.Errorf("internal error during processing: %v", r)
		}
	}()
	return wp.process(ctx, item, documentID, collection, stats)
}

// process fetches, chunks, embeds (reusing vectors where possible), and inserts one document.
func (wp *WorkerPool) process(ctx context.Context, item queueItem, documentID string, collection meta.Collection, stats *ingestStats) (int, error) {
	job := item.job

	reader, err := wp.fetchReader(ctx, item, stats)
	if err != nil {
		return 0, err
	}
	// reader is closed by the parser inside Parse() — no defer here.

	staleIDs, err := wp.vectorDb.ListChunkIDsByDocument(ctx, collection.TableName, documentID)
	if err != nil {
		return 0, fmt.Errorf("list existing chunks: %w", err)
	}
	previousChunkCount := len(staleIDs)

	parser, err := parserpkg.New(job.MimeType, job.FileUri)
	if err != nil {
		return 0, fmt.Errorf("build parser: %w", err)
	}

	chunkCfg := resolveChunkConfig(wp.chunkCfg, collection)
	chunker, err := chunkerpkg.New(job.MimeType, chunkCfg)
	if err != nil {
		return 0, fmt.Errorf("build chunker: %w", err)
	}

	emb, err := wp.registry.Get(collection.EmbedModel)
	if err != nil {
		return 0, fmt.Errorf("embedder: %w", err)
	}

	metaSlots, err := wp.resolveMetadataSlots(ctx, job, collection)
	if err != nil {
		return 0, err
	}

	flusher := &batchFlusher{
		wp:         wp,
		documentID: documentID,
		collection: collection,
		job:        job,
		sourceName: util.NameFromURI(job.FileUri),
		chunkCfg:   chunkCfg,
		emb:        emb,
		meta:       metaSlots,
		stats:      stats,
		now:        time.Now().UTC(),
	}

	// Stream: parser → chunker → flush (embed-or-reuse in batches) → insert.
	// Only batchSize chunks are in memory at once.
	loopStart := time.Now()
	var batch []chunkerpkg.Chunk
	for chunk, err := range chunker.Chunk(parser.Parse(ctx, reader)) {
		if err != nil {
			return 0, fmt.Errorf("chunk: %w", err)
		}
		batch = append(batch, chunk)
		if len(batch) >= batchSize {
			if err := flusher.flush(ctx, batch); err != nil {
				return 0, err
			}
			batch = batch[:0]
		}
	}
	if err := flusher.flush(ctx, batch); err != nil {
		return 0, err
	}
	stats.loopMs = time.Since(loopStart).Milliseconds()

	// Every new/reused row is inserted now, so staleIDs (read up front) is exactly what predates this reingest.
	if err := wp.vectorDb.DeleteChunksByIDs(ctx, collection.TableName, documentID, staleIDs); err != nil {
		return 0, fmt.Errorf("delete stale chunks: %w", err)
	}

	// Clamped at 0: a duplicate paragraph matching the same old hash twice isn't "removal."
	stats.staleChunks = previousChunkCount - stats.reusedChunks
	if stats.staleChunks < 0 {
		stats.staleChunks = 0
	}
	if stats.reusedChunks > 0 || stats.staleChunks > 0 {
		log.Printf("ingester: job %s: reused %d/%d chunks, dropped %d stale", job.ID, stats.reusedChunks, flusher.chunkCount, stats.staleChunks)
	}

	// Fold any new chunks into existing metadata indexes so queries stay fast.
	optimizeStart := time.Now()
	if err := wp.vectorDb.OptimizeIndex(ctx, collection.TableName); err != nil {
		log.Printf("ingester: job %s: optimize index: %v", job.ID, err)
	}
	stats.optimizeMs = time.Since(optimizeStart).Milliseconds()

	wp.saveDocumentName(ctx, documentID, job, parser)

	return flusher.chunkCount, nil
}

// fetchReader returns the direct-upload reader already provided, or fetches one from the source URI.
func (wp *WorkerPool) fetchReader(ctx context.Context, item queueItem, stats *ingestStats) (io.ReadCloser, error) {
	if item.reader != nil {
		return item.reader, nil
	}
	job := item.job
	if strings.HasPrefix(job.FileUri, "upload://") {
		return nil, fmt.Errorf("uploaded file is no longer available; please re-submit the file")
	}
	fetchStart := time.Now()
	src, err := fetcher.New(ctx, job.FileUri)
	if err != nil {
		return nil, fmt.Errorf("build fetcher: %w", err)
	}
	reader, err := src.Fetch(ctx, job.FileUri)
	stats.fetchMs = time.Since(fetchStart).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	return reader, nil
}

// resolveChunkConfig applies collection-level overrides on top of the server default.
func resolveChunkConfig(base chunkerpkg.Config, collection meta.Collection) chunkerpkg.Config {
	cfg := base
	if collection.ChunkStrategy != nil {
		cfg.Strategy = *collection.ChunkStrategy
	}
	if collection.ChunkSize != nil {
		cfg.ChunkSize = *collection.ChunkSize
	}
	if collection.ChunkOverlap != nil {
		cfg.Overlap = *collection.ChunkOverlap
	}
	return cfg
}

// saveDocumentName is best-effort — a failure here doesn't affect ingestion's success/failure.
func (wp *WorkerPool) saveDocumentName(ctx context.Context, documentID string, job meta.Job, parser parserpkg.Parser) {
	name := ""
	if title := parser.Title(); title != nil {
		name = *title
	}
	if name == "" {
		name = util.NameFromURI(job.FileUri)
	}
	if name == "" {
		return
	}
	if err := wp.metaStore.UpdateDocument(ctx, documentID, meta.DocumentPatch{Name: &name}); err != nil {
		log.Printf("ingester: job %s: failed to save document name: %v", job.ID, err)
	}
}
