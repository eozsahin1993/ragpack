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

	intent := meta.JobIntentIngest
	if c.Query("refresh") == "true" {
		intent = meta.JobIntentRefresh
	}
	force := c.Query("force") == "true"

	// multipart upload
	if file, err := c.FormFile("file"); err == nil {
		mimeType := file.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "text/plain"
		}
		if err := validateFile(file.Filename, mimeType); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		}
		fileUri := "upload://" + file.Filename

		if doc, skip := h.skipIfComplete(c, col.ID, fileUri, intent, force); skip {
			return c.Status(fiber.StatusOK).JSON(doc)
		}

		f, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
		}

		job, err := h.meta.CreateJob(c.Context(), col.ID, fileUri, mimeType, intent, force)
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
	if !validateURI(req.FileURI) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "unsupported URI scheme — use https://, http://, or s3://"})
	}
	if req.MimeType == "" {
		req.MimeType = detectMimeType(c.Context(), req.FileURI)
	}
	if err := validateFile(req.FileURI, req.MimeType); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	if doc, skip := h.skipIfComplete(c, col.ID, req.FileURI, intent, force); skip {
		return c.Status(fiber.StatusOK).JSON(doc)
	}

	job, err := h.meta.CreateJob(c.Context(), col.ID, req.FileURI, req.MimeType, intent, force)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.ing.Submit(job, nil)
	return c.Status(fiber.StatusAccepted).JSON(job)
}

// skipIfComplete returns the existing document and true when intent=ingest (no force) and
// the document is already complete. Returns nil, false otherwise.
func (h *Handler) skipIfComplete(c *fiber.Ctx, collectionID, fileUri string, intent meta.JobIntent, force bool) (meta.Document, bool) {
	if intent != meta.JobIntentIngest || force {
		return meta.Document{}, false
	}
	doc, err := h.meta.FindDocumentByFileUri(c.Context(), collectionID, fileUri)
	if err != nil || doc == nil {
		return meta.Document{}, false
	}
	if doc.Status == meta.DocumentStatusComplete {
		return *doc, true
	}
	return meta.Document{}, false
}
