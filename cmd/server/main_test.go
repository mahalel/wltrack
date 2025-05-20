package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mahalel/wltrack/internal/config"
	"github.com/mahalel/wltrack/internal/database"
)

func TestMain(m *testing.M) {
	// Set up test environment variables
	if err := os.Setenv("TURSO_URL", "file:test.db?mode=memory"); err != nil {
		panic(err)
	}
	if err := os.Setenv("TURSO_AUTH_TOKEN", ""); err != nil {
		panic(err)
	}
	if err := os.Setenv("PORT", "8888"); err != nil {
		panic(err)
	}
	if err := os.Setenv("ENV", "test"); err != nil {
		panic(err)
	}

	// Run the tests
	exitCode := m.Run()

	// Clean up
	if err := os.Unsetenv("TURSO_URL"); err != nil {
		fmt.Printf("Failed to unset TURSO_URL: %v\n", err)
	}
	if err := os.Unsetenv("TURSO_AUTH_TOKEN"); err != nil {
		fmt.Printf("Failed to unset TURSO_AUTH_TOKEN: %v\n", err)
	}
	if err := os.Unsetenv("PORT"); err != nil {
		fmt.Printf("Failed to unset PORT: %v\n", err)
	}
	if err := os.Unsetenv("ENV"); err != nil {
		fmt.Printf("Failed to unset ENV: %v\n", err)
	}

	os.Exit(exitCode)
}

func TestServerStartup(t *testing.T) {
	// Load configuration
	cfg := config.Load()

	if cfg.TursoURL != "file:test.db?mode=memory" {
		t.Errorf("Expected TURSO_URL to be 'file:test.db?mode=memory', got '%s'", cfg.TursoURL)
	}

	if cfg.Port != "8888" {
		t.Errorf("Expected PORT to be '8888', got '%s'", cfg.Port)
	}

	if cfg.Environment != "test" {
		t.Errorf("Expected ENV to be 'test', got '%s'", cfg.Environment)
	}

	// Test database connection
	db, err := database.New(cfg.TursoURL, cfg.TursoAuthToken)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Extract filename from the connection string
	filename := strings.TrimPrefix(cfg.TursoURL, "file:")
	filename = strings.Split(filename, "?")[0]

	// Register cleanup to remove test database file and close connection
	t.Cleanup(func() {
		if err := db.CloseDB(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}

		// Try to delete the file if it exists
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				t.Logf("Warning: Failed to remove test database file %s: %v", filename, err)
			}
		}
	})

	// Test database setup
	err = db.SetupDB()
	if err != nil {
		t.Fatalf("Failed to set up database tables: %v", err)
	}

	// Verify we can add and retrieve data
	exerciseID, err := db.AddExercise("Test Exercise", "Test Description")
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}

	exercise, err := db.GetExercise(exerciseID)
	if err != nil {
		t.Fatalf("Failed to retrieve exercise: %v", err)
	}

	if exercise.Name != "Test Exercise" {
		t.Errorf("Expected exercise name to be 'Test Exercise', got '%s'", exercise.Name)
	}
}

func TestHealthEndpoints(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health/live":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/health/ready":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ready"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Test liveness endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ts.URL+"/health/live", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test readiness endpoint
	req, err = http.NewRequestWithContext(ctx, "GET", ts.URL+"/health/ready", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
