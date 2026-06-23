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

func (h *Handler) GetJobsByCollection(c *fiber.Ctx) error {
	collection, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	if statusParam := c.Query("status"); statusParam != "" {
		status := meta.JobStatus(statusParam)
		switch status {
		case meta.JobStatusPending, meta.JobStatusProcessing, meta.JobStatusComplete, meta.JobStatusFailed:
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status"})
		}
		js, err := h.meta.ListJobsByCollectionAndStatus(c.Context(), collection.ID, status)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"jobs": js})
	}

	js, err := h.meta.ListJobsByCollection(c.Context(), collection.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jobs": js})
}

func (h *Handler) GetAllJobs(c *fiber.Ctx) error {
	if statusParam := c.Query("status"); statusParam != "" {
		status := meta.JobStatus(statusParam)
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

	js, err := h.meta.ListAllJobs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jobs": js})
}

func (h *Handler) GetJob(c *fiber.Ctx) error {
	job, err := h.meta.GetJob(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	return c.JSON(job)
}
