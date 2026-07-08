package documents

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/db"
	"ragpack/pkg/meta"
)

func (h *Handler) GetMetadata(c *fiber.Ctx) error {
	doc, err := h.meta.GetDocument(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}
	col, err := h.meta.GetCollectionByID(c.Context(), doc.CollectionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "collection not found"})
	}
	fields, err := h.meta.ListMetadataFields(c.Context(), col.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load metadata fields"})
	}
	if len(fields) == 0 {
		return c.JSON(fiber.Map{"metadata": fiber.Map{}})
	}
	chunks, err := h.vec.ListChunksByDocument(c.Context(), col.TableName, doc.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if len(chunks) == 0 {
		return c.JSON(fiber.Map{"metadata": fiber.Map{}})
	}
	return c.JSON(fiber.Map{"metadata": consistentMetadata(chunks, fields)})
}

// consistentMetadata returns only fields whose value is identical across every chunk.
// This guards against partially-written state being shown as the current value.
func consistentMetadata(chunks []db.ChunkDbRecord, fields []meta.MetadataField) map[string]interface{} {
	out := make(map[string]interface{})
	for _, field := range fields {
		idx := field.Slot - 1
		val, ok := consistentFieldValue(chunks, field.Type, idx)
		if ok {
			out[field.Name] = val
		}
	}
	return out
}

func consistentFieldValue(chunks []db.ChunkDbRecord, typ string, idx int) (interface{}, bool) {
	var ref string
	for i, ch := range chunks {
		var cur string
		switch typ {
		case "str":
			if ch.MetadataStr[idx] == nil {
				return nil, false
			}
			cur = *ch.MetadataStr[idx]
		case "num":
			if ch.MetadataNum[idx] == nil {
				return nil, false
			}
			cur = fmt.Sprintf("%g", *ch.MetadataNum[idx])
		case "bool":
			if ch.MetadataBool[idx] == nil {
				return nil, false
			}
			if *ch.MetadataBool[idx] {
				cur = "true"
			} else {
				cur = "false"
			}
		case "date":
			if ch.MetadataDate[idx] == nil {
				return nil, false
			}
			cur = time.Unix(*ch.MetadataDate[idx], 0).UTC().Format("2006-01-02")
		case "arr":
			if ch.MetadataArr[idx] == nil {
				return nil, false
			}
			cur = fmt.Sprintf("%v", ch.MetadataArr[idx])
		default:
			return nil, false
		}

		if i == 0 {
			ref = cur
		} else if cur != ref {
			return nil, false
		}
	}

	// Convert ref string back to the right type for the JSON response
	switch typ {
	case "num":
		var f float64
		fmt.Sscanf(ref, "%g", &f)
		return f, true
	case "bool":
		return ref == "true", true
	case "arr":
		return chunks[0].MetadataArr[idx], true
	default: // str, date — return as string (date is already YYYY-MM-DD)
		return ref, true
	}
}
