package lancedb

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/eozsahin1993/lancedb-go/pkg/contracts"
	sdk "github.com/eozsahin1993/lancedb-go/pkg/lancedb"

	"ragpack/pkg/db"
	"ragpack/pkg/db/lancedb/migrations"
	"ragpack/pkg/meta"
)

type VectorDb struct {
	conn contracts.IConnection

	locksMu       sync.Mutex
	optimizeLocks map[string]*sync.Mutex
}

func New() *VectorDb {
	return &VectorDb{optimizeLocks: make(map[string]*sync.Mutex)}
}

// tableOptimizeLock serializes compaction per table (in-process): a bulk
// re-ingest can put every ingester worker through Delete+Compact on the
// same collection table at once, and retry alone doesn't survive N workers
// all retrying into each other. Only one worker compacting a table at a
// time makes the "Retryable commit conflict" structurally impossible
// (from this process) rather than something to hope retries outlast.
func (l *VectorDb) tableOptimizeLock(tableName string) *sync.Mutex {
	l.locksMu.Lock()
	defer l.locksMu.Unlock()
	m, ok := l.optimizeLocks[tableName]
	if !ok {
		m = &sync.Mutex{}
		l.optimizeLocks[tableName] = m
	}
	return m
}

func (l *VectorDb) Connect(ctx context.Context, connectionUrl string) error {
	conn, err := sdk.Connect(ctx, connectionUrl, nil)
	if err != nil {
		return fmt.Errorf("lancedb: connectivity failure: %w", err)
	}
	l.conn = conn
	return nil
}

func (l *VectorDb) MigrateAll(ctx context.Context, collections []meta.Collection) error {
	return migrations.MigrateAll(ctx, l.conn, collections)
}

// Close releases the underlying native connection. Not part of the db.VectorDb
// interface — the production server holds one connection for its whole
// process lifetime and never needs to close it early, so callers that do
// (e.g. tests opening a fresh connection per test case) use the concrete
// type directly.
func (l *VectorDb) Close() error {
	return l.conn.Close()
}

func (l *VectorDb) DropTable(ctx context.Context, name string) error {
	if err := l.conn.DropTable(ctx, name); err != nil {
		return fmt.Errorf("lancedb: drop table %s: %w", name, err)
	}
	return nil
}

func (l *VectorDb) CreateTable(ctx context.Context, name string, collectionID string, vectorDim int) error {
	schema, err := sdk.NewSchema(chunkArrowSchema(vectorDim))
	if err != nil {
		return fmt.Errorf("lancedb: schema build failed: %w", err)
	}

	tbl, err := l.conn.CreateTable(ctx, name, schema)
	if err != nil {
		return fmt.Errorf("lancedb: table provisioning failed for %s: %w", name, err)
	}
	defer tbl.Close()

	if err := tbl.CreateIndexWithName(ctx, []string{colChunkText}, contracts.IndexTypeFts, colChunkText); err != nil {
		return fmt.Errorf("lancedb: create FTS index on %s.%s: %w", name, colChunkText, err)
	}

	if err := migrations.MarkUpToDate(ctx, l.conn, collectionID); err != nil {
		return fmt.Errorf("lancedb: mark migrations for %s: %w", name, err)
	}
	return nil
}

func (l *VectorDb) InsertBatch(ctx context.Context, tableName string, records []db.ChunkDbRecord) error {
	if len(records) == 0 {
		return nil
	}
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: unable to open target table %s: %w", tableName, err)
	}
	defer tbl.Close()

	arrowRecord, err := chunksToArrowRecord(records, len(records[0].Vector))
	if err != nil {
		return fmt.Errorf("lancedb: failed to build arrow record: %w", err)
	}
	defer arrowRecord.Release()

	if err := tbl.Add(ctx, arrowRecord, nil); err != nil {
		return fmt.Errorf("lancedb: batch commit failed on table %s: %w", tableName, err)
	}
	return nil
}

// DeleteChunksByDocument is the unconditional final-deletion path (document delete) — Delete() hides a row
// from a plain Select immediately (verified empirically, no tombstone window), so compact is storage-reclaim only.
func (l *VectorDb) DeleteChunksByDocument(ctx context.Context, tableName, documentID string) error {
	return l.deleteChunksByDocumentFilter(ctx, tableName, fmt.Sprintf("document_id = '%s'", documentID))
}

// DeleteChunksByIDs deletes exactly the given chunk IDs within one document (chunkIDs read pre-reingest, see ListChunkIDsByDocument).
func (l *VectorDb) DeleteChunksByIDs(ctx context.Context, tableName, documentID string, chunkIDs []string) error {
	if len(chunkIDs) == 0 {
		return nil
	}
	quoted := make([]string, len(chunkIDs))
	for i, id := range chunkIDs {
		quoted[i] = fmt.Sprintf("'%s'", id) // uuids only — no quoting risk
	}
	filter := fmt.Sprintf("document_id = '%s' AND id IN (%s)", documentID, strings.Join(quoted, ","))
	return l.deleteChunksByDocumentFilter(ctx, tableName, filter)
}

