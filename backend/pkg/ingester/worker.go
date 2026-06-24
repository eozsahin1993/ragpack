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
)

func (wp *WorkerPool) processJob(ctx context.Context, item queueItem) error {
	jobID := item.job.ID

	if err := wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusProcessing, nil); err != nil {
		return err
	}

	doc, err := wp.metaStore.CreateDocument(ctx, item.job.CollectionID, jobID, item.job.FileUri, item.job.MimeType)
	if err != nil {
		errStr := err.Error()
		if logErr := wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusFailed, &errStr); logErr != nil {
			log.Printf("ingester: job %s: failed to persist failure status: %v", jobID, logErr)
		}
		return err
	}

	chunkCount, processErr := wp.process(ctx, item, doc.ID)
	if processErr != nil {
		errStr := processErr.Error()
		if logErr := wp.metaStore.UpdateDocumentStatus(ctx, doc.ID, meta.DocumentStatusFailed, 0, &errStr); logErr != nil {
			log.Printf("ingester: job %s: failed to mark document failed: %v", jobID, logErr)
		}
		if logErr := wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusFailed, &errStr); logErr != nil {
			log.Printf("ingester: job %s: failed to persist failure status: %v", jobID, logErr)
		}
		return processErr
	}

	// Propagate document status failure so the job is retried. Retry is safe because
	// process() deletes stale chunks before inserting.
	if err := wp.metaStore.UpdateDocumentStatus(ctx, doc.ID, meta.DocumentStatusComplete, chunkCount, nil); err != nil {
		errStr := err.Error()
		log.Printf("ingester: job %s: failed to mark document complete: %v", jobID, err)
		if logErr := wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusFailed, &errStr); logErr != nil {
			log.Printf("ingester: job %s: failed to persist failure status: %v", jobID, logErr)
		}
		return err
	}

	return wp.metaStore.UpdateJobStatus(ctx, jobID, meta.JobStatusComplete, nil)
}

func (wp *WorkerPool) process(ctx context.Context, item queueItem, documentID string) (int, error) {
	job := item.job

	collection, err := wp.metaStore.GetCollectionByID(ctx, job.CollectionID)
	if err != nil {
		return 0, fmt.Errorf("get collection: %w", err)
	}

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

	// Delete any chunks from a previous partial or failed attempt so retries are idempotent.
	if err := wp.vectorDb.DeleteChunksByDocument(ctx, collection.TableName, documentID); err != nil {
		return 0, fmt.Errorf("delete stale chunks: %w", err)
	}

	p, err := parserpkg.New(job.MimeType)
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
			texts[i] = ch.Text
		}
		if err := wp.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}
		vectors, err := emb.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("embed batch at chunk %d: %w", chunkCount, err)
		}
		for i, ch := range batch {
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ch.Text)))
			rec := db.ChunkDbRecord{
				ID:         uuid.NewString(),
				DocumentID: documentID,
				ChunkHash:  hash,
				ChunkIndex: ch.Index,
				Vector:     embedder.Normalize(vectors[i]),
				CreatedAt:  now,
				UpdatedAt:  now,
				MimeType:   job.MimeType,
				FileUri:    job.FileUri,
				SourceName: collection.Name,
				ChunkText:  &ch.Text,
			}
			if err := wp.vectorDb.InsertRecord(ctx, collection.TableName, rec); err != nil {
				return fmt.Errorf("insert chunk %d: %w", ch.Index, err)
			}
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

	return chunkCount, nil
}
