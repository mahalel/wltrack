package handlers

import (
	"net/http"

	"github.com/user/wltrak/internal/database"
	"github.com/user/wltrak/internal/templates"
)

// HomeHandler handles the GET / route
func HomeHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get recent workouts (last 5)
		workouts, err := db.GetWorkouts()
		if err != nil {
			http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
			return
		}

		// Limit to 5 recent workouts
		recentWorkouts := workouts
		if len(workouts) > 5 {
			recentWorkouts = workouts[:5]
		}

		// Get all exercises
		exercises, err := db.GetAllExercises()
		if err != nil {
			http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
			return
		}

		if err := templates.Home(recentWorkouts, exercises).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 - Page not found"))
	}
}
