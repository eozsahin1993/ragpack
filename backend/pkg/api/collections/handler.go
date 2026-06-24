package collections

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/validate"
	"ragpack/pkg/db"
	"ragpack/pkg/embedder"
	"ragpack/pkg/meta"
)


type Handler struct {
	meta meta.MetaStore
	vec  db.VectorDb
}

func NewHandler(ms meta.MetaStore, vec db.VectorDb) *Handler {
	return &Handler{meta: ms, vec: vec}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	dim, ok := embedder.DimensionsForModel(req.EmbedModel)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unknown embed_model: vector dimensions not found — add it to embedder.ModelDimensions",
		})
	}

	col, err := h.meta.CreateCollection(c.Context(), req.Name, req.EmbedModel, dim)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.vec.CreateTable(c.Context(), col.TableName, col.VectorDim); err != nil {
		_ = h.meta.DeleteCollection(c.Context(), col.ID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(col)
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

	return c.JSON(fiber.Map{"collections": cols, "total": total, "limit": limit, "offset": offset})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return c.JSON(col)
}

func (h *Handler) GetBySlug(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return c.JSON(col)
}

func (h *Handler) PatchCollection(c *fiber.Ctx) error {
	var req PatchRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	col, err := h.meta.UpdateCollectionName(c.Context(), c.Params("id"), req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(col)
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
