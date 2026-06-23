package ingester

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	chunkerpkg "ragpack/pkg/chunker"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/fetcher"
	"ragpack/pkg/meta"
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
		f, err := fetcher.New(ctx, job.FileUri)
		if err != nil {
			return 0, fmt.Errorf("build fetcher: %w", err)
		}
		reader, err = f.Fetch(ctx, job.FileUri)
		if err != nil {
			return 0, fmt.Errorf("fetch: %w", err)
		}
	}
	defer reader.Close()

	// Delete any chunks from a previous partial or failed attempt so retries are idempotent.
	if err := wp.vectorDb.DeleteChunksByDocument(ctx, collection.TableName, documentID); err != nil {
		return 0, fmt.Errorf("delete stale chunks: %w", err)
	}

	chunker, err := chunkerpkg.New(job.MimeType, wp.chunkCfg)
	if err != nil {
		return 0, fmt.Errorf("build chunker: %w", err)
	}

	chunks, err := chunker.Chunk(ctx, reader)
	if err != nil {
		return 0, fmt.Errorf("chunk: %w", err)
	}

	emb, err := wp.registry.Get(collection.EmbedModel)
	if err != nil {
		return 0, fmt.Errorf("embedder: %w", err)
	}

	now := time.Now().UTC()

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		texts := make([]string, len(batch))
		for j, ch := range batch {
			texts[j] = ch.Text
		}

		if err := wp.limiter.Wait(ctx); err != nil {
			return 0, fmt.Errorf("rate limiter: %w", err)
		}

		vectors, err := emb.Embed(ctx, texts)
		if err != nil {
			return 0, fmt.Errorf("embed batch starting at chunk %d: %w", i, err)
		}

		for j, ch := range batch {
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ch.Text)))
			rec := db.ChunkDbRecord{
				ID:         uuid.NewString(),
				DocumentID: documentID,
				ChunkHash:  hash,
				ChunkIndex: ch.Index,
				Vector:     embedder.Normalize(vectors[j]),
				CreatedAt:  now,
				UpdatedAt:  now,
				MimeType:   job.MimeType,
				FileUri:    job.FileUri,
				SourceName: collection.Name,
				ChunkText:  &ch.Text,
			}
			if err := wp.vectorDb.InsertRecord(ctx, collection.TableName, rec); err != nil {
				return 0, fmt.Errorf("insert chunk %d: %w", ch.Index, err)
			}
		}
	}

	return len(chunks), nil
}
