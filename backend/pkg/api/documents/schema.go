package documents

import (
	"encoding/json"
	"fmt"
)

type UpdateRequest struct {
	Name      *string `json:"name"`
	ExtraJSON *string `json:"extra_json"`
}

func (r *UpdateRequest) Validate() error {
	if r.Name == nil && r.ExtraJSON == nil {
		return fmt.Errorf("at least one field (name, extra_json) must be provided")
	}
	if r.ExtraJSON != nil && !json.Valid([]byte(*r.ExtraJSON)) {
		return fmt.Errorf("extra_json must be valid JSON")
	}
	return nil
}
