package validate

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/meta"
)

// ErrResponseWritten is returned by Body/Query when they've already written
// an error response to c. c.JSON's own return value can't signal this (it's
// nil on a successful write, which a validation-error response always is),
// so callers must check for this sentinel via errors.Is instead of the
// c.JSON result — the app's Fiber ErrorHandler treats it as already-handled
// and skips writing a second response over it (see cmd/main.go).
var ErrResponseWritten = errors.New("validate: response already written")

var reservedMetadataNames = map[string]bool{
	"created_at": true, "updated_at": true, "mime_type": true,
	"source_name": true, "external_id": true, "document_id": true,
	"chunk_text": true, "chunk_header": true, "file_uri": true, "extra_json": true,
}

var v = validator.New()

func init() {
	v.RegisterValidation("notreservedmeta", func(fl validator.FieldLevel) bool { //nolint:errcheck
		return !reservedMetadataNames[fl.Field().String()]
	})
	RegisterSortValidator("documentsortfield", meta.DocumentSortSpec)
}

// Body parses and validates a request body. Returns ErrResponseWritten on
// failure, after writing a structured 400.
func Body(c *fiber.Ctx, req any) error {
	if err := c.BodyParser(req); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		return ErrResponseWritten
	}
	if err := v.Struct(req); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": fieldErrors(err)})
		return ErrResponseWritten
	}
	return nil
}

// Query parses and validates request query parameters into req (fields tagged
// with `query:"..."`). Returns ErrResponseWritten on failure, after writing a
// structured 400.
func Query(c *fiber.Ctx, req any) error {
	if err := c.QueryParser(req); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query parameters"})
		return ErrResponseWritten
	}
	if err := v.Struct(req); err != nil {
		_ = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": fieldErrors(err)})
		return ErrResponseWritten
	}
	return nil
}

func fieldErrors(err error) map[string]string {
	errs := make(map[string]string)
	for _, fe := range err.(validator.ValidationErrors) {
		errs[fe.Field()] = errorMessage(fe)
	}
	return errs
}

func errorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "notreservedmeta":
		return "name conflicts with a built-in field and cannot be used as a metadata field"
	default:
		if msg, ok := sortValidatorMessage(fe); ok {
			return msg
		}
		return fmt.Sprintf("failed %s validation", fe.Tag())
	}
}
