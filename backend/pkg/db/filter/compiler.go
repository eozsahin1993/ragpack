// Package filter compiles MongoDB-style filter DSL into DataFusion SQL predicates.
package filter

import (
	"encoding/json"
	"fmt"
)

// Compile takes a raw JSON filter object and returns a DataFusion SQL predicate string.
// fieldMap must contain all declared metadata fields for the collection.
// Returns an empty string if raw is nil or empty.
func Compile(raw json.RawMessage, fieldMap FieldMap) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}

	var node interface{}
	if err := json.Unmarshal(raw, &node); err != nil {
		return "", fmt.Errorf("filter: invalid JSON: %w", err)
	}

	return compileNode(node, fieldMap)
}
