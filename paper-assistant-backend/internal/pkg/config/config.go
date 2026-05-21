package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr  string
	MySQL     MySQLConfig
	LLM       LLMConfig
	Embedding EmbeddingConfig
	Vector    VectorConfig
}

type MySQLConfig struct {
	DSN string
}

type VectorConfig struct {
	Path string
}

type LLMConfig struct {
	// provider: volcengine | aliyun | openai-compatible
	Provider string
	// OpenAI compatible API key.
	APIKey string
	// Base URL，不要带 /chat/completions
	BaseURL string
	// 模型名，例如 qwen-plus / doubao-1-5-pro-32k
	Model string
}

type EmbeddingConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{
		HTTPAddr:  addr,
		MySQL:     loadMySQLConfig(),
		LLM:       loadLLMConfig(),
		Embedding: loadEmbeddingConfig(),
		Vector:    loadVectorConfig(),
	}
}

func loadMySQLConfig() MySQLConfig {
	return MySQLConfig{
		DSN: getEnv("MYSQL_DSN", "paper_assistant:paper_assistant@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=True&loc=Local"),
	}
}

func loadVectorConfig() VectorConfig {
	return VectorConfig{
		Path: getEnv("VECTOR_DB_PATH", "vectordb"),
	}
}

func loadLLMConfig() LLMConfig {
	provider := getEnv("LLM_PROVIDER", "aliyun")

	defaultBaseURL := "https://dashscope.aliyuncs.com/compatible-mode/v1"
	if provider == "volcengine" {
		defaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}

	return LLMConfig{
		Provider: provider,
		APIKey:   getEnv("LLM_API_KEY", ""),
		BaseURL:  normalizeBaseURL(getEnv("LLM_BASE_URL", defaultBaseURL)),
		Model:    getEnv("LLM_MODEL", "qwen-plus"),
	}
}

func loadEmbeddingConfig() EmbeddingConfig {
	llmCfg := loadLLMConfig()
	return EmbeddingConfig{
		APIKey:  getEnv("EMBEDDING_API_KEY", llmCfg.APIKey),
		BaseURL: normalizeBaseURL(getEnv("EMBEDDING_BASE_URL", llmCfg.BaseURL)),
		Model:   getEnv("EMBEDDING_MODEL", "text-embedding-v3"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func normalizeBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "`")
	raw = strings.Trim(raw, "\"")
	raw = strings.TrimSpace(raw)

	raw = strings.TrimSuffix(raw, "/chat/completions")
	raw = strings.TrimSuffix(raw, "/")
	return raw
}
