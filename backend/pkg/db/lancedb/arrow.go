package lancedb

import (
	"fmt"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"

	"ragpack/pkg/db"
)

// Column names for the ChunkDbRecord Arrow schema.
const (
	colID          = "id"
	colDocumentID  = "document_id"
	colChunkHash   = "chunk_hash"
	colChunkIndex  = "chunk_index"
	colVector      = "vector"
	colCreatedAt   = "created_at"
	colUpdatedAt   = "updated_at"
	colMimeType    = "mime_type"
	colFileUri     = "file_uri"
	colSourceName  = "source_name"
	colChunkText   = "chunk_text"
	colChunkHeader = "chunk_header"
	colExternalId  = "external_id"
	colExtraJSON   = "extra_json"
)

const (
	metaStrSlots  = 20
	metaNumSlots  = 10
	metaBoolSlots = 10
	metaDateSlots = 10
	metaArrSlots  = 10
)

func chunkArrowSchema(vectorDim int) *arrow.Schema {
	fields := []arrow.Field{
		{Name: colID, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: colDocumentID, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: colChunkHash, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: colChunkIndex, Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: colVector, Type: arrow.FixedSizeListOf(int32(vectorDim), arrow.PrimitiveTypes.Float32), Nullable: false},
		{Name: colCreatedAt, Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: colUpdatedAt, Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: colMimeType, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: colFileUri, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: colSourceName, Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: colChunkText, Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: colChunkHeader, Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: colExternalId, Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: colExtraJSON, Type: arrow.BinaryTypes.String, Nullable: true},
	}
	for i := 1; i <= metaStrSlots; i++ {
		fields = append(fields, arrow.Field{Name: db.MetadataSlotColumn("str", i), Type: arrow.BinaryTypes.String, Nullable: true})
	}
	for i := 1; i <= metaNumSlots; i++ {
		fields = append(fields, arrow.Field{Name: db.MetadataSlotColumn("num", i), Type: arrow.PrimitiveTypes.Float64, Nullable: true})
	}
	for i := 1; i <= metaBoolSlots; i++ {
		fields = append(fields, arrow.Field{Name: db.MetadataSlotColumn("bool", i), Type: arrow.FixedWidthTypes.Boolean, Nullable: true})
	}
	for i := 1; i <= metaDateSlots; i++ {
		fields = append(fields, arrow.Field{Name: db.MetadataSlotColumn("date", i), Type: arrow.PrimitiveTypes.Int64, Nullable: true})
	}
	for i := 1; i <= metaArrSlots; i++ {
		fields = append(fields, arrow.Field{Name: db.MetadataSlotColumn("arr", i), Type: arrow.ListOf(arrow.BinaryTypes.String), Nullable: true})
	}
	return arrow.NewSchema(fields, nil)
}

type chunkBuilders struct {
	id, docID, hash, mime, fileUri, srcName, chunkText, chunkHeader, extID, extra *array.StringBuilder
	idx                                                                           *array.Int32Builder
	vec                                                                           *array.FixedSizeListBuilder
	created, updated                                                              *array.Int64Builder
	metaStr                                                                       [metaStrSlots]*array.StringBuilder
	metaNum                                                                       [metaNumSlots]*array.Float64Builder
	metaBool                                                                      [metaBoolSlots]*array.BooleanBuilder
	metaDate                                                                      [metaDateSlots]*array.Int64Builder
	metaArr                                                                       [metaArrSlots]*array.ListBuilder
}

func newChunkBuilders(pool memory.Allocator, vectorDim int) chunkBuilders {
	vecB := array.NewFixedSizeListBuilder(pool, int32(vectorDim), arrow.PrimitiveTypes.Float32)
	b := chunkBuilders{
		id: array.NewStringBuilder(pool), docID: array.NewStringBuilder(pool),
		hash: array.NewStringBuilder(pool), mime: array.NewStringBuilder(pool),
		fileUri: array.NewStringBuilder(pool), srcName: array.NewStringBuilder(pool),
		chunkText: array.NewStringBuilder(pool), chunkHeader: array.NewStringBuilder(pool),
		extID: array.NewStringBuilder(pool), extra: array.NewStringBuilder(pool),
		idx: array.NewInt32Builder(pool),
		vec: vecB, created: array.NewInt64Builder(pool), updated: array.NewInt64Builder(pool),
	}
	for i := 0; i < metaStrSlots; i++ {
		b.metaStr[i] = array.NewStringBuilder(pool)
	}
	for i := 0; i < metaNumSlots; i++ {
		b.metaNum[i] = array.NewFloat64Builder(pool)
	}
	for i := 0; i < metaBoolSlots; i++ {
		b.metaBool[i] = array.NewBooleanBuilder(pool)
	}
	for i := 0; i < metaDateSlots; i++ {
		b.metaDate[i] = array.NewInt64Builder(pool)
	}
	for i := 0; i < metaArrSlots; i++ {
		b.metaArr[i] = array.NewListBuilder(pool, arrow.BinaryTypes.String)
	}
	return b
}

