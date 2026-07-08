package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/db"
	"ragpack/pkg/db/filter"
	"ragpack/pkg/meta"
)

type filterErr struct {
	status int
	msg    string
}

func (e *filterErr) Error() string { return e.msg }

// resolveFilter loads metadata fields and compiles the raw filter DSL.
// Returns a *filterErr on failure — callers should use errors.As to get the HTTP status.
func (h *Handler) resolveFilter(ctx context.Context, collectionID string, rawFilters json.RawMessage) (fields []meta.MetadataField, sqlFilter string, err error) {
	fields, err = h.meta.ListMetadataFields(ctx, collectionID)
	if err != nil {
		return nil, "", &filterErr{fiber.StatusInternalServerError, "failed to load metadata fields"}
	}
	sqlFilter, err = filter.Compile(rawFilters, buildFieldMap(fields))
	if err != nil {
		return nil, "", &filterErr{fiber.StatusBadRequest, err.Error()}
	}
	return fields, sqlFilter, nil
}

func writeFilterErr(c *fiber.Ctx, err error) error {
	var fe *filterErr
	if errors.As(err, &fe) {
		return c.Status(fe.status).JSON(fiber.Map{"error": fe.msg})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("filter: %v", err)})
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
