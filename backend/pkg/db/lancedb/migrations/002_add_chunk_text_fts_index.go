package migrations

import (
	"context"

	"github.com/eozsahin1993/lancedb-go/pkg/contracts"
)

func init() {
	Register(Migration{
		Version: 2,
		Up:      addChunkTextFtsIndex,
	})
}

// addChunkTextFtsIndex creates a full-text-search index on chunk_text for
// collections created before hybrid search existed. New collections get
// this index directly in CreateTable; this backfills the rest.
func addChunkTextFtsIndex(ctx context.Context, tbl contracts.ITable) error {
	return tbl.CreateIndexWithName(ctx, []string{"chunk_text"}, contracts.IndexTypeFts, "chunk_text")
}
