package ingest

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/db"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
	vec  db.VectorDb
}

func NewHandler(ms meta.MetaStore, vec db.VectorDb) *Handler {
	return &Handler{meta: ms, vec: vec}
}

func (h *Handler) Ingest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented"})
}
