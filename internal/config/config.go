package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string

	// S3
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3BucketName      string
	S3UseSSL          bool

	// OpenRouter
	OpenRouterAPIKey string
	OpenRouterModel  string

	// Upload limits
	MaxFileSize int64
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/docapi?sslmode=disable"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		S3Endpoint:        getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", "minioadmin"),
		S3SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", "minioadmin"),
		S3BucketName:      getEnv("S3_BUCKET_NAME", "documents"),
		S3UseSSL:          getEnv("S3_USE_SSL", "false") == "true",
		OpenRouterAPIKey:  getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:   getEnv("OPENROUTER_MODEL", "openai/gpt-4o-mini"),
		MaxFileSize:       5 * 1024 * 1024,
	}

	if cfg.OpenRouterAPIKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
