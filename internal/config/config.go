package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port           string
	AllowedOrigins []string
	WorkspacePath  string
	APIKey         string
	LogLevel       string
	// Auth settings
	JWTSecret      string
	JWTExpireHours int
	DBPath         string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		AllowedOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		WorkspacePath:  getEnv("WORKSPACE_PATH", "./workspaces"),
		APIKey:         getEnv("API_KEY", ""),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production-use-random-string"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 72),
		DBPath:         getEnv("DB_PATH", "./data/docker-visual.db"),
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

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
