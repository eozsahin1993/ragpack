package lancedb

import (
	"context"
	"fmt"

	"github.com/eozsahin1993/lancedb-go/pkg/contracts"
	sdk "github.com/eozsahin1993/lancedb-go/pkg/lancedb"

	"ragpack/pkg/db"
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

func (l *VectorDb) InsertRecord(ctx context.Context, tableName string, record db.ChunkDbRecord) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: unable to open target table %s: %w", tableName, err)
	}
	defer tbl.Close()

	arrowRecord, err := chunkToArrowRecord(record, len(record.Vector))
	if err != nil {
		return fmt.Errorf("lancedb: failed to build arrow record: %w", err)
	}
	defer arrowRecord.Release()

	if err := tbl.Add(ctx, arrowRecord, nil); err != nil {
		return fmt.Errorf("lancedb: record commit failed on table %s: %w", tableName, err)
	}
	return nil
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
