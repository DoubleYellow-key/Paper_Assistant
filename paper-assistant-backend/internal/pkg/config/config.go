package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr string
	LLM      LLMConfig
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

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{
		HTTPAddr: addr,
		LLM:      loadLLMConfig(),
	}
}

func loadLLMConfig() LLMConfig {
	provider := getEnv("LLM_PROVIDER", "volcengine")

	defaultBaseURL := "https://ark.cn-beijing.volces.com/api/v3"
	if provider == "aliyun" {
		defaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}

	return LLMConfig{
		Provider: provider,
		APIKey:   getEnv("LLM_API_KEY", ""),
		BaseURL:  normalizeBaseURL(getEnv("LLM_BASE_URL", defaultBaseURL)),
		Model:    getEnv("LLM_MODEL", "qwen-plus"),
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
