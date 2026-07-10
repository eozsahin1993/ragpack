package query

import "encoding/json"

type QueryRequest struct {
	Query   string          `json:"query"   validate:"required,min=1"`
	TopK    int             `json:"top_k"   validate:"omitempty,min=1,max=100"`
	Filters json.RawMessage `json:"filters"`
	// Hybrid (vector+keyword) search runs by default; set true to skip the keyword/FTS pass.
	VectorSearchOnly bool `json:"vector_search_only"`
	// Optional per-request override of the weighted RRF merge; unset fields
	HybridSettings *HybridSettings `json:"hybrid_settings"`
}

// HybridSettings overrides the weighted RRF merge for hybrid search.
type HybridSettings struct {
	FullTextWeight *float32 `json:"full_text_weight" validate:"omitempty,gte=0"`
	SemanticWeight *float32 `json:"semantic_weight"  validate:"omitempty,gte=0"`
	RRFK           *float32 `json:"rrf_k"             validate:"omitempty,gt=0"`
	FullTextLimit  *int     `json:"full_text_limit"   validate:"omitempty,min=1,max=1000"`
}

type QueryResultItem struct {
	Source           string  `json:"source"`
	FileUri          string  `json:"file_uri"`
	MimeType         string  `json:"mime_type"`
	ChunkIndex       int     `json:"chunk_index"`
	ChunkHeader      *string `json:"chunk_header"`
	ChunkText        *string `json:"chunk_text"`
	ExtraJSON        *string `json:"extra_json"`
	VectorDistance   float32 `json:"vector_distance"`
	VectorSimilarity float32 `json:"vector_similarity"`
	// Present only for hybrid (vector_search_only=false) results; see db.ChunkQueryResult.
	KeywordBM25Score   float32                `json:"keyword_bm25_score,omitempty"`
	RRFScoreNormalized float32                `json:"rrf_score_normalized,omitempty"`
	RRFScore           float32                `json:"rrf_score,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

type QueryResponse struct {
	Results []QueryResultItem `json:"results"`
}

type RagRequest struct {
	Query         string          `json:"query"          validate:"required,min=1"`
	TopK          int             `json:"top_k"          validate:"omitempty,min=1,max=100"`
	PromptSlug    string          `json:"prompt_slug"`
	Model         string          `json:"model"`
	MinSimilarity *float32        `json:"min_similarity" validate:"omitempty,min=0,max=100"`
	Filters       json.RawMessage `json:"filters"`
	// Hybrid (vector+keyword) search runs by default; set true to skip the keyword/FTS pass.
	VectorSearchOnly bool            `json:"vector_search_only"`
	HybridSettings   *HybridSettings `json:"hybrid_settings"`
}

type RagChunk struct {
	Source             string  `json:"source"`
	FileUri            string  `json:"file_uri"`
	ChunkIndex         int     `json:"chunk_index"`
	ChunkHeader        *string `json:"chunk_header"`
	ChunkText          *string `json:"chunk_text"`
	VectorSimilarity   float32 `json:"vector_similarity"`
	KeywordBM25Score   float32 `json:"keyword_bm25_score,omitempty"`
	RRFScoreNormalized float32 `json:"rrf_score_normalized,omitempty"`
	RRFScore           float32 `json:"rrf_score,omitempty"`
}

type RagResponse struct {
	FormattedPrompt string     `json:"formatted_prompt"`
	Answer          string     `json:"answer"`
	Chunks          []RagChunk `json:"chunks"`
	PromptSlug      string     `json:"prompt_slug"`
}
