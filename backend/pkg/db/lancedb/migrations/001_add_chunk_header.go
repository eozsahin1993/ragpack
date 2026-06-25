package migrations

import (
	"context"
	"fmt"

	"github.com/eozsahin1993/lancedb-go/pkg/contracts"
)

func init() {
	Register(Migration{
		Version: 1,
		Up:      addChunkHeader,
	})
}

// addChunkHeader adds the chunk_header column to all collection tables.
// chunk_header stores the heading breadcrumb for structured documents (Markdown, HTML),
// e.g. "Introduction > Background", for use in RAG prompt formatting.
func addChunkHeader(ctx context.Context, tbl contracts.ITable) error {
	se, ok := tbl.(contracts.ITableSchemaEvolve)
	if !ok {
		return fmt.Errorf("table does not support schema evolution")
	}
	_, err := se.AddColumns(ctx, []contracts.NewColumnTransform{
		{Name: "chunk_header", Expression: "cast(NULL as string)"},
	})
	return err
}
