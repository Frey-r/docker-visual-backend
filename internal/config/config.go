package config

import (
	"os"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port           string
	AllowedOrigins []string
	WorkspacePath  string
	APIKey         string
	LogLevel       string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		WorkspacePath:  getEnv("WORKSPACE_PATH", "./workspaces"),
		APIKey:         getEnv("API_KEY", ""),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		return strings.Split(v, ",")
	}
	return fallback
}
