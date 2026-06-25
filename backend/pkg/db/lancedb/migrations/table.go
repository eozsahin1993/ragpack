package migrations

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/eozsahin1993/lancedb-go/pkg/contracts"
	sdk "github.com/eozsahin1993/lancedb-go/pkg/lancedb"
)

const migrationsTable = "_ragpack_migrations"

func openOrCreateMigrationsTable(ctx context.Context, conn contracts.IConnection) (contracts.ITable, error) {
	tbl, err := conn.OpenTable(ctx, migrationsTable)
	if err == nil {
		return tbl, nil
	}
	schema, err := sdk.NewSchema(migrationsArrowSchema())
	if err != nil {
		return nil, fmt.Errorf("build migrations schema: %w", err)
	}
	tbl, err = conn.CreateTable(ctx, migrationsTable, schema)
	if err != nil {
		return nil, fmt.Errorf("create migrations table: %w", err)
	}
	return tbl, nil
}

func hasApplied(ctx context.Context, mTbl contracts.ITable, collectionID string, version int) (bool, error) {
	rows, err := mTbl.SelectWithFilter(ctx,
		fmt.Sprintf("collection_id = '%s' AND version = %d", collectionID, version))
	if err != nil {
		return false, fmt.Errorf("lancedb migrate: check %s/%d: %w", collectionID, version, err)
	}
	return len(rows) > 0, nil
}

func recordApplied(ctx context.Context, mTbl contracts.ITable, collectionID string, version int) error {
	pool := memory.NewGoAllocator()
	colIDBuilder := array.NewStringBuilder(pool)
	versionBuilder := array.NewInt32Builder(pool)
	colIDBuilder.Append(collectionID)
	versionBuilder.Append(int32(version))

	cols := []arrow.Array{colIDBuilder.NewArray(), versionBuilder.NewArray()}
	for _, c := range cols {
		defer c.Release()
	}
	rec := array.NewRecord(migrationsArrowSchema(), cols, 1)
	defer rec.Release()

	if err := mTbl.Add(ctx, rec, nil); err != nil {
		return fmt.Errorf("lancedb migrate: record %s/%d: %w", collectionID, version, err)
	}
	return nil
}

func migrationsArrowSchema() *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
		{Name: "collection_id", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "version", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	}, nil)
}
