package config

import (
	"os"
)

type Config struct {
	HTTPAddr string
	LLM      LLMConfig
	MySQL    MySQLConfig
}

type LLMConfig struct {
	// OpenAI compatible API key.
	APIKey string
	// 固定使用阿里云兼容端点
	BaseURL string
	// 固定模型名
	Model string
}

type MySQLConfig struct {
	DSN string
}

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{
		HTTPAddr: addr,
		LLM:      loadLLMConfig(),
		MySQL:    loadMySQLConfig(),
	}
}

func loadLLMConfig() LLMConfig {
	return LLMConfig{
		APIKey:  getEnv("LLM_API_KEY", ""),
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1/",
		Model:   "deepseek-v4-flash",
	}
}

func loadMySQLConfig() MySQLConfig {
	return MySQLConfig{
		DSN: getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=true&loc=Local"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
