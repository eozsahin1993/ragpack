package keys

import (
	"fmt"

	"ragpack/pkg/meta"
)

// GrantRequest is a collection grant. An omitted CollectionSlug means every collection. See meta.CollectionGrant.
type GrantRequest struct {
	CollectionSlug string          `json:"collection_slug"`
	Permission     meta.Permission `json:"permission" validate:"required,oneof=read write both"`
}

// AdminGrantRequest is an instance-administration grant. See meta.AdminGrant.
type AdminGrantRequest struct {
	ResourceType meta.ResourceType `json:"resource_type" validate:"required,oneof=keys prompts collections *"`
	Permission   meta.Permission   `json:"permission"     validate:"required,oneof=read write both"`
}

type CreateRequest struct {
	Name        string              `json:"name"         validate:"required,min=1,max=100"`
	Grants      []GrantRequest      `json:"grants"        validate:"omitempty,dive"`
	AdminGrants []AdminGrantRequest `json:"admin_grants"  validate:"omitempty,dive"`
}

func (r *CreateRequest) Validate() error {
	if len(r.Grants) == 0 && len(r.AdminGrants) == 0 {
		return fmt.Errorf("at least one grant (collection or admin) is required")
	}
	return nil
}

// KeyResponse is an API key plus its grants — the shape returned by List,
// and embedded in CreateResponse (which adds the one field List must never
// expose: the plaintext key).
type KeyResponse struct {
	meta.APIKey
	Grants      []meta.CollectionGrant `json:"grants"`
	AdminGrants []meta.AdminGrant      `json:"admin_grants,omitempty"`
}

// CreateResponse is only returned once, from Create — the plaintext key is
// never retrievable again afterward.
type CreateResponse struct {
	KeyResponse
	Key string `json:"key"`
}
