package ingester

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	chunkerpkg "ragpack/pkg/chunker"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/meta"
)

// batchFlusher fingerprints, reuses-or-embeds, and inserts one document's chunks in batches, holding state across process()'s repeated flush() calls.
type batchFlusher struct {
	wp         *WorkerPool
	documentID string
	collection meta.Collection
	job        meta.Job
	sourceName string
	chunkCfg   chunkerpkg.Config
	emb        embedder.Embedder
	meta       metadataSlots
	stats      *ingestStats
	now        time.Time

	chunkCount int
}

// flush embeds-or-reuses and inserts one batch — a fingerprint match against a still-live old row (not deleted until process() finishes every batch) reuses its vector instead of embedding.
func (f *batchFlusher) flush(ctx context.Context, batch []chunkerpkg.Chunk) error {
	if len(batch) == 0 {
		return nil
	}
	wp := f.wp

	embeddedTexts := make([]string, len(batch))
	hashes := make([]string, len(batch))
	for i, ch := range batch {
		if ch.Header != nil {
			embeddedTexts[i] = *ch.Header + "\n" + ch.Text
		} else {
			embeddedTexts[i] = ch.Text
		}
		// Fingerprint, not a raw text hash — covers pipeline version and chunk
		// config too, so a chunk whose *effective* embedded text is unchanged
		// reuses its old vector instead of paying for a fresh embed call.
		hashes[i] = chunkerpkg.Fingerprint(f.chunkCfg, embeddedTexts[i])
	}

	// Bounded lookup — only this batch's hashes, not the whole document.
	existing, err := wp.vectorDb.ChunkVectorsForHashes(ctx, f.collection.TableName, f.documentID, hashes)
	if err != nil {
		return fmt.Errorf("look up existing chunk vectors: %w", err)
	}

	vectors := make([][]float32, len(batch))
	var toEmbed []int
	for i, h := range hashes {
		if v, ok := existing[h]; ok {
			vectors[i] = v // reused — already normalized when first stored
			f.stats.reusedChunks++
			continue
		}
		toEmbed = append(toEmbed, i)
	}

	if len(toEmbed) > 0 {
		texts := make([]string, len(toEmbed))
		for i, at := range toEmbed {
			texts[i] = embeddedTexts[at]
		}
		waitStart := time.Now()
		if err := wp.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}
		f.stats.waitMs += time.Since(waitStart).Milliseconds()

		embedStart := time.Now()
		embedded, usage, err := f.emb.Embed(ctx, texts)
		f.stats.embedMs += time.Since(embedStart).Milliseconds()
		if err != nil {
			return fmt.Errorf("embed batch at chunk %d: %w", f.chunkCount, err)
		}
		f.stats.embedTokens += usage.TotalTokens
		for i, v := range embedded {
			vectors[toEmbed[i]] = embedder.Normalize(v)
		}
	}

	records := make([]db.ChunkDbRecord, len(batch))
	for i, ch := range batch {
		records[i] = db.ChunkDbRecord{
			ID:           uuid.NewString(),
			DocumentID:   f.documentID,
			ChunkHash:    hashes[i],
			ChunkIndex:   ch.Index,
			Vector:       vectors[i],
			CreatedAt:    f.now,
			UpdatedAt:    f.now,
			MimeType:     f.job.MimeType,
			FileUri:      f.job.FileUri,
			SourceName:   f.sourceName,
			ChunkText:    &ch.Text,
			ChunkHeader:  ch.Header,
			ExtraJSON:    f.job.ExtraJSON,
			MetadataStr:  f.meta.str,
			MetadataNum:  f.meta.num,
			MetadataBool: f.meta.boolean,
			MetadataDate: f.meta.date,
			MetadataArr:  f.meta.arr,
		}
	}
	insertStart := time.Now()
	if err := wp.vectorDb.InsertBatch(ctx, f.collection.TableName, records); err != nil {
		return fmt.Errorf("insert batch at chunk %d: %w", f.chunkCount, err)
	}
	f.stats.insertMs += time.Since(insertStart).Milliseconds()
	f.chunkCount += len(batch)
	return nil
}
