package query

type QueryRequest struct {
	Query string `json:"query" validate:"required,min=1"`
	TopK  int    `json:"top_k" validate:"min=1,max=100"`
}

type QueryResultItem struct {
	Source     string  `json:"source"`
	ChunkIndex int     `json:"chunk_index"`
	ChunkText  *string `json:"chunk_text"`
	Distance   float32 `json:"distance"`
	Similarity float32 `json:"similarity"`
}
