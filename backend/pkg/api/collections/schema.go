package collections

type CreateRequest struct {
	Name       string `json:"name"        validate:"required,min=1,max=100"`
	EmbedModel string `json:"embed_model" validate:"required"`
}

type PatchRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}
