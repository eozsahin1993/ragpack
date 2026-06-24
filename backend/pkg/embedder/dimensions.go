package embedder

import "strings"

// ModelDimensions maps well-known model names to their fixed output dimensions.
// Lookup is done on the base model name (tag stripped), case-insensitive.
var ModelDimensions = map[string]int{
	// Ollama
	"nomic-embed-text":       768,
	"mxbai-embed-large":      1024,
	"all-minilm":             384,
	"bge-m3":                 1024,
	"snowflake-arctic-embed": 1024,
	"bge-large":              1024,
	"bge-base":               768,

	// OpenAI
	"text-embedding-ada-002":   1536,
	"text-embedding-3-small":   1536,
	"text-embedding-3-large":   3072,
}

// DimensionsForModel returns the vector dimension for the given model name.
// The tag suffix (e.g. ":latest") is stripped before lookup.
func DimensionsForModel(model string) (int, bool) {
	base := strings.ToLower(strings.SplitN(model, ":", 2)[0])
	dim, ok := ModelDimensions[base]
	return dim, ok
}