func (b chunkBuilders) append(r db.ChunkDbRecord) {
	b.id.Append(r.ID)
	b.docID.Append(r.DocumentID)
	b.hash.Append(r.ChunkHash)
	b.idx.Append(int32(r.ChunkIndex))
	b.vec.Append(true)
	b.vec.ValueBuilder().(*array.Float32Builder).AppendValues(r.Vector, nil)
	b.created.Append(r.CreatedAt.Unix())
	b.updated.Append(r.UpdatedAt.Unix())
	b.mime.Append(r.MimeType)
	b.fileUri.Append(r.FileUri)
	b.srcName.Append(r.SourceName)
	appendOptionalString(b.chunkText, r.ChunkText)
	appendOptionalString(b.chunkHeader, r.ChunkHeader)
	appendOptionalString(b.extID, r.ExternalId)
	appendOptionalString(b.extra, r.ExtraJSON)

	for i := 0; i < metaStrSlots; i++ {
		appendOptionalString(b.metaStr[i], r.MetadataStr[i])
	}
	for i := 0; i < metaNumSlots; i++ {
		if r.MetadataNum[i] == nil {
			b.metaNum[i].AppendNull()
		} else {
			b.metaNum[i].Append(*r.MetadataNum[i])
		}
	}
	for i := 0; i < metaBoolSlots; i++ {
		if r.MetadataBool[i] == nil {
			b.metaBool[i].AppendNull()
		} else {
			b.metaBool[i].Append(*r.MetadataBool[i])
		}
	}
	for i := 0; i < metaDateSlots; i++ {
		if r.MetadataDate[i] == nil {
			b.metaDate[i].AppendNull()
		} else {
			b.metaDate[i].Append(*r.MetadataDate[i])
		}
	}
	for i := 0; i < metaArrSlots; i++ {
		if r.MetadataArr[i] == nil {
			b.metaArr[i].AppendNull()
		} else {
			b.metaArr[i].Append(true)
			sb := b.metaArr[i].ValueBuilder().(*array.StringBuilder)
			for _, s := range r.MetadataArr[i] {
				sb.Append(s)
			}
		}
	}
}

func (b chunkBuilders) finish(schema *arrow.Schema, n int64) arrow.Record {
	cols := []arrow.Array{
		b.id.NewArray(), b.docID.NewArray(), b.hash.NewArray(), b.idx.NewArray(),
		b.vec.NewArray(), b.created.NewArray(), b.updated.NewArray(),
		b.mime.NewArray(), b.fileUri.NewArray(), b.srcName.NewArray(),
		b.chunkText.NewArray(), b.chunkHeader.NewArray(), b.extID.NewArray(), b.extra.NewArray(),
	}
	for i := 0; i < metaStrSlots; i++ {
		cols = append(cols, b.metaStr[i].NewArray())
	}
	for i := 0; i < metaNumSlots; i++ {
		cols = append(cols, b.metaNum[i].NewArray())
	}
	for i := 0; i < metaBoolSlots; i++ {
		cols = append(cols, b.metaBool[i].NewArray())
	}
	for i := 0; i < metaDateSlots; i++ {
		cols = append(cols, b.metaDate[i].NewArray())
	}
	for i := 0; i < metaArrSlots; i++ {
		cols = append(cols, b.metaArr[i].NewArray())
	}
	for _, c := range cols {
		defer c.Release()
	}
	return array.NewRecord(schema, cols, n)
}

func chunksToArrowRecord(records []db.ChunkDbRecord, vectorDim int) (arrow.Record, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("chunksToArrowRecord: empty batch")
	}
	pool := memory.NewGoAllocator()
	b := newChunkBuilders(pool, vectorDim)
	for _, r := range records {
		b.append(r)
	}
	return b.finish(chunkArrowSchema(vectorDim), int64(len(records))), nil
}

