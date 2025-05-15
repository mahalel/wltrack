package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/wltrak/internal/config"
	"github.com/user/wltrak/internal/database"
	"github.com/user/wltrak/internal/handlers"
)

// WLTrak - Weightlifting Tracking Application
//
// Environment variables:
// - TURSO_URL: Turso database URL (required)
// - TURSO_AUTH_TOKEN: Turso authentication token (required)
// - PORT: HTTP server port (default: 8080)
// - ENV: Runtime environment (default: development)

func main() {
	// Load configuration from OS environment variables
	cfg := config.Load()

	// Connect to Turso database using environment variables
	db, err := database.New(cfg.TursoURL, cfg.TursoAuthToken)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		_ = db.CloseDB()
	}()

	// Setup database tables
	if err := db.SetupDB(); err != nil {
		log.Fatalf("Failed to set up database tables: %v", err)
	}

	// Create a new HTTP server
	mux := http.NewServeMux()

	// static file path
	staticDir := "static"

	// Check if static directory exists and is accessible
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("Warning: Static directory '%s' not found, trying alternative paths", staticDir)
		// Try alternative paths
		for _, alt := range []string{"/static", "./static", "../static"} {
			if _, err := os.Stat(alt); err == nil {
				staticDir = alt
				break
			}
		}
	}

	// Static file serving
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	log.Printf("Serving static files from: %s", staticDir)

	// Routes
	// Home
	mux.HandleFunc("GET /", handlers.HomeHandler(db))

	// Exercises
	mux.HandleFunc("GET /exercises", handlers.ExercisesHandler(db))
	mux.HandleFunc("GET /exercises/new", handlers.NewExerciseHandler(db))
	mux.HandleFunc("GET /exercises/{id}/edit", handlers.EditExerciseHandler(db))
	mux.HandleFunc("GET /exercises/{id}", handlers.ExerciseDetailHandler(db))
	// The dedicated 1RM form route is no longer needed as we set 1RM directly in the exercise edit form
	// mux.HandleFunc("GET /exercises/{id}/1rm/new", handlers.NewOneRepMaxFormHandler())

	// Workouts
	mux.HandleFunc("GET /workouts", handlers.WorkoutsHandler(db))
	mux.HandleFunc("GET /workouts/new", handlers.NewWorkoutHandler(db))
	mux.HandleFunc("GET /workouts/{id}", handlers.WorkoutDetailHandler(db))
	mux.HandleFunc("GET /workouts/{id}/edit", handlers.EditWorkoutHandler(db))

	// API Routes
	mux.HandleFunc("POST /api/exercises", handlers.CreateExerciseHandler(db))
	mux.HandleFunc("PUT /api/exercises/{id}", handlers.UpdateExerciseHandler(db))
	mux.HandleFunc("DELETE /api/exercises/{id}", handlers.DeleteExerciseHandler(db))
	// We still need the API endpoint for backward compatibility
	mux.HandleFunc("POST /api/exercises/{id}/1rm", handlers.SaveOneRepMaxHandler(db))
	mux.HandleFunc("GET /api/exercises/{id}/1rm", handlers.GetOneRepMaxHandler(db))
	mux.HandleFunc("GET /api/exercises/{id}/history", handlers.GetExerciseHistoryHandler(db))

	mux.HandleFunc("POST /api/workouts", handlers.CreateWorkoutHandler(db))
	mux.HandleFunc("PUT /api/workouts/{id}", handlers.UpdateWorkoutHandler(db))
	mux.HandleFunc("DELETE /api/workouts/{id}", handlers.DeleteWorkoutHandler(db))
	mux.HandleFunc("GET /api/workouts/{id}/exercise-count", handlers.GetExerciseCountHandler(db))

	// Register NotFoundHandler as the default handler for any unmatched patterns
	mux.Handle("/", http.HandlerFunc(handlers.NotFoundHandler()))

	// Create a middleware for CORS headers and HX-Redirect handling
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, HX-Request, HX-Trigger")
			w.Header().Set("Access-Control-Expose-Headers", "HX-Redirect, HX-Refresh")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}

	// Create the server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: corsMiddleware(mux),
	}

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Server starting on port %s (environment: %s)...\n", cfg.Port, cfg.Environment)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Server shutting down...")
	if err := db.CloseDB(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	fmt.Println("Server stopped")
}
