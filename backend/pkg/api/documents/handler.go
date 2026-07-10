package documents

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/db"
	"ragpack/pkg/ingester"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
	vec  db.VectorDb
}

func NewHandler(ms meta.MetaStore, vec db.VectorDb) *Handler {
	return &Handler{meta: ms, vec: vec}
}

func (h *Handler) List(c *fiber.Ctx) error {
	var q ListQuery
	if err := validate.Query(c, &q); err != nil {
		return err
	}

	filter := meta.DocumentFilter{}
	if slug := c.Params("slug"); slug != "" {
		col, err := h.meta.GetCollectionBySlug(c.Context(), slug)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
		}
		filter.CollectionID = &col.ID
	}
	if q.Status != "" {
		status := meta.DocumentStatus(q.Status)
		filter.Status = &status
	}

	sort := meta.DocumentSort{
		Field: q.SortBy,
		Dir:   meta.SortDir(q.SortDir),
	}

	limit, offset := validate.Pagination(c)

	docs, err := h.meta.ListDocuments(c.Context(), filter, sort, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	total, err := h.meta.CountDocuments(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"documents": docs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	doc, err := h.meta.GetDocument(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}
	return c.JSON(doc)
}

func (h *Handler) Chunks(c *fiber.Ctx) error {
	doc, err := h.meta.GetDocument(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}

	col, err := h.meta.GetCollectionByID(c.Context(), doc.CollectionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "collection not found"})
	}

	chunks, err := h.vec.ListChunksByDocument(c.Context(), col.TableName, doc.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"chunks": chunks, "total": len(chunks)})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	docID := c.Params("id")

	if err := h.meta.UpdateDocument(c.Context(), docID, meta.DocumentPatch{
		Name:      req.Name,
		ExtraJSON: req.ExtraJSON,
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if req.ExtraJSON != nil || req.Metadata != nil {
		doc, err := h.meta.GetDocument(c.Context(), docID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
		}
		col, err := h.meta.GetCollectionByID(c.Context(), doc.CollectionID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "collection not found"})
		}

		patch := db.ChunkPatch{ExtraJSON: req.ExtraJSON}

		if req.Metadata != nil {
			fields, err := h.meta.ListMetadataFields(c.Context(), col.ID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load metadata fields"})
			}
			patch = ingester.MergeMetadataSlots(patch, req.Metadata, fields)
		}

		if err := h.vec.UpdateChunks(c.Context(), col.TableName, docID, patch); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	doc, err := h.meta.GetDocument(c.Context(), docID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(doc)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	doc, err := h.meta.GetDocument(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}

	col, err := h.meta.GetCollectionByID(c.Context(), doc.CollectionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "collection not found"})
	}

	if doc.ChunkCount > 0 {
		if err := h.vec.DeleteChunksByDocument(c.Context(), col.TableName, doc.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	if err := h.meta.DeleteDocument(c.Context(), doc.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