func chunkToArrowRecord(r db.ChunkDbRecord, vectorDim int) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	b := newChunkBuilders(pool, vectorDim)
	b.append(r)
	return b.finish(chunkArrowSchema(vectorDim), 1), nil
}

func appendOptionalString(b *array.StringBuilder, v *string) {
	if v == nil {
		b.AppendNull()
	} else {
		b.Append(*v)
	}
}

func mapResultsToChunks(rows []map[string]interface{}) ([]db.ChunkQueryResult, error) {
	results := make([]db.ChunkQueryResult, 0, len(rows))
	for i, row := range rows {
		rec, err := rowToChunk(row)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}
		var distance float32
		if d, ok := row["_distance"]; ok {
			if f, ok := d.(float64); ok {
				distance = float32(f)
			}
		}
		// L2 distance with unit vectors: cosine_sim = 1 - d²/2, clipped to [0, 1]
		cosineSim := float32(1) - (distance*distance)/2
		if cosineSim < 0 {
			cosineSim = 0
		}
		similarity := cosineSim * 100

		// Raw BM25 score from the FTS pass; see ChunkQueryResult.KeywordBM25Score.
		var keywordScore float32
		if s, ok := row["_score"]; ok {
			if f, ok := s.(float64); ok {
				keywordScore = float32(f)
			}
		}

		results = append(results, db.ChunkQueryResult{
			ChunkDbRecord:    rec,
			VectorDistance:   distance,
			VectorSimilarity: similarity,
			KeywordBM25Score: keywordScore,
		})
	}
	return results, nil
}

func rowToChunk(row map[string]interface{}) (db.ChunkDbRecord, error) {
	var rec db.ChunkDbRecord
	var err error

	if rec.ID, err = extractString(row, colID); err != nil {
		return rec, err
	}
	if rec.DocumentID, err = extractString(row, colDocumentID); err != nil {
		return rec, err
	}
	if rec.ChunkHash, err = extractString(row, colChunkHash); err != nil {
		return rec, err
	}
	chunkIndex, err := extractInt32(row, colChunkIndex)
	if err != nil {
		return rec, err
	}
	rec.ChunkIndex = int(chunkIndex)

	if rec.Vector, err = extractFloat32Slice(row, colVector); err != nil {
		return rec, err
	}

	createdAt, err := extractInt64(row, colCreatedAt)
	if err != nil {
		return rec, err
	}
	rec.CreatedAt = time.Unix(createdAt, 0)

	updatedAt, err := extractInt64(row, colUpdatedAt)
	if err != nil {
		return rec, err
	}
	rec.UpdatedAt = time.Unix(updatedAt, 0)

	if rec.MimeType, err = extractString(row, colMimeType); err != nil {
		return rec, err
	}
	if rec.FileUri, err = extractString(row, colFileUri); err != nil {
		return rec, err
	}
	if rec.SourceName, err = extractString(row, colSourceName); err != nil {
		return rec, err
	}

	rec.ChunkText = extractOptionalString(row, colChunkText)
	rec.ChunkHeader = extractOptionalString(row, colChunkHeader)
	rec.ExternalId = extractOptionalString(row, colExternalId)
	rec.ExtraJSON = extractOptionalString(row, colExtraJSON)

	for i := 0; i < metaStrSlots; i++ {
		rec.MetadataStr[i] = extractOptionalString(row, db.MetadataSlotColumn("str", i+1))
	}
	for i := 0; i < metaNumSlots; i++ {
		rec.MetadataNum[i] = extractOptionalFloat64(row, db.MetadataSlotColumn("num", i+1))
	}
	for i := 0; i < metaBoolSlots; i++ {
		rec.MetadataBool[i] = extractOptionalBool(row, db.MetadataSlotColumn("bool", i+1))
	}
	for i := 0; i < metaDateSlots; i++ {
		rec.MetadataDate[i] = extractOptionalInt64(row, db.MetadataSlotColumn("date", i+1))
	}
	for i := 0; i < metaArrSlots; i++ {
		rec.MetadataArr[i] = extractOptionalStringSlice(row, db.MetadataSlotColumn("arr", i+1))
	}

	return rec, nil
}

