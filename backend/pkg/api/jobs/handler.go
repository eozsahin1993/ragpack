package jobs

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/middleware"
	"ragpack/pkg/api/validate"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta       meta.MetaStore
	enforceACL bool
}

// NewHandler builds a jobs handler. Like documents, mounted both under
// /collections/:slug and at the top level, so the access check happens here.
// enforceACL is false only from RegisterAdmin, which has no Auth middleware.
func NewHandler(ms meta.MetaStore, enforceACL bool) *Handler {
	return &Handler{meta: ms, enforceACL: enforceACL}
}

func (h *Handler) checkAccess(c *fiber.Ctx, collectionID string, required meta.Permission) error {
	if !h.enforceACL {
		return nil
	}
	return middleware.CheckAccess(c, h.meta, collectionID, required)
}

// ListJobs returns jobs, optionally scoped by collection (via middleware) and/or status.
func (h *Handler) ListJobs(c *fiber.Ctx) error {
	limit, offset := validate.Pagination(c)

	filter := meta.JobFilter{}
	if col, ok := c.Locals(middleware.LocalCollection).(meta.Collection); ok {
		if err := h.checkAccess(c, col.ID, meta.PermissionRead); err != nil {
			return err
		}
		filter.CollectionID = &col.ID
	} else if h.enforceACL {
		// No collection in the URL: only an unrestricted key can list across
		// every collection — see documents.List for the same reasoning.
		if err := middleware.RequireWildcard(h.meta, meta.PermissionRead)(c); err != nil {
			return err
		}
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
	if err := h.checkAccess(c, job.CollectionID, meta.PermissionRead); err != nil {
		return err
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
	if err := h.checkAccess(c, job.CollectionID, meta.PermissionWrite); err != nil {
		return err
	}
	if job.Status == meta.JobStatusPending || job.Status == meta.JobStatusProcessing {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "cannot delete an active job"})
	}
	if err := h.meta.DeleteJob(c.Context(), job.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
