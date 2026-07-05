package documents

type UpdateRequest struct {
	Name string `json:"name" validate:"required"`
}
