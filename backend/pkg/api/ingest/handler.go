package ingest

import (
	"io"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/ingester"
	"ragpack/pkg/meta"
)

type Handler struct {
	store         meta.MetaStore
	ingester      ingester.Ingester
	maxUploadSize int64
}

func NewHandler(store meta.MetaStore, queue ingester.Ingester, maxUploadSize int) *Handler {
	return &Handler{
		store:         store,
		ingester:      queue,
		maxUploadSize: int64(maxUploadSize) * 1024 * 1024,
	}
}

// Ingest dispatches to ingestMultipart or ingestURI based on whether a multipart file is present.
func (handler *Handler) Ingest(c *fiber.Ctx) error {
	collection, err := handler.store.GetCollectionBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "collection not found"})
	}

	intent := meta.JobIntentIngest
	if c.Query("refresh") == "true" {
		intent = meta.JobIntentRefresh
	}
	force := c.Query("force") == "true"

	if _, err := c.FormFile("file"); err == nil {
		return handler.ingestMultipart(c, collection, intent, force)
	}
	return handler.ingestURI(c, collection, intent, force)
}

func (handler *Handler) ingestMultipart(c *fiber.Ctx, collection meta.Collection, intent meta.JobIntent, force bool) error {
	file, _ := c.FormFile("file")
	if file.Size > handler.maxUploadSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{"error": "file exceeds max upload size"})
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "text/plain"
	}
	if err := validateFile(file.Filename, mimeType); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	var extraJSON *string
	if raw := c.FormValue("extra_json"); raw != "" {
		if !validateExtraJSONString(&raw) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "extra_json must be valid JSON"})
		}
		extraJSON = &raw
	}

	reader, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
	}

	fileUri := "upload://" + file.Filename
	return handler.submitJob(c, collection, intent, force, fileUri, mimeType, extraJSON, reader)
}

func (handler *Handler) ingestURI(c *fiber.Ctx, collection meta.Collection, intent meta.JobIntent, force bool) error {
	var req URIRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provide a file upload or {file_uri}"})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.MimeType == "" {
		req.MimeType = detectMimeType(c.Context(), req.FileURI)
	}
	if err := validateFile(req.FileURI, req.MimeType); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	return handler.submitJob(c, collection, intent, force, req.FileURI, req.MimeType, req.ExtraJSON, nil)
}

// submitJob checks for a completed duplicate, creates the job record, and queues it for processing.
func (handler *Handler) submitJob(c *fiber.Ctx, collection meta.Collection, intent meta.JobIntent, force bool, fileUri, mimeType string, extraJSON *string, reader io.ReadCloser) error {
	if doc, skip := handler.skipIfComplete(c, collection.ID, fileUri, intent, force); skip {
		return c.Status(fiber.StatusOK).JSON(doc)
	}

	job, err := handler.store.CreateJob(c.Context(), collection.ID, fileUri, mimeType, intent, force, extraJSON)
	if err != nil {
		if reader != nil {
			reader.Close()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	handler.ingester.Submit(job, reader)
	return c.Status(fiber.StatusAccepted).JSON(job)
}

// skipIfComplete returns the existing document and true when intent=ingest (no force) and
// the document is already complete.
func (handler *Handler) skipIfComplete(c *fiber.Ctx, collectionID, fileUri string, intent meta.JobIntent, force bool) (meta.Document, bool) {
	if intent != meta.JobIntentIngest || force {
		return meta.Document{}, false
	}
	doc, err := handler.store.FindDocumentByFileUri(c.Context(), collectionID, fileUri)
	if err != nil || doc == nil {
		return meta.Document{}, false
	}
	if doc.Status == meta.DocumentStatusComplete {
		return *doc, true
	}
	return meta.Document{}, false
}
