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
// Keep in sync with json tags on db.ChunkDbRecord.
const (
	colID         = "id"
	colDocumentID = "document_id"
	colChunkHash  = "chunk_hash"
	colChunkIndex = "chunk_index"
	colVector     = "vector"
	colCreatedAt  = "created_at"
	colUpdatedAt  = "updated_at"
	colMimeType   = "mime_type"
	colFileUri    = "file_uri"
	colSourceName = "source_name"
	colChunkText  = "chunk_text"
	colExternalId = "external_id"
	colExtraJSON  = "extra_json"
)

func chunkArrowSchema(vectorDim int) *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
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
		{Name: colExternalId, Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: colExtraJSON, Type: arrow.BinaryTypes.String, Nullable: true},
	}, nil)
}

func chunkToArrowRecord(r db.ChunkDbRecord, vectorDim int) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	schema := chunkArrowSchema(vectorDim)

	idB := array.NewStringBuilder(pool)
	idB.Append(r.ID)
	idArr := idB.NewArray()
	defer idArr.Release()

	docIDB := array.NewStringBuilder(pool)
	docIDB.Append(r.DocumentID)
	docIDArr := docIDB.NewArray()
	defer docIDArr.Release()

	hashB := array.NewStringBuilder(pool)
	hashB.Append(r.ChunkHash)
	hashArr := hashB.NewArray()
	defer hashArr.Release()

	idxB := array.NewInt32Builder(pool)
	idxB.Append(int32(r.ChunkIndex))
	idxArr := idxB.NewArray()
	defer idxArr.Release()

	vecB := array.NewFixedSizeListBuilder(pool, int32(vectorDim), arrow.PrimitiveTypes.Float32)
	vb := vecB.ValueBuilder().(*array.Float32Builder)
	vecB.Append(true)
	vb.AppendValues(r.Vector, nil)
	vecArr := vecB.NewArray()
	defer vecArr.Release()

	createdB := array.NewInt64Builder(pool)
	createdB.Append(r.CreatedAt.Unix())
	createdArr := createdB.NewArray()
	defer createdArr.Release()

	updatedB := array.NewInt64Builder(pool)
	updatedB.Append(r.UpdatedAt.Unix())
	updatedArr := updatedB.NewArray()
	defer updatedArr.Release()

	mimeB := array.NewStringBuilder(pool)
	mimeB.Append(r.MimeType)
	mimeArr := mimeB.NewArray()
	defer mimeArr.Release()

	fileUriB := array.NewStringBuilder(pool)
	fileUriB.Append(r.FileUri)
	fileUriArr := fileUriB.NewArray()
	defer fileUriArr.Release()

	srcNameB := array.NewStringBuilder(pool)
	srcNameB.Append(r.SourceName)
	srcNameArr := srcNameB.NewArray()
	defer srcNameArr.Release()

	chunkTextB := array.NewStringBuilder(pool)
	appendOptionalString(chunkTextB, r.ChunkText)
	chunkTextArr := chunkTextB.NewArray()
	defer chunkTextArr.Release()

	extIdB := array.NewStringBuilder(pool)
	appendOptionalString(extIdB, r.ExternalId)
	extIdArr := extIdB.NewArray()
	defer extIdArr.Release()

	extraB := array.NewStringBuilder(pool)
	appendOptionalString(extraB, r.ExtraJSON)
	extraArr := extraB.NewArray()
	defer extraArr.Release()

	return array.NewRecord(schema, []arrow.Array{
		idArr, docIDArr, hashArr, idxArr, vecArr, createdArr, updatedArr,
		mimeArr, fileUriArr, srcNameArr,
		chunkTextArr, extIdArr, extraArr,
	}, 1), nil
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
		results = append(results, db.ChunkQueryResult{
			ChunkDbRecord: rec,
			Distance:      distance,
			Similarity:    similarity,
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
	rec.ExternalId = extractOptionalString(row, colExternalId)
	rec.ExtraJSON = extractOptionalString(row, colExtraJSON)

	return rec, nil
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
