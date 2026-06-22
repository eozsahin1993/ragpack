package db

import (
	"context"
	"fmt"
	"github.com/lancedb/lancedb-go/pkg/lancedb" 
)

type LanceVectorDb struct {
	conn lancedb.Connection
}

func NewLanceVectorDb() *LanceVectorDb {
	return &LanceVectorDb{}
}

func (l *LanceVectorDb) Connect(ctx context.Context, connectionUrl string) error {
	conn, err := lancedb.Connect(connectionUrl)
	if err != nil {
		return fmt.Errorf("lancedb: connectivity failure: %w", err)
	}
	l.conn = conn
	return nil
}

func (l *LanceVectorDb) CreateTable(ctx context.Context, name string) error {
	schemaBlueprint := []ChunkDbRecord{}
	
	_, err := l.conn.CreateTable(ctx, name, schemaBlueprint, nil)
	if err != nil {
		return fmt.Errorf("lancedb: table provisioning failed for %s: %w", name, err)
	}
	return nil
}

func (l *LanceVectorDb) InsertRecord(ctx context.Context, tableName string, record ChunkDbRecord) error {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("lancedb: unable to open target table %s: %w", tableName, err)
	}

	recordsBatch := []ChunkDbRecord{record}
	if err := tbl.Add(ctx, recordsBatch); err != nil {
		return fmt.Errorf("lancedb: record commit failed on table %s: %w", tableName, err)
	}
	return nil
}

func (l *LanceVectorDb) QuerySimilarVectors(ctx context.Context, tableName string, vector []float32, topK int) ([]ChunkDbRecord, error) {
	tbl, err := l.conn.OpenTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("lancedb: open query table failed %s: %w", tableName, err)
	}

	queryPipeline := tbl.Search(vector).Limit(topK)

	rawResults, err := queryPipeline.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("lancedb: query execution failed on %s: %w", tableName, err)
	}

	var matchedRecords []ChunkDbRecord
	if err := rawResults.Scan(&matchedRecords); err != nil {
		return nil, fmt.Errorf("lancedb: struct mapping extraction failure: %w", err)
	}

	return matchedRecords, nil
}
