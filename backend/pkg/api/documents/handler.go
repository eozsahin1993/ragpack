package documents

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/middleware"
	"ragpack/pkg/api/validate"
	"ragpack/pkg/db"
	"ragpack/pkg/ingester"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta       meta.MetaStore
	vec        db.VectorDb
	enforceACL bool
}

// chunksDefaultLimit is the page size for GET .../documents/:id/chunks when
// the caller doesn't specify one — matches the web-admin's chunk list page size.
const chunksDefaultLimit = 20

// NewHandler builds a documents handler. Mounted both under /collections/:slug
// and at the top level, so the access check happens here, not in router
// middleware. enforceACL is false only from RegisterAdmin, which has no Auth middleware.
func NewHandler(ms meta.MetaStore, vec db.VectorDb, enforceACL bool) *Handler {
	return &Handler{meta: ms, vec: vec, enforceACL: enforceACL}
}

func (h *Handler) checkAccess(c *fiber.Ctx, collectionID string, required meta.Permission) error {
	if !h.enforceACL {
		return nil
	}
	return middleware.CheckAccess(c, h.meta, collectionID, required)
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
		if err := h.checkAccess(c, col.ID, meta.PermissionRead); err != nil {
			return err
		}
		filter.CollectionID = &col.ID
	} else if h.enforceACL {
		// No collection to check grants against, so require wildcard rather than fail open.
		if err := middleware.RequireWildcard(h.meta, meta.PermissionRead)(c); err != nil {
			return err
		}
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
	if err := h.checkAccess(c, doc.CollectionID, meta.PermissionRead); err != nil {
		return err
	}
	return c.JSON(doc)
}

func (h *Handler) Chunks(c *fiber.Ctx) error {
	doc, err := h.meta.GetDocument(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}
	if err := h.checkAccess(c, doc.CollectionID, meta.PermissionRead); err != nil {
		return err
	}

	col, err := h.meta.GetCollectionByID(c.Context(), doc.CollectionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "collection not found"})
	}

	limit, offset := validate.PaginationWithDefault(c, chunksDefaultLimit)

	chunks, total, err := h.vec.ListChunksByDocumentPage(c.Context(), col.TableName, doc.ID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"chunks": chunks, "total": total, "limit": limit, "offset": offset})
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

	existing, err := h.meta.GetDocument(c.Context(), docID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}
	if err := h.checkAccess(c, existing.CollectionID, meta.PermissionWrite); err != nil {
		return err
	}

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
	if err := h.checkAccess(c, doc.CollectionID, meta.PermissionWrite); err != nil {
		return err
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
