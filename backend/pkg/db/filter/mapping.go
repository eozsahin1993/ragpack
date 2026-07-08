package filter

import (
	"fmt"

	"ragpack/pkg/meta"
)

// FieldMap is an inverted lookup: user field name → MetadataField.
// Build it from ListMetadataFields results.
type FieldMap map[string]meta.MetadataField

// resolveColumn resolves a user field name to a DataFusion column name and logical type.
// Built-in columns are returned directly; metadata fields are looked up in fieldMap.
func resolveColumn(name string, fieldMap FieldMap) (col, fieldType string, err error) {
	switch name {
	case "created_at", "updated_at":
		return name, "timestamp", nil
	case "mime_type", "source_name", "external_id", "file_uri":
		return name, "str", nil
	}

	field, ok := fieldMap[name]
	if !ok {
		return "", "", fmt.Errorf("filter: field %q is not registered on this collection", name)
	}
	return slotColumn(field.Type, field.Slot), field.Type, nil
}

func slotColumn(fieldType string, slot int) string {
	return fmt.Sprintf("metadata_%s_%d", fieldType, slot)
}
