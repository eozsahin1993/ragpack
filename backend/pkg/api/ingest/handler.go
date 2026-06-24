package ingest

import (
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/ingester"
	"ragpack/pkg/meta"
)

type Handler struct {
	meta meta.MetaStore
	ing  ingester.Ingester
}

func NewHandler(ms meta.MetaStore, ing ingester.Ingester) *Handler {
	return &Handler{meta: ms, ing: ing}
}

func (h *Handler) Ingest(c *fiber.Ctx) error {
	col, err := h.meta.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	// multipart upload
	if file, err := c.FormFile("file"); err == nil {
		mimeType := file.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "text/plain"
		}

		f, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
		}

		job, err := h.meta.CreateJob(c.Context(), col.ID, "upload://"+file.Filename, mimeType)
		if err != nil {
			f.Close()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		h.ing.Submit(job, f)
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	// URI-based
	var req URIRequest
	if err := c.BodyParser(&req); err != nil || req.FileURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provide a file upload or {file_uri}"})
	}
	if req.MimeType == "" {
		req.MimeType = detectMimeType(c.Context(), req.FileURI)
	}

	job, err := h.meta.CreateJob(c.Context(), col.ID, req.FileURI, req.MimeType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.ing.Submit(job, nil)
	return c.Status(fiber.StatusAccepted).JSON(job)
}