func (l *VectorDb) deleteChunksByDocumentFilter(ctx context.Context, tableName, filter string) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	lock := l.tableOptimizeLock(tableName)
	lock.Lock()
	defer lock.Unlock()

	if err := tbl.Delete(ctx, filter); err != nil {
		return fmt.Errorf("lancedb: delete chunks on %s: %w", tableName, err)
	}

	materialize := true
	action := contracts.OptimizeAction{
		Kind:       contracts.OptimizeCompact,
		Compaction: contracts.CompactionParams{MaterializeDeletions: &materialize},
	}
	if err := optimizeWithRetry(ctx, tbl, action); err != nil {
		return fmt.Errorf("lancedb: compact deletes on %s: %w", tableName, err)
	}
	return nil
}

// ChunkVectorsForHashes narrow-projects (hash + vector only) the given
// document's hashes that already exist — bounded per call, not whole-document.
func (l *VectorDb) ChunkVectorsForHashes(ctx context.Context, tableName, documentID string, hashes []string) (map[string][]float32, error) {
	if len(hashes) == 0 {
		return nil, nil
	}
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	quoted := make([]string, len(hashes))
	for i, h := range hashes {
		quoted[i] = fmt.Sprintf("'%s'", h) // hex sha256 output only — no quoting risk
	}
	rows, err := tbl.Select(ctx, contracts.QueryConfig{
		Where:   fmt.Sprintf("document_id = '%s' AND chunk_hash IN (%s)", documentID, strings.Join(quoted, ",")),
		Columns: []string{colChunkHash, colVector}, // narrow projection — no text, no metadata
	})
	if err != nil {
		return nil, fmt.Errorf("lancedb: select chunk vectors for %s: %w", documentID, err)
	}

	out := make(map[string][]float32, len(rows))
	for i, row := range rows {
		hash, err := extractString(row, colChunkHash)
		if err != nil {
			return nil, fmt.Errorf("lancedb: row %d: %w", i, err)
		}
		vec, err := extractFloat32Slice(row, colVector)
		if err != nil {
			return nil, fmt.Errorf("lancedb: row %d: %w", i, err)
		}
		out[hash] = vec
	}
	return out, nil
}

// ListChunkIDsByDocument narrow-projects (id only) every chunk ID for a document — read before a reingest starts.
func (l *VectorDb) ListChunkIDsByDocument(ctx context.Context, tableName, documentID string) ([]string, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	rows, err := tbl.Select(ctx, contracts.QueryConfig{
		Where:   fmt.Sprintf("document_id = '%s'", documentID),
		Columns: []string{colID},
	})
	if err != nil {
		return nil, fmt.Errorf("lancedb: list chunk ids for %s: %w", documentID, err)
	}
	ids := make([]string, len(rows))
	for i, row := range rows {
		id, err := extractString(row, colID)
		if err != nil {
			return nil, fmt.Errorf("lancedb: row %d: %w", i, err)
		}
		ids[i] = id
	}
	return ids, nil
}

const maxOptimizeRetries = 5

// optimizeWithRetry retries on LanceDB's "Retryable commit conflict" — concurrent workers optimizing the same table can race for the same base version.
func optimizeWithRetry(ctx context.Context, tbl contracts.ITable, action contracts.OptimizeAction) error {
	var err error
	for attempt := 0; attempt < maxOptimizeRetries; attempt++ {
		if _, err = tbl.OptimizeWithAction(ctx, action); err == nil {
			return nil
		}
		if !isRetryableCommitConflict(err) {
			return err
		}
		if sleepErr := sleepBackoff(ctx, attempt); sleepErr != nil {
			return err
		}
	}
	return err
}

func isRetryableCommitConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Retryable commit conflict")
}

