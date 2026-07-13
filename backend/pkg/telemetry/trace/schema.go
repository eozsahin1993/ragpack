package trace

import (
	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"

	"ragpack/pkg/telemetry/schema"
)

var (
	chunkType = arrow.StructOf(
		arrow.Field{Name: "source_name", Type: arrow.BinaryTypes.String},
		arrow.Field{Name: "chunk_header", Type: arrow.BinaryTypes.String, Nullable: true},
		arrow.Field{Name: "chunk_text", Type: arrow.BinaryTypes.String},
		arrow.Field{Name: "similarity", Type: arrow.PrimitiveTypes.Float64},
		arrow.Field{Name: "bm25_score", Type: arrow.PrimitiveTypes.Float64},
	)

	arrowSchema = arrow.NewSchema([]arrow.Field{
		{Name: "event_id", Type: arrow.BinaryTypes.String},
		{Name: "occurred_at", Type: schema.TsType},
		{Name: "chunks", Type: arrow.ListOf(chunkType)},
		{Name: "formatted_prompt", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "answer", Type: arrow.BinaryTypes.String, Nullable: true},
	}, nil)
)

// Table is this package's registration into the recorder — see
// pkg/telemetry/sink.go for how it's wired up generically.
var Table = schema.Table[*Event]{
	Name:   "query_traces",
	Schema: arrowSchema,
	Build:  buildRecord,
}

func buildRecord(traces []*Event) arrow.Record {
	b := array.NewRecordBuilder(memory.DefaultAllocator, arrowSchema)
	defer b.Release()
	for _, tr := range traces {
		b.Field(0).(*array.StringBuilder).Append(tr.EventID)
		b.Field(1).(*array.TimestampBuilder).Append(schema.Ts(tr.OccurredAt))

		lb := b.Field(2).(*array.ListBuilder)
		lb.Append(true)
		cb := lb.ValueBuilder().(*array.StructBuilder)
		for _, c := range tr.Chunks {
			cb.Append(true)
			cb.FieldBuilder(0).(*array.StringBuilder).Append(c.SourceName)
			schema.OptStr(cb.FieldBuilder(1).(*array.StringBuilder), c.ChunkHeader)
			cb.FieldBuilder(2).(*array.StringBuilder).Append(c.ChunkText)
			cb.FieldBuilder(3).(*array.Float64Builder).Append(c.Similarity)
			cb.FieldBuilder(4).(*array.Float64Builder).Append(c.BM25Score)
		}

		schema.OptStr(b.Field(3).(*array.StringBuilder), tr.FormattedPrompt)
		schema.OptStr(b.Field(4).(*array.StringBuilder), tr.Answer)
	}
	return b.NewRecord()
}
