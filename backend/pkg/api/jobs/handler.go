package jobs

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
}

func NewHandler(ms meta.MetaStore) *Handler {
	return &Handler{meta: ms}
}

func (h *Handler) ListByCollection(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionByName(c.Context(), c.Params("name"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	js, err := h.meta.ListJobsByCollection(c.Context(), col.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jobs": js})
}

func (h *Handler) GetJob(c *fiber.Ctx) error {
	j, err := h.meta.GetJob(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	return c.JSON(j)
}

func (h *Handler) ListByStatus(c *fiber.Ctx) error {
	status := meta.JobStatus(c.Params("status"))
	switch status {
	case meta.JobStatusPending, meta.JobStatusProcessing, meta.JobStatusComplete, meta.JobStatusFailed:
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status"})
	}

	js, err := h.meta.ListJobsByStatus(c.Context(), status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jobs": js})
}