// metadataSlotColumns lists every metadata slot column name, for queries
// that only need metadata values (e.g. the document metadata consistency
// check) rather than full chunk records.
func metadataSlotColumns() []string {
	cols := make([]string, 0, metaStrSlots+metaNumSlots+metaBoolSlots+metaDateSlots+metaArrSlots)
	for i := 1; i <= metaStrSlots; i++ {
		cols = append(cols, db.MetadataSlotColumn("str", i))
	}
	for i := 1; i <= metaNumSlots; i++ {
		cols = append(cols, db.MetadataSlotColumn("num", i))
	}
	for i := 1; i <= metaBoolSlots; i++ {
		cols = append(cols, db.MetadataSlotColumn("bool", i))
	}
	for i := 1; i <= metaDateSlots; i++ {
		cols = append(cols, db.MetadataSlotColumn("date", i))
	}
	for i := 1; i <= metaArrSlots; i++ {
		cols = append(cols, db.MetadataSlotColumn("arr", i))
	}
	return cols
}

// rowToChunkMetadataOnly decodes a row containing only metadata slot columns
// (see metadataSlotColumns). Unlike rowToChunk, it never errors on missing
// fields — required chunk fields are intentionally absent from these rows.
func rowToChunkMetadataOnly(row map[string]interface{}) db.ChunkDbRecord {
	var rec db.ChunkDbRecord
	for i := 0; i < metaStrSlots; i++ {
		rec.MetadataStr[i] = extractOptionalString(row, db.MetadataSlotColumn("str", i+1))
	}
	for i := 0; i < metaNumSlots; i++ {
		rec.MetadataNum[i] = extractOptionalFloat64(row, db.MetadataSlotColumn("num", i+1))
	}
	for i := 0; i < metaBoolSlots; i++ {
		rec.MetadataBool[i] = extractOptionalBool(row, db.MetadataSlotColumn("bool", i+1))
	}
	for i := 0; i < metaDateSlots; i++ {
		rec.MetadataDate[i] = extractOptionalInt64(row, db.MetadataSlotColumn("date", i+1))
	}
	for i := 0; i < metaArrSlots; i++ {
		rec.MetadataArr[i] = extractOptionalStringSlice(row, db.MetadataSlotColumn("arr", i+1))
	}
	return rec
}

func extractString(row map[string]interface{}, key string) (string, error) {
	v, ok := row[key]
	if !ok {
		return "", fmt.Errorf("missing field %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("field %q: expected string, got %T", key, v)
	}
	return s, nil
}

func extractOptionalString(row map[string]interface{}, key string) *string {
	v, ok := row[key]
	if !ok || v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

func extractOptionalFloat64(row map[string]interface{}, key string) *float64 {
	v, ok := row[key]
	if !ok || v == nil {
		return nil
	}
	f, ok := v.(float64)
	if !ok {
		return nil
	}
	return &f
}

func extractOptionalBool(row map[string]interface{}, key string) *bool {
	v, ok := row[key]
	if !ok || v == nil {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

func extractOptionalInt64(row map[string]interface{}, key string) *int64 {
	v, ok := row[key]
	if !ok || v == nil {
		return nil
	}
	switch n := v.(type) {
	case int64:
		return &n
	case float64:
		i := int64(n)
		return &i
	default:
		return nil
	}
}

func extractOptionalStringSlice(row map[string]interface{}, key string) []string {
	v, ok := row[key]
	if !ok || v == nil {
		return nil
	}
	switch items := v.(type) {
	case []string:
		return items
	case []interface{}:
		out := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func extractInt32(row map[string]interface{}, key string) (int32, error) {
	v, ok := row[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}
	switch n := v.(type) {
	case int32:
		return n, nil
	case float64:
		return int32(n), nil
	default:
		return 0, fmt.Errorf("field %q: expected int32, got %T", key, v)
	}
}

func extractInt64(row map[string]interface{}, key string) (int64, error) {
	v, ok := row[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}
	switch n := v.(type) {
	case int64:
		return n, nil
	case float64:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("field %q: expected int64, got %T", key, v)
	}
}

func extractFloat32Slice(row map[string]interface{}, key string) ([]float32, error) {
	v, ok := row[key]
	if !ok {
		return nil, fmt.Errorf("missing field %q", key)
	}
	switch f := v.(type) {
	case []float32:
		return f, nil
	case []interface{}:
		out := make([]float32, len(f))
		for i, elem := range f {
			n, ok := elem.(float64)
			if !ok {
				return nil, fmt.Errorf("field %q: element %d expected float64, got %T", key, i, elem)
			}
			out[i] = float32(n)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("field %q: expected []float32, got %T", key, v)
	}
}
