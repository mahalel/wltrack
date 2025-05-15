package config

import (
	"log"
	"os"
)

// Config holds the application configuration
type Config struct {
	TursoURL      string
	TursoAuthToken string
	Port          string
	Environment   string
}

// Load loads configuration from environment variables
func Load() Config {
	cfg := Config{
		TursoURL:      getEnv("TURSO_URL", ""),
		TursoAuthToken: getEnv("TURSO_AUTH_TOKEN", ""),
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENV", "development"),
	}

	if cfg.TursoURL == "" {
		if cfg.Environment == "production" {
			log.Fatal("TURSO_URL not set")
		} else {
			log.Println("TURSO_URL not set, using in-memory SQLite for development")
			cfg.TursoURL = "file::memory:?cache=shared"
		}
	}

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}