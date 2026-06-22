package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	SqlitePath  string
	LanceDBPath string

	EmbedProvider string

	OpenAI OpenAIConfig
	Ollama OllamaConfig

	Ingester IngesterConfig
}

type OpenAIConfig struct {
	APIKey string
	Model  string
}

type OllamaConfig struct {
	BaseURL string
	Model   string
}

type IngesterConfig struct {
	WorkerCount   int
	EmbedRateLimit float64
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	return Config{
		Port:          getEnv("PORT", "9000"),
		SqlitePath:    getEnv("SQLITE_PATH", "./ragpack.db"),
		LanceDBPath:   getEnv("LANCEDB_PATH", "./lancedb_data"),
		EmbedProvider: getEnv("EMBED_PROVIDER", "ollama"),

		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
			Model:  getEnv("OPENAI_EMBED_MODEL", "text-embedding-3-small"),
		},

		Ollama: OllamaConfig{
			BaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			Model:   getEnv("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
		},

		Ingester: IngesterConfig{
			WorkerCount:    getEnvInt("WORKER_COUNT", 5),
			EmbedRateLimit: getEnvFloat("EMBED_RATE_LIMIT", 10),
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
