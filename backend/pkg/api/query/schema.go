package query

type QueryRequest struct {
	Query string `json:"query" validate:"required,min=1"`
	TopK  int    `json:"top_k" validate:"min=1,max=100"`
}

type QueryResultItem struct {
	Source      string  `json:"source"`
	FileUri     string  `json:"file_uri"`
	MimeType    string  `json:"mime_type"`
	ChunkIndex  int     `json:"chunk_index"`
	ChunkHeader *string `json:"chunk_header"`
	ChunkText   *string `json:"chunk_text"`
	ExtraJSON   *string `json:"extra_json"`
	Distance    float32 `json:"distance"`
	Similarity  float32 `json:"similarity"`
}

type RagRequest struct {
	Query      string `json:"query"       validate:"required,min=1"`
	TopK       int    `json:"top_k"       validate:"min=1,max=100"`
	PromptSlug string `json:"prompt_slug" validate:"required,min=1"`
}

type RagChunk struct {
	Source      string  `json:"source"`
	FileUri     string  `json:"file_uri"`
	ChunkIndex  int     `json:"chunk_index"`
	ChunkHeader *string `json:"chunk_header"`
	ChunkText   *string `json:"chunk_text"`
	Similarity  float32 `json:"similarity"`
}

type RagResponse struct {
	FormattedPrompt string     `json:"formatted_prompt"`
	Chunks          []RagChunk `json:"chunks"`
	PromptSlug      string     `json:"prompt_slug"`
}
