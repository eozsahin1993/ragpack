package query

type QueryRequest struct {
	Query string `json:"query" validate:"required,min=1"`
	TopK  int    `json:"top_k" validate:"min=1,max=100"`
}
