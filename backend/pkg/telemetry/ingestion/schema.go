package ingestion

import (
	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"

	"ragpack/pkg/telemetry/schema"
)

var arrowSchema = arrow.NewSchema([]arrow.Field{
	{Name: "event_id", Type: arrow.BinaryTypes.String},
	{Name: "occurred_at", Type: schema.TsType},
	{Name: "job_id", Type: arrow.BinaryTypes.String},
	{Name: "document_id", Type: arrow.BinaryTypes.String},
	{Name: "collection_id", Type: arrow.BinaryTypes.String},
	{Name: "collection_slug", Type: arrow.BinaryTypes.String},
	{Name: "file_uri", Type: arrow.BinaryTypes.String},
	{Name: "mime_type", Type: arrow.BinaryTypes.String},
	{Name: "intent", Type: arrow.BinaryTypes.String},
	{Name: "status", Type: arrow.BinaryTypes.String},
	{Name: "error", Type: arrow.BinaryTypes.String, Nullable: true},
	{Name: "chunk_count", Type: arrow.PrimitiveTypes.Int32},
	{Name: "fetch_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "parse_chunk_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "rate_limit_wait_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "embed_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "insert_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "optimize_index_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "total_ms", Type: arrow.PrimitiveTypes.Int64},
	{Name: "embed_model", Type: arrow.BinaryTypes.String},
	{Name: "embed_tokens", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
	{Name: "embed_cost_usd", Type: arrow.PrimitiveTypes.Float64, Nullable: true},
}, nil)

// Table is this package's registration into the recorder — see
// pkg/telemetry/sink.go for how it's wired up generically.
var Table = schema.Table[*Event]{
	Name:   "ingestion_events",
	Schema: arrowSchema,
	Build:  buildRecord,
}

func buildRecord(events []*Event) arrow.Record {
	b := array.NewRecordBuilder(memory.DefaultAllocator, arrowSchema)
	defer b.Release()
	for _, ev := range events {
		b.Field(0).(*array.StringBuilder).Append(ev.EventID)
		b.Field(1).(*array.TimestampBuilder).Append(schema.Ts(ev.OccurredAt))
		b.Field(2).(*array.StringBuilder).Append(ev.JobID)
		b.Field(3).(*array.StringBuilder).Append(ev.DocumentID)
		b.Field(4).(*array.StringBuilder).Append(ev.CollectionID)
		b.Field(5).(*array.StringBuilder).Append(ev.CollectionSlug)
		b.Field(6).(*array.StringBuilder).Append(ev.FileUri)
		b.Field(7).(*array.StringBuilder).Append(ev.MimeType)
		b.Field(8).(*array.StringBuilder).Append(ev.Intent)
		b.Field(9).(*array.StringBuilder).Append(ev.Status)
		schema.OptStr(b.Field(10).(*array.StringBuilder), ev.Error)
		b.Field(11).(*array.Int32Builder).Append(int32(ev.ChunkCount))
		b.Field(12).(*array.Int64Builder).Append(ev.FetchMs)
		b.Field(13).(*array.Int64Builder).Append(ev.ParseChunkMs)
		b.Field(14).(*array.Int64Builder).Append(ev.RateLimitWaitMs)
		b.Field(15).(*array.Int64Builder).Append(ev.EmbedMs)
		b.Field(16).(*array.Int64Builder).Append(ev.InsertMs)
		b.Field(17).(*array.Int64Builder).Append(ev.OptimizeIndexMs)
		b.Field(18).(*array.Int64Builder).Append(ev.TotalMs)
		b.Field(19).(*array.StringBuilder).Append(ev.EmbedModel)
		schema.OptI64(b.Field(20).(*array.Int64Builder), ev.EmbedTokens)
		schema.OptF64(b.Field(21).(*array.Float64Builder), ev.EmbedCostUSD)
	}
	return b.NewRecord()
}
