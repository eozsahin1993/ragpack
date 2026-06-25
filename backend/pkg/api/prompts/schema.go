package prompts

type CreateRequest struct {
	Name    string `json:"name"    validate:"required,min=1,max=100"`
	Content string `json:"content" validate:"required,min=1"`
}

type UpdateRequest struct {
	Name    *string `json:"name"    validate:"omitempty,min=1,max=100"`
	Content *string `json:"content" validate:"omitempty,min=1"`
}
