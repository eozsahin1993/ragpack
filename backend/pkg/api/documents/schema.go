package documents

import (
	"encoding/json"
	"fmt"
)

type UpdateRequest struct {
	Name      *string        `json:"name"`
	ExtraJSON *string        `json:"extra_json"`
	Metadata  map[string]any `json:"metadata"`
}

func (r *UpdateRequest) Validate() error {
	if r.Name == nil && r.ExtraJSON == nil && r.Metadata == nil {
		return fmt.Errorf("at least one field (name, extra_json, metadata) must be provided")
	}
	if r.ExtraJSON != nil && !json.Valid([]byte(*r.ExtraJSON)) {
		return fmt.Errorf("extra_json must be valid JSON")
	}
	return nil
}
