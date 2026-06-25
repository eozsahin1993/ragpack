package lancedb

import (
	"context"
	"fmt"

	"github.com/eozsahin1993/lancedb-go/pkg/contracts"
	sdk "github.com/eozsahin1993/lancedb-go/pkg/lancedb"

	"ragpack/pkg/db"
	"ragpack/pkg/db/lancedb/migrations"
	"ragpack/pkg/meta"
)

type VectorDb struct {
	conn contracts.IConnection
}

func New() *VectorDb {
	return &VectorDb{}
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

func (l *VectorDb) DropTable(ctx context.Context, name string) error {
	if err := l.conn.DropTable(ctx, name); err != nil {
		return fmt.Errorf("lancedb: drop table %s: %w", name, err)
	}
	return nil
}

func (l *VectorDb) CreateTable(ctx context.Context, name string, vectorDim int) error {
	schema, err := sdk.NewSchema(chunkArrowSchema(vectorDim))
	if err != nil {
		return fmt.Errorf("lancedb: schema build failed: %w", err)
	}

	_, err = l.conn.CreateTable(ctx, name, schema)
	if err != nil {
		return fmt.Errorf("lancedb: table provisioning failed for %s: %w", name, err)
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


func (l *VectorDb) DeleteChunksByDocument(ctx context.Context, tableName, documentID string) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	if err := tbl.Delete(ctx, fmt.Sprintf("document_id = '%s'", documentID)); err != nil {
		return fmt.Errorf("lancedb: delete chunks for document %s: %w", documentID, err)
	}

	// Physically remove tombstoned rows so they don't leak into subsequent
	// SelectWithFilter or VectorSearch results.
	if _, err := tbl.Optimize(ctx); err != nil {
		return fmt.Errorf("lancedb: optimize after delete on %s: %w", tableName, err)
	}
	return nil
}

func (l *VectorDb) ListChunksByDocument(ctx context.Context, tableName, documentID string) ([]db.ChunkDbRecord, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open table %s: %w", tableName, err)
	}
	defer tbl.Close()

	rows, err := tbl.SelectWithFilter(ctx, fmt.Sprintf("document_id = '%s'", documentID))
	if err != nil {
		return nil, fmt.Errorf("lancedb: list chunks for document %s: %w", documentID, err)
	}

	chunks := make([]db.ChunkDbRecord, 0, len(rows))
	for i, row := range rows {
		rec, err := rowToChunk(row)
		if err != nil {
			return nil, fmt.Errorf("lancedb: row %d: %w", i, err)
		}
		chunks = append(chunks, rec)
	}
	return chunks, nil
}

func (l *VectorDb) QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int) ([]db.ChunkQueryResult, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open query table failed %s: %w", tableName, err)
	}
	defer tbl.Close()

	rawResults, err := tbl.VectorSearch(ctx, "vector", vector, topK)
	if err != nil {
		return nil, fmt.Errorf("lancedb: query execution failed on %s: %w", tableName, err)
	}

	return mapResultsToChunks(rawResults)
}
