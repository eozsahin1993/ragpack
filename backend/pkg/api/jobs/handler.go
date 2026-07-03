package jobs

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/middleware"
	"ragpack/pkg/api/validate"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
}

func NewHandler(ms meta.MetaStore) *Handler {
	return &Handler{meta: ms}
}

// ListJobs returns jobs, optionally scoped by collection (via middleware) and/or status.
func (h *Handler) ListJobs(c *fiber.Ctx) error {
	limit, offset := validate.Pagination(c)

	filter := meta.JobFilter{}
	if col, ok := c.Locals(middleware.LocalCollection).(meta.Collection); ok {
		filter.CollectionID = &col.ID
	}
	if s := c.Query("status"); s != "" {
		status := meta.JobStatus(s)
		switch status {
		case meta.JobStatusPending, meta.JobStatusProcessing, meta.JobStatusComplete, meta.JobStatusFailed:
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status"})
		}
		filter.Status = &status
	}

	js, err := h.meta.ListJobs(c.Context(), filter, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	total, err := h.meta.CountJobs(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jobs": js, "total": total, "limit": limit, "offset": offset})
}

// GetJob fetches a single job by :id. When Collection middleware is active,
// the job must belong to that collection.
func (h *Handler) GetJob(c *fiber.Ctx) error {
	job, err := h.meta.GetJob(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	if col, scoped := c.Locals(middleware.LocalCollection).(meta.Collection); scoped && job.CollectionID != col.ID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	return c.JSON(job)
}

// DeleteJob deletes a completed or failed job by :id. When Collection middleware
// is active, the job must belong to that collection.
func (h *Handler) DeleteJob(c *fiber.Ctx) error {
	job, err := h.meta.GetJob(c.Context(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	if col, scoped := c.Locals(middleware.LocalCollection).(meta.Collection); scoped && job.CollectionID != col.ID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "job not found"})
	}
	if job.Status == meta.JobStatusPending || job.Status == meta.JobStatusProcessing {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "cannot delete an active job"})
	}
	if err := h.meta.DeleteJob(c.Context(), job.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
