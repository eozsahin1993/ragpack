package keys

import "ragpack/pkg/meta"

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
	Name string `json:"name" validate:"required,min=1,max=100"`
	// Non-empty: a key with no grants can access nothing.
	Grants []GrantRequest `json:"grants" validate:"required,min=1,dive"`
	// Optional — most keys have none.
	AdminGrants []AdminGrantRequest `json:"admin_grants" validate:"omitempty,dive"`
}

type CreateResponse struct {
	Key         string                 `json:"key"`
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	KeyHint     string                 `json:"key_hint"`
	Grants      []meta.CollectionGrant `json:"grants"`
	AdminGrants []meta.AdminGrant      `json:"admin_grants,omitempty"`
}
