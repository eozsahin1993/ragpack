package collections

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta              meta.MetaStore
	vec               db.VectorDb
	registry          *embedder.Registry
	enforceACL        bool
	minRefreshSeconds int
}

// enforceACL is false only from RegisterAdmin (no Auth middleware); also gates internal response fields like table_name.
// minRefreshSeconds is an explicit dependency (not global validator state) so it can vary per-app, e.g. per test.
func NewHandler(ms meta.MetaStore, vec db.VectorDb, registry *embedder.Registry, enforceACL bool, minRefreshSeconds int) *Handler {
	return &Handler{meta: ms, vec: vec, registry: registry, enforceACL: enforceACL, minRefreshSeconds: minRefreshSeconds}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	model := req.EmbedModel
	if model == "" {
		var err error
		model, _, err = h.registry.Default()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	emb, err := h.registry.Get(model)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	dim, err := emb.Dimensions()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": fmt.Sprintf("embedding service unavailable: %v", err)})
	}

	input := meta.CreateCollectionInput{
		Name:       req.Name,
		EmbedModel: model,
		VectorDim:  dim,
	}
	if req.ChunkConfig != nil {
		input.ChunkStrategy = req.ChunkConfig.Strategy
		input.ChunkSize = req.ChunkConfig.Size
		input.ChunkOverlap = req.ChunkConfig.Overlap
	}
	col, err := h.meta.CreateCollection(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.vec.CreateTable(c.Context(), col.TableName, col.ID, col.VectorDim); err != nil {
		_ = h.meta.DeleteCollection(c.Context(), col.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(toResponse(col, !h.enforceACL))
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, offset := validate.Pagination(c)

	cols, err := h.meta.ListCollections(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	total, err := h.meta.CountCollections(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	responses := make([]CollectionResponse, len(cols))
	for i, col := range cols {
		responses[i] = toResponse(col, !h.enforceACL)
	}
	return c.JSON(fiber.Map{"collections": responses, "total": total, "limit": limit, "offset": offset})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return c.JSON(toResponse(col, !h.enforceACL))
}

func (h *Handler) GetBySlug(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return c.JSON(toResponse(col, !h.enforceACL))
}

func (h *Handler) PatchCollection(c *fiber.Ctx) error {
	var req PatchRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	id := c.Params("id")
	col, err := h.meta.GetCollectionByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	if h.enforceACL && (req.RefreshEnabled != nil || req.RefreshIntervalSeconds != nil) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "refresh_enabled/refresh_interval_seconds are only configurable via the admin surface"})
	}
	if req.RefreshIntervalSeconds != nil && *req.RefreshIntervalSeconds < h.minRefreshSeconds {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("refresh_interval_seconds must be at least %d", h.minRefreshSeconds)})
	}

	patch := meta.CollectionPatch{
		Name:                   req.Name,
		RefreshEnabled:         req.RefreshEnabled,
		RefreshIntervalSeconds: req.RefreshIntervalSeconds,
	}
	col, err = h.meta.UpdateCollection(c.Context(), id, patch)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(toResponse(col, !h.enforceACL))
}

func (h *Handler) DeleteByID(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return h.deleteCollection(c, col)
}

func (h *Handler) DeleteBySlug(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return h.deleteCollection(c, col)
}

func (h *Handler) deleteCollection(c *fiber.Ctx, col meta.Collection) error {
	if err := h.vec.DropTable(c.Context(), col.TableName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.meta.DeleteCollection(c.Context(), col.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
