package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	TursoURL       string
	TursoAuthToken string
	Port           string
	Environment    string

	// GitHub App Authentication
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
	AllowedGithubUsers []string

	// Authentication enabled flag
	AuthEnabled bool
}

// Load loads configuration from environment variables
func Load() Config {
	cfg := Config{
		TursoURL:       getEnv("TURSO_URL", ""),
		TursoAuthToken: getEnv("TURSO_AUTH_TOKEN", ""),
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENV", "development"),

		// GitHub App Authentication
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),
		AllowedGithubUsers: getStringSliceEnv("ALLOWED_GITHUB_USERS", nil),

		// Authentication enabled flag
		AuthEnabled: getBoolEnv("AUTH_ENABLED", false),
	}

	if cfg.TursoURL == "" {
		if cfg.Environment == "production" {
			log.Fatal("TURSO_URL not set")
		} else {
			log.Println("TURSO_URL not set, using local SQLite database for development")
			// Create a data directory if it doesn't exist
			if _, err := os.Stat("data"); os.IsNotExist(err) {
				err := os.Mkdir("data", 0755)
				if err != nil {
					log.Printf("Warning: Could not create data directory: %v", err)
					log.Println("Using in-memory database as fallback")
					cfg.TursoURL = "file::memory:?cache=shared"
				} else {
					cfg.TursoURL = "file:data/wltrack.db?cache=shared"
				}
			} else {
				cfg.TursoURL = "file:data/wltrack.db?cache=shared"
			}
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

// getBoolEnv gets a bool environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Warning: Could not parse %s as bool: %v", key, err)
		return defaultValue
	}

	return boolValue
}

// getStringSliceEnv gets a comma-separated string slice from environment or returns default
func getStringSliceEnv(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	values := strings.Split(value, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return values
}
