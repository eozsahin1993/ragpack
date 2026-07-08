package query

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/db"
	"ragpack/pkg/db/filter"
	"ragpack/pkg/meta"
)

// resolveFilter loads metadata fields for a collection, compiles the raw filter DSL,
// and returns both so callers can use fields for reconstruction and sqlFilter for the vector query.
// On error it writes the HTTP response and returns a non-nil error.
func (h *Handler) resolveFilter(c *fiber.Ctx, collectionID string, rawFilters json.RawMessage) (fields []meta.MetadataField, sqlFilter string, err error) {
	fields, err = h.meta.ListMetadataFields(c.Context(), collectionID)
	if err != nil {
		return nil, "", c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load metadata fields"})
	}
	sqlFilter, err = filter.Compile(rawFilters, buildFieldMap(fields))
	if err != nil {
		return nil, "", c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return fields, sqlFilter, nil
}

func buildFieldMap(fields []meta.MetadataField) filter.FieldMap {
	fieldMap := make(filter.FieldMap, len(fields))
	for _, f := range fields {
		fieldMap[f.Name] = f
	}
	return fieldMap
}

func reconstructMetadata(rec db.ChunkDbRecord, fields []meta.MetadataField) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		idx := field.Slot - 1
		switch field.Type {
		case "str":
			if rec.MetadataStr[idx] != nil {
				out[field.Name] = *rec.MetadataStr[idx]
			}
		case "num":
			if rec.MetadataNum[idx] != nil {
				out[field.Name] = *rec.MetadataNum[idx]
			}
		case "bool":
			if rec.MetadataBool[idx] != nil {
				out[field.Name] = *rec.MetadataBool[idx]
			}
		case "date":
			if rec.MetadataDate[idx] != nil {
				out[field.Name] = time.Unix(*rec.MetadataDate[idx], 0).UTC().Format(time.RFC3339)
			}
		case "arr":
			if rec.MetadataArr[idx] != nil {
				out[field.Name] = rec.MetadataArr[idx]
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
