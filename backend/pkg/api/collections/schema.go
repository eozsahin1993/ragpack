package collections

import (
	"time"

	"ragpack/pkg/meta"
)

// — Request types —

type ChunkConfigRequest struct {
	Strategy *string `json:"strategy" validate:"omitempty,oneof=auto unit paragraph sliding_window section row_group"`
	Size     *int    `json:"size"     validate:"omitempty,min=100,max=32000"`
	Overlap  *int    `json:"overlap"  validate:"omitempty,min=0,max=1000"`
}

type CreateRequest struct {
	Name        string              `json:"name"         validate:"required,min=1,max=100"`
	EmbedModel  string              `json:"embed_model"`
	ChunkConfig *ChunkConfigRequest `json:"chunk_config"`
}

type PatchRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// — Response types —

type CollectionResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	TableName   string               `json:"table_name"`
	EmbedModel  string               `json:"embed_model"`
	VectorDim   int                  `json:"vector_dim"`
	CreatedAt   time.Time            `json:"created_at"`
	ChunkConfig *ChunkConfigResponse `json:"chunk_config,omitempty"`
}

type ChunkConfigResponse struct {
	Strategy *string `json:"strategy,omitempty"`
	Size     *int    `json:"size,omitempty"`
	Overlap  *int    `json:"overlap,omitempty"`
}

func toResponse(c meta.Collection) CollectionResponse {
	r := CollectionResponse{
		ID:         c.ID,
		Name:       c.Name,
		Slug:       c.Slug,
		TableName:  c.TableName,
		EmbedModel: c.EmbedModel,
		VectorDim:  c.VectorDim,
		CreatedAt:  c.CreatedAt,
	}
	if c.ChunkStrategy != nil || c.ChunkSize != nil || c.ChunkOverlap != nil {
		r.ChunkConfig = &ChunkConfigResponse{
			Strategy: c.ChunkStrategy,
			Size:     c.ChunkSize,
			Overlap:  c.ChunkOverlap,
		}
	}
	return r
}
