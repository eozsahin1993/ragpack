package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	AdminPort   string
	DataPath    string
	SqlitePath  string
	LanceDBPath string

	EmbedProvider string
	LLMProvider   string

	OpenAI      OpenAIConfig
	Ollama      OllamaConfig
	HuggingFace HuggingFaceConfig
	TEI         TEIConfig
	Anthropic   AnthropicConfig

	Ingester          IngesterConfig
	DefaultPromptSlug string
	MaxUploadSizeMB   int
}

type OpenAIConfig struct {
	APIKey   string
	Model    string
	LLMModel string
}

type OllamaConfig struct {
	BaseURL  string
	Model    string
	LLMModel string
}

type AnthropicConfig struct {
	APIKey string
	Model  string
}

type HuggingFaceConfig struct {
	APIKey string
	Model  string
}

type TEIConfig struct {
	BaseURL string
	Model   string
}

type IngesterConfig struct {
	WorkerCount    int
	EmbedRateLimit float64
	ChunkSize      int
	ChunkOverlap   int
	ChunkStrategy  string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	return Config{
		Port:          getEnv("PORT", DefaultPort),
		AdminPort:     getEnv("ADMIN_PORT", DefaultAdminPort),
		DataPath:      getEnv("DATA_PATH", DefaultDataPath),
		SqlitePath:    getEnv("SQLITE_PATH", DefaultSqlitePath),
		LanceDBPath:   getEnv("LANCEDB_PATH", DefaultLanceDBPath),
		EmbedProvider: getEnv("DEFAULT_EMBED_PROVIDER", DefaultEmbedProvider),
		LLMProvider:   getEnv("DEFAULT_LLM_PROVIDER", ""),

		OpenAI: OpenAIConfig{
			APIKey:   getEnv("OPENAI_API_KEY", ""),
			Model:    getEnv("OPENAI_EMBED_MODEL", DefaultOpenAIModel),
			LLMModel: getEnv("OPENAI_LLM_MODEL", ""),
		},

		Ollama: OllamaConfig{
			BaseURL:  getEnv("OLLAMA_BASE_URL", DefaultOllamaBaseURL),
			Model:    getEnv("OLLAMA_EMBED_MODEL", DefaultOllamaModel),
			LLMModel: getEnv("OLLAMA_LLM_MODEL", ""),
		},

		Anthropic: AnthropicConfig{
			APIKey: getEnv("ANTHROPIC_API_KEY", ""),
			Model:  getEnv("ANTHROPIC_MODEL", DefaultAnthropicModel),
		},

		HuggingFace: HuggingFaceConfig{
			APIKey: getEnv("HF_API_KEY", ""),
			Model:  getEnv("HF_EMBED_MODEL", DefaultHFModel),
		},

		TEI: TEIConfig{
			BaseURL: getEnv("TEI_BASE_URL", DefaultTEIBaseURL),
			Model:   getEnv("TEI_EMBED_MODEL", DefaultTEIModel),
		},

		DefaultPromptSlug: getEnv("DEFAULT_PROMPT_SLUG", DefaultPromptSlug),
		MaxUploadSizeMB:   getEnvInt("MAX_UPLOAD_SIZE_MB", DefaultMaxUploadSize),

		Ingester: IngesterConfig{
			WorkerCount:    getEnvInt("WORKER_COUNT", DefaultWorkerCount),
			EmbedRateLimit: getEnvFloat("EMBED_RATE_LIMIT", DefaultEmbedRateLimit),
			ChunkSize:      getEnvInt("CHUNK_SIZE", DefaultChunkSize),
			ChunkOverlap:   getEnvInt("CHUNK_OVERLAP", DefaultChunkOverlap),
			ChunkStrategy:  getEnv("CHUNK_STRATEGY", DefaultChunkStrategy),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
