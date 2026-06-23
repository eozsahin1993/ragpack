package ingester

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"

	chunkerpkg "ragpack/pkg/chunker"
	"ragpack/pkg/db"
	"ragpack/pkg/fetcher"
	"ragpack/pkg/meta"
)

func (wp *WorkerPool) processJob(ctx context.Context, item queueItem) error {
	if err := wp.metaStore.UpdateJobStatus(ctx, item.job.ID, meta.JobStatusProcessing, nil); err != nil {
		return err
	}

	if err := wp.process(ctx, item); err != nil {
		errStr := err.Error()
		_ = wp.metaStore.UpdateJobStatus(ctx, item.job.ID, meta.JobStatusFailed, &errStr)
		return err
	}

	return wp.metaStore.UpdateJobStatus(ctx, item.job.ID, meta.JobStatusComplete, nil)
}

func (wp *WorkerPool) process(ctx context.Context, item queueItem) error {
	job := item.job

	collection, err := wp.metaStore.GetCollectionByID(ctx, job.CollectionID)
	if err != nil {
		return fmt.Errorf("get collection: %w", err)
	}

	reader := item.reader
	if reader == nil {
		f, err := fetcher.New(ctx, job.FileUri)
		if err != nil {
			return fmt.Errorf("build fetcher: %w", err)
		}
		reader, err = f.Fetch(ctx, job.FileUri)
		if err != nil {
			return fmt.Errorf("fetch: %w", err)
		}
	}

	chunker, err := chunkerpkg.New(job.MimeType, wp.chunkCfg)
	if err != nil {
		return fmt.Errorf("build chunker: %w", err)
	}

	chunks, err := chunker.Chunk(ctx, reader)
	if err != nil {
		return fmt.Errorf("chunk: %w", err)
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
			return fmt.Errorf("rate limiter: %w", err)
		}

		vectors, err := wp.emb.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("embed batch starting at chunk %d: %w", i, err)
		}

		for j, ch := range batch {
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ch.Text)))
			rec := db.ChunkDbRecord{
				ID:         uuid.NewString(),
				ChunkHash:  hash,
				ChunkIndex: ch.Index,
				Vector:     vectors[j],
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
	}

	return nil
}
