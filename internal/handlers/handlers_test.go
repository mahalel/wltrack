package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mahalel/wltrack/internal/database"
)

func setupTestServer(t *testing.T) (*database.DB, func()) {
	// Create a test database that won't leave files behind
	dbName := fmt.Sprintf("file:test_handlers_%s.db?mode=memory", t.Name())
	db, err := database.New(dbName, "")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Setup database schema
	if err := db.SetupDB(); err != nil {
		t.Fatalf("Failed to set up database tables: %v", err)
	}

	// Extract filename from the connection string
	filename := strings.TrimPrefix(dbName, "file:")
	filename = strings.Split(filename, "?")[0]

	// Return cleanup function
	cleanup := func() {
		if err := db.CloseDB(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}

		// Try to delete the file if it exists
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				t.Logf("Warning: Failed to remove test database file %s: %v", filename, err)
			}
		}
	}

	return db, cleanup
}

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		handler    func() http.HandlerFunc
		wantStatus int
	}{
		{
			name:       "Health Liveness",
			endpoint:   "/health/live",
			handler:    HealthLiveHandler,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Health Readiness",
			endpoint:   "/health/ready",
			handler:    HealthReadyHandler,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := tt.handler()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestHomeHandler(t *testing.T) {
	db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := HomeHandler(db)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestNotFoundHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/nonexistent-page", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := NotFoundHandler()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

func TestExercisesHandler(t *testing.T) {
	db, cleanup := setupTestServer(t)
	defer cleanup()

	// Add a sample exercise to the database
	_, err := db.AddExercise("Test Exercise", "Test Description")
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}

	// Create test request
	req, err := http.NewRequest("GET", "/exercises", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := ExercisesHandler(db)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestWorkoutsHandler(t *testing.T) {
	db, cleanup := setupTestServer(t)
	defer cleanup()

	// Add a sample workout to the database
	_, err := db.AddWorkout("2023-05-15", "Test Workout Notes")
	if err != nil {
		t.Fatalf("Failed to add workout: %v", err)
	}

	// Create test request
	req, err := http.NewRequest("GET", "/workouts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := WorkoutsHandler(db)
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
