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
	DefaultTEIBaseURL = "http://localhost:8080"
	DefaultTEIModel   = "BAAI/bge-small-en-v1.5"
	DefaultHFModel    = "BAAI/bge-small-en-v1.5"

	// LLM
	DefaultOpenAILLMModel  = "gpt-4o-mini"
	DefaultOllamaLLMModel  = "llama3.2"
	DefaultAnthropicModel  = "claude-haiku-4-5-20251001"

	// Ingester
	DefaultWorkerCount    = 5
	DefaultEmbedRateLimit = 10.0

	// Chunking
	DefaultChunkStrategy = chunker.StrategyAuto
	DefaultChunkSize     = 2000
	DefaultChunkOverlap  = 200

	// RAG
	DefaultPromptSlug = "basic_rag"
)
