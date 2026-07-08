package metadata_fields

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"ragpack/pkg/api/middleware"
	"ragpack/pkg/api/validate"
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

func (h *Handler) Register(c *fiber.Ctx) error {
	col := c.Locals(middleware.LocalCollection).(meta.Collection)

	var req RegisterFieldsRequest
	if err := validate.Body(c, &req); err != nil {
		return err
	}

	if err := checkForDuplicates(req.Fields); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.verifyNoExistingFields(c, col.ID, req.Fields); err != nil {
		return err
	}

	inputs := make([]meta.MetadataFieldInput, len(req.Fields))
	for i, f := range req.Fields {
		inputs[i] = meta.MetadataFieldInput{Name: f.Name, Type: f.Type}
	}

	fields, err := h.meta.RegisterMetadataFields(c.Context(), col.ID, inputs)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.createIndexesInVectorDb(c, col.ID, col.TableName, fields); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"fields": fields})
}

func (h *Handler) List(c *fiber.Ctx) error {
	col := c.Locals(middleware.LocalCollection).(meta.Collection)

	fields, err := h.meta.ListMetadataFields(c.Context(), col.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"fields": fields})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	col := c.Locals(middleware.LocalCollection).(meta.Collection)
	name := c.Params("name")

	// Delete SQLite row first — new ingest/queries stop routing to this slot immediately
	field, err := h.meta.DeleteMetadataField(c.Context(), col.ID, name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("field %q not found", name)})
	}

	colName := db.MetadataSlotColumn(field.Type, field.Slot)

	// Null out the slot data in LanceDB, drop the index, then optimize
	if err := h.vec.NullMetadataSlot(c.Context(), col.TableName, colName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if dropErr := h.vec.DropMetadataIndex(c.Context(), col.TableName, colName); dropErr != nil {
		log.Printf("metadata-fields: drop index %s: %v", colName, dropErr)
	}
	if optErr := h.vec.OptimizeIndex(c.Context(), col.TableName); optErr != nil {
		log.Printf("metadata-fields: optimize after delete %s: %v", colName, optErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// checkForDuplicates returns an error if the same field name appears more than once in a single request.
func checkForDuplicates(fields []RegisterFieldItem) error {
	seen := make(map[string]bool, len(fields))
	for _, f := range fields {
		key := strings.ToLower(f.Name)
		if seen[key] {
			return fmt.Errorf("duplicate field name %q in request", f.Name)
		}
		seen[key] = true
	}
	return nil
}

// verifyNoExistingFields returns a 409 if any of the requested names are already registered.
func (h *Handler) verifyNoExistingFields(c *fiber.Ctx, collectionID string, requested []RegisterFieldItem) error {
	existing, err := h.meta.ListMetadataFields(c.Context(), collectionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load existing fields"})
	}
	existingNames := make(map[string]bool, len(existing))
	for _, f := range existing {
		existingNames[strings.ToLower(f.Name)] = true
	}
	for _, f := range requested {
		if existingNames[strings.ToLower(f.Name)] {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": fmt.Sprintf("field %q is already registered on this collection", f.Name),
			})
		}
	}
	return nil
}

// createIndexesInVectorDb creates a LanceDB index for each registered field.
// On any failure, previously created indexes and their SQLite rows are rolled back.
func (h *Handler) createIndexesInVectorDb(c *fiber.Ctx, collectionID, tableName string, fields []meta.MetadataField) error {
	for i, f := range fields {
		colName := db.MetadataSlotColumn(f.Type, f.Slot)
		// Null the slot before use — clears stale data if the slot was previously
		// occupied by a deleted property whose NullMetadataSlot call failed.
		if nullErr := h.vec.NullMetadataSlot(c.Context(), tableName, colName); nullErr != nil {
			h.rollbackIndexes(c, collectionID, tableName, fields[:i])
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to clear slot for field %q: %v", f.Name, nullErr),
			})
		}
		if indexErr := h.vec.CreateMetadataIndex(c.Context(), tableName, colName, f.Type); indexErr != nil {
			h.rollbackIndexes(c, collectionID, tableName, fields[:i])
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to create index for field %q: %v", f.Name, indexErr),
			})
		}
	}
	return nil
}

// rollbackIndexes drops any indexes already created and removes their SQLite rows.
func (h *Handler) rollbackIndexes(c *fiber.Ctx, collectionID, tableName string, created []meta.MetadataField) {
	for _, f := range created {
		colName := db.MetadataSlotColumn(f.Type, f.Slot)
		if dropErr := h.vec.DropMetadataIndex(c.Context(), tableName, colName); dropErr != nil {
			log.Printf("metadata-fields: rollback drop index %s: %v", colName, dropErr)
		}
		if _, delErr := h.meta.DeleteMetadataField(c.Context(), collectionID, f.Name); delErr != nil {
			log.Printf("metadata-fields: rollback delete field %q: %v", f.Name, delErr)
		}
	}
}
