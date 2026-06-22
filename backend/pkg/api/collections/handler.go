package collections

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/backend/pkg/db"
	"ragpack/backend/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
	vec  db.VectorDb
}

func NewHandler(ms meta.MetaStore, vec db.VectorDb) *Handler {
	return &Handler{meta: ms, vec: vec}
}

type createRequest struct {
	Name       string `json:"name"`
	EmbedModel string `json:"embed_model"`
	VectorDim  int    `json:"vector_dim"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req createRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Name == "" || req.EmbedModel == "" || req.VectorDim <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, embed_model, and vector_dim are required"})
	}

	col, err := h.meta.CreateCollection(c.Context(), req.Name, req.EmbedModel, req.VectorDim)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.vec.CreateTable(c.Context(), col.TableName, col.VectorDim); err != nil {
		_ = h.meta.DeleteCollection(c.Context(), col.Name)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(col)
}

func (h *Handler) List(c *fiber.Ctx) error {
	cols, err := h.meta.ListCollections(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"collections": cols})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return c.JSON(col)
}

func (h *Handler) GetByName(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByName(c.Context(), c.Params("name"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return c.JSON(col)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByName(c.Context(), c.Params("name"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return h.deleteCollection(c, col)
}

func (h *Handler) DeleteByID(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByID(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}
	return h.deleteCollection(c, col)
}

func (h *Handler) deleteCollection(c *fiber.Ctx, col meta.Collection) error {
	if err := h.vec.DropTable(c.Context(), col.TableName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.meta.DeleteCollection(c.Context(), col.Name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
