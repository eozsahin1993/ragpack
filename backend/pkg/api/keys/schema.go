package keys

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type CreateResponse struct {
	Key     string `json:"key"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	KeyHint string `json:"key_hint"`
}
