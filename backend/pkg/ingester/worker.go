package ingester

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	chunkerpkg "ragpack/pkg/chunker"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
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
		if job.Intent == meta.JobIntentIngest && !job.Force {
			if existing.Status == meta.DocumentStatusComplete {
				// Already ingested — skip.
				return wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusComplete, nil)
			}
			if existing.Status == meta.DocumentStatusIngesting {
				return wp.failJob(ctx, jobID, fmt.Errorf("document is already being ingested"))
			}
		}
		// Reuse existing document — reset status and bind to new job.
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

	chunkCount, processErr := wp.process(ctx, item, docID, collection)
	if processErr != nil {
		errStr := processErr.Error()
		if err := wp.metaStore.UpdateDocumentStatus(ctx, docID, meta.DocumentStatusFailed, 0, &errStr); err != nil {
			log.Printf("ingester: job %s: failed to mark document failed: %v", jobID, err)
		}
		return wp.failJob(ctx, jobID, processErr)
	}

	if err := wp.metaStore.UpdateDocumentStatus(ctx, docID, meta.DocumentStatusComplete, chunkCount, nil); err != nil {
		log.Printf("ingester: job %s: failed to mark document complete: %v", jobID, err)
		return wp.failJob(ctx, jobID, err)
	}

	return wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusComplete, nil)
}

func (wp *WorkerPool) process(ctx context.Context, item queueItem, documentID string, collection meta.Collection) (int, error) {
	job := item.job

	reader := item.reader
	if reader == nil {
		if strings.HasPrefix(job.FileUri, "upload://") {
			return 0, fmt.Errorf("uploaded file is no longer available; please re-submit the file")
		}
		f, err := fetcher.New(ctx, job.FileUri)
		if err != nil {
			return 0, fmt.Errorf("build fetcher: %w", err)
		}
		reader, err = f.Fetch(ctx, job.FileUri)
		if err != nil {
			return 0, fmt.Errorf("fetch: %w", err)
		}
	}
	// reader is closed by the parser inside Parse() — no defer here.

	// Delete any stale chunks before re-embedding (handles retries and refresh).
	if err := wp.vectorDb.DeleteChunksByDocument(ctx, collection.TableName, documentID); err != nil {
		return 0, fmt.Errorf("delete stale chunks: %w", err)
	}

	p, err := parserpkg.New(job.MimeType, job.FileUri)
	if err != nil {
		return 0, fmt.Errorf("build parser: %w", err)
	}

	// Resolve chunk config: collection overrides take precedence over server defaults.
	chunkCfg := wp.chunkCfg
	if collection.ChunkStrategy != nil {
		chunkCfg.Strategy = *collection.ChunkStrategy
	}
	if collection.ChunkSize != nil {
		chunkCfg.ChunkSize = *collection.ChunkSize
	}
	if collection.ChunkOverlap != nil {
		chunkCfg.Overlap = *collection.ChunkOverlap
	}

	c, err := chunkerpkg.New(job.MimeType, chunkCfg)
	if err != nil {
		return 0, fmt.Errorf("build chunker: %w", err)
	}

	emb, err := wp.registry.Get(collection.EmbedModel)
	if err != nil {
		return 0, fmt.Errorf("embedder: %w", err)
	}

	now := time.Now().UTC()
	chunkCount := 0

	// Stream: parser → chunker → embed in batches → insert.
	// Only batchSize chunks are in memory at once.
	var batch []chunkerpkg.Chunk
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		texts := make([]string, len(batch))
		for i, ch := range batch {
			if ch.Header != nil {
				texts[i] = *ch.Header + "\n" + ch.Text
			} else {
				texts[i] = ch.Text
			}
		}
		if err := wp.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}
		vectors, err := emb.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("embed batch at chunk %d: %w", chunkCount, err)
		}
		records := make([]db.ChunkDbRecord, len(batch))
		for i, ch := range batch {
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ch.Text)))
			records[i] = db.ChunkDbRecord{
				ID:          uuid.NewString(),
				DocumentID:  documentID,
				ChunkHash:   hash,
				ChunkIndex:  ch.Index,
				Vector:      embedder.Normalize(vectors[i]),
				CreatedAt:   now,
				UpdatedAt:   now,
				MimeType:    job.MimeType,
				FileUri:     job.FileUri,
				SourceName:  collection.Name,
				ChunkText:   &ch.Text,
				ChunkHeader: ch.Header,
				ExtraJSON:   job.ExtraJSON,
			}
		}
		if err := wp.vectorDb.InsertBatch(ctx, collection.TableName, records); err != nil {
			return fmt.Errorf("insert batch at chunk %d: %w", chunkCount, err)
		}
		chunkCount += len(batch)
		batch = batch[:0]
		return nil
	}

	for chunk, err := range c.Chunk(p.Parse(ctx, reader)) {
		if err != nil {
			return 0, fmt.Errorf("chunk: %w", err)
		}
		batch = append(batch, chunk)
		if len(batch) >= batchSize {
			if err := flush(); err != nil {
				return 0, err
			}
		}
	}
	if err := flush(); err != nil {
		return 0, err
	}

	name := ""
	if t := p.Title(); t != nil {
		name = *t
	}
	if name == "" {
		name = util.NameFromURI(job.FileUri)
	}
	if name != "" {
		if err := wp.metaStore.UpdateDocument(ctx, documentID, meta.DocumentPatch{Name: &name}); err != nil {
			log.Printf("ingester: job %s: failed to save document name: %v", job.ID, err)
		}
	}

	return chunkCount, nil
}

