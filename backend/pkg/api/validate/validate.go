package validate

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var v = validator.New()

const (
	DefaultLimit = 50
	MaxLimit     = 500
)

// Pagination parses ?limit= and ?offset= from the request query string.
// limit is clamped to [1, MaxLimit]; offset defaults to 0.
func Pagination(c *fiber.Ctx) (limit, offset int) {
	limit = DefaultLimit
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > MaxLimit {
				v = MaxLimit
			}
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}

// Body parses and validates a request body. Returns a structured 400 on failure.
func Body(c *fiber.Ctx, req any) error {
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := v.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": fieldErrors(err)})
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
	default:
		return fmt.Sprintf("failed %s validation", fe.Tag())
	}
}