// sleepBackoff returns ctx.Err() if the context is done before the delay elapses.
func sleepBackoff(ctx context.Context, attempt int) error {
	delay := time.Duration(50*(1<<attempt))*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ListChunksByDocumentPage returns one page of chunks for a document plus the
// total matching count, for chunk-browsing UI. total is fetched via a
// separate id-only Select since LanceDB's Count is unfiltered-only.
func (l *VectorDb) ListChunksByDocumentPage(ctx context.Context, tableName, documentID string, limit, offset int) ([]db.ChunkDbRecord, int, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, 0, fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	filter := fmt.Sprintf("document_id = '%s'", documentID)

	countRows, err := tbl.Select(ctx, contracts.QueryConfig{Where: filter, Columns: []string{colID}})
	if err != nil {
		return nil, 0, fmt.Errorf("lancedb: count chunks for document %s: %w", documentID, err)
	}
	total := len(countRows)

	rows, err := tbl.Select(ctx, contracts.QueryConfig{Where: filter, Limit: &limit, Offset: &offset})
	if err != nil {
		return nil, 0, fmt.Errorf("lancedb: list chunks for document %s: %w", documentID, err)
	}

	chunks := make([]db.ChunkDbRecord, 0, len(rows))
	for i, row := range rows {
		rec, err := rowToChunk(row)
		if err != nil {
			return nil, 0, fmt.Errorf("lancedb: row %d: %w", i, err)
		}
		chunks = append(chunks, rec)
	}
	return chunks, total, nil
}

// ListChunkMetadataByDocument returns every chunk's metadata slot values for
// a document (unpaginated — GetMetadata's consistency check needs all rows),
// fetching only the metadata slot columns rather than full chunk records.
func (l *VectorDb) ListChunkMetadataByDocument(ctx context.Context, tableName, documentID string) ([]db.ChunkDbRecord, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	rows, err := tbl.Select(ctx, contracts.QueryConfig{
		Where:   fmt.Sprintf("document_id = '%s'", documentID),
		Columns: metadataSlotColumns(),
	})
	if err != nil {
		return nil, fmt.Errorf("lancedb: list chunk metadata for document %s: %w", documentID, err)
	}

	chunks := make([]db.ChunkDbRecord, 0, len(rows))
	for _, row := range rows {
		chunks = append(chunks, rowToChunkMetadataOnly(row))
	}
	return chunks, nil
}

func (l *VectorDb) UpdateChunks(ctx context.Context, tableName, documentID string, patch db.ChunkPatch) error {
	if patch.IsEmpty() {
		return nil
	}
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	if err := tbl.Update(ctx, fmt.Sprintf("document_id = '%s'", documentID), patch.ToMap()); err != nil {
		return fmt.Errorf("lancedb: update chunks for document %s: %w", documentID, err)
	}
	return nil
}

// QuerySimilarVectors runs vector search, optionally filtered, and fuses in
// an FTS pass on chunk_text when keywordQuery is non-empty (hybrid search).
func (l *VectorDb) QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int, filter string, keywordQuery string, hybrid db.HybridSettings) ([]db.ChunkQueryResult, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open query table failed %s: %w", tableName, err)
	}
	defer tbl.Close()

	vectorRows, err := tbl.Select(ctx, contracts.QueryConfig{
		VectorSearch: &contracts.VectorSearch{Column: colVector, Vector: vector, K: topK},
		Where:        filter,
	})
	if err != nil {
		return nil, fmt.Errorf("lancedb: query execution failed on %s: %w", tableName, err)
	}
	vectorResults, err := mapResultsToChunks(vectorRows)
	if err != nil {
		return nil, fmt.Errorf("lancedb: mapping vector results on %s: %w", tableName, err)
	}

	if keywordQuery == "" {
		return vectorResults, nil
	}

	var ftsRows []map[string]any
	if filter != "" {
		ftsRows, err = tbl.FullTextSearchWithFilter(ctx, colChunkText, keywordQuery, filter)
	} else {
		ftsRows, err = tbl.FullTextSearch(ctx, colChunkText, keywordQuery)
	}
	if err != nil {
		return nil, fmt.Errorf("lancedb: FTS query failed on %s: %w", tableName, err)
	}
	// FullTextSearch(WithFilter) has no limit param; truncate manually.
	if len(ftsRows) > hybrid.FullTextLimit {
		ftsRows = ftsRows[:hybrid.FullTextLimit]
	}
	ftsResults, err := mapResultsToChunks(ftsRows)
	if err != nil {
		return nil, fmt.Errorf("lancedb: mapping FTS results on %s: %w", tableName, err)
	}

	return db.MergeWeightedRRF(vectorResults, ftsResults, topK, hybrid), nil
}

func (l *VectorDb) OptimizeIndex(ctx context.Context, tableName string) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	lock := l.tableOptimizeLock(tableName)
	lock.Lock()
	defer lock.Unlock()

	if err := optimizeWithRetry(ctx, tbl, contracts.OptimizeAction{Kind: contracts.OptimizeIndex}); err != nil {
		return fmt.Errorf("lancedb: optimize index on %s: %w", tableName, err)
	}
	return nil
}

func (l *VectorDb) CreateMetadataIndex(ctx context.Context, tableName, colName, fieldType string) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	var indexType contracts.IndexType
	switch fieldType {
	case "str", "bool":
		indexType = contracts.IndexTypeBitmap
	case "num", "date":
		indexType = contracts.IndexTypeBTree
	case "arr":
		indexType = contracts.IndexTypeLabelList
	default:
		return fmt.Errorf("lancedb: unknown metadata field type %q", fieldType)
	}

	if err := tbl.CreateIndexWithName(ctx, []string{colName}, indexType, colName); err != nil {
		return fmt.Errorf("lancedb: create metadata index on %s.%s: %w", tableName, colName, err)
	}
	return nil
}

func (l *VectorDb) DropMetadataIndex(ctx context.Context, tableName, indexName string) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	if err := tbl.DropIndex(ctx, indexName); err != nil {
		return fmt.Errorf("lancedb: drop index %s on %s: %w", indexName, tableName, err)
	}
	return nil
}

func (l *VectorDb) NullMetadataSlot(ctx context.Context, tableName, colName string) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	if err := tbl.Update(ctx, "true", map[string]any{colName: nil}); err != nil {
		return fmt.Errorf("lancedb: null slot %s on %s: %w", colName, tableName, err)
	}
	return nil
}
