package config

import "ragpack/pkg/chunker"

const (
	// Server
	DefaultPort     = "9000"
	DefaultDataPath = "./data"

	// Storage
	DefaultSqlitePath  = "./ragpack.db"
	DefaultLanceDBPath = "./lancedb_data"

	// Embedder
	DefaultEmbedProvider = "ollama"
	DefaultOpenAIModel   = "text-embedding-3-small"
	DefaultOllamaBaseURL = "http://localhost:11434"
	DefaultOllamaModel   = "nomic-embed-text"

	// Ingester
	DefaultWorkerCount    = 5
	DefaultEmbedRateLimit = 10.0

	// Chunking
	DefaultChunkStrategy = chunker.StrategyAuto
	DefaultChunkSize     = 2000
	DefaultChunkOverlap  = 200
)
