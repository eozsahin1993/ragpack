package query

import (
	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"

	"ragpack/pkg/telemetry/schema"
)

var (
	hybridType = arrow.StructOf(
		arrow.Field{Name: "full_text_weight", Type: arrow.PrimitiveTypes.Float64},
		arrow.Field{Name: "semantic_weight", Type: arrow.PrimitiveTypes.Float64},
		arrow.Field{Name: "rrf_k", Type: arrow.PrimitiveTypes.Float64},
		arrow.Field{Name: "full_text_limit", Type: arrow.PrimitiveTypes.Int32},
	)
	resultType = arrow.StructOf(
		arrow.Field{Name: "source_name", Type: arrow.BinaryTypes.String},
		arrow.Field{Name: "similarity", Type: arrow.PrimitiveTypes.Float64},
		arrow.Field{Name: "bm25_score", Type: arrow.PrimitiveTypes.Float64},
		arrow.Field{Name: "rrf_score", Type: arrow.PrimitiveTypes.Float64},
	)

	arrowSchema = arrow.NewSchema([]arrow.Field{
		{Name: "event_id", Type: arrow.BinaryTypes.String},
		{Name: "occurred_at", Type: schema.TsType},
		{Name: "collection_id", Type: arrow.BinaryTypes.String},
		{Name: "collection_slug", Type: arrow.BinaryTypes.String},
		{Name: "api_key_id", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "origin", Type: arrow.BinaryTypes.String},
		{Name: "endpoint", Type: arrow.BinaryTypes.String},
		{Name: "query_text", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "top_k", Type: arrow.PrimitiveTypes.Int32},
		{Name: "vector_search_only", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "hybrid_settings", Type: hybridType},
		{Name: "filters_json", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "embed_model", Type: arrow.BinaryTypes.String},
		{Name: "embed_query_tokens", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "embed_ms", Type: arrow.PrimitiveTypes.Int64},
		{Name: "vector_search_ms", Type: arrow.PrimitiveTypes.Int64},
		{Name: "result_count", Type: arrow.PrimitiveTypes.Int32},
		{Name: "results", Type: arrow.ListOf(resultType)},
		{Name: "status", Type: arrow.BinaryTypes.String},
		{Name: "error", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "total_ms", Type: arrow.PrimitiveTypes.Int64},
		{Name: "prompt_slug", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "llm_model", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "llm_input_tokens", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "llm_output_tokens", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "llm_cost_usd", Type: arrow.PrimitiveTypes.Float64, Nullable: true},
		{Name: "llm_ms", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
	}, nil)
)

// Table is this package's registration into the recorder — see
// pkg/telemetry/sink.go for how it's wired up generically.
var Table = schema.Table[*Event]{
	Name:   "query_events",
	Schema: arrowSchema,
	Build:  buildRecord,
}

func buildRecord(events []*Event) arrow.Record {
	b := array.NewRecordBuilder(memory.DefaultAllocator, arrowSchema)
	defer b.Release()
	for _, ev := range events {
		b.Field(0).(*array.StringBuilder).Append(ev.EventID)
		b.Field(1).(*array.TimestampBuilder).Append(schema.Ts(ev.OccurredAt))
		b.Field(2).(*array.StringBuilder).Append(ev.CollectionID)
		b.Field(3).(*array.StringBuilder).Append(ev.CollectionSlug)
		schema.OptStr(b.Field(4).(*array.StringBuilder), ev.APIKeyID)
		b.Field(5).(*array.StringBuilder).Append(ev.Origin)
		b.Field(6).(*array.StringBuilder).Append(ev.Endpoint)
		schema.OptStr(b.Field(7).(*array.StringBuilder), ev.QueryText)
		b.Field(8).(*array.Int32Builder).Append(int32(ev.TopK))
		b.Field(9).(*array.BooleanBuilder).Append(ev.VectorSearchOnly)

		hb := b.Field(10).(*array.StructBuilder)
		hb.Append(true)
		hb.FieldBuilder(0).(*array.Float64Builder).Append(ev.Hybrid.FullTextWeight)
		hb.FieldBuilder(1).(*array.Float64Builder).Append(ev.Hybrid.SemanticWeight)
		hb.FieldBuilder(2).(*array.Float64Builder).Append(ev.Hybrid.RRFK)
		hb.FieldBuilder(3).(*array.Int32Builder).Append(ev.Hybrid.FullTextLimit)

		schema.OptStr(b.Field(11).(*array.StringBuilder), ev.FiltersJSON)
		b.Field(12).(*array.StringBuilder).Append(ev.EmbedModel)
		schema.OptI64(b.Field(13).(*array.Int64Builder), ev.EmbedQueryTokens)
		b.Field(14).(*array.Int64Builder).Append(ev.EmbedMs)
		b.Field(15).(*array.Int64Builder).Append(ev.VectorSearchMs)
		b.Field(16).(*array.Int32Builder).Append(int32(ev.ResultCount))

		lb := b.Field(17).(*array.ListBuilder)
		lb.Append(true)
		rb := lb.ValueBuilder().(*array.StructBuilder)
		for _, r := range ev.Results {
			rb.Append(true)
			rb.FieldBuilder(0).(*array.StringBuilder).Append(r.SourceName)
			rb.FieldBuilder(1).(*array.Float64Builder).Append(r.Similarity)
			rb.FieldBuilder(2).(*array.Float64Builder).Append(r.BM25Score)
			rb.FieldBuilder(3).(*array.Float64Builder).Append(r.RRFScore)
		}

		b.Field(18).(*array.StringBuilder).Append(ev.Status)
		schema.OptStr(b.Field(19).(*array.StringBuilder), ev.Error)
		b.Field(20).(*array.Int64Builder).Append(ev.TotalMs)
		schema.OptStr(b.Field(21).(*array.StringBuilder), ev.PromptSlug)
		schema.OptStr(b.Field(22).(*array.StringBuilder), ev.LLMModel)
		schema.OptI64(b.Field(23).(*array.Int64Builder), ev.LLMInputTokens)
		schema.OptI64(b.Field(24).(*array.Int64Builder), ev.LLMOutputTokens)
		schema.OptF64(b.Field(25).(*array.Float64Builder), ev.LLMCostUSD)
		schema.OptI64(b.Field(26).(*array.Int64Builder), ev.LLMMs)
	}
	return b.NewRecord()
}
