package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/user/wltrak/internal/database"
	"github.com/user/wltrak/internal/templates"
)

// ExercisesHandler handles the GET /exercises route
func ExercisesHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exercises, err := db.GetAllExercises()
		if err != nil {
			http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
			return
		}
		
		if err := templates.ExerciseList(exercises).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// NewExerciseHandler handles the GET /exercises/new route
func NewExerciseHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExerciseForm(nil).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// EditExerciseHandler handles the GET /exercises/:id/edit route
func EditExerciseHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
			return
		}

		exercise, err := db.GetExercise(id)
		if err != nil {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}

		// Get the current 1RM value if available
		oneRepMax, err := db.GetLatestOneRepMax(id)
		if err == nil {
			// Add the 1RM value to the form
			exercise.OneRepMax = oneRepMax.Weight
		}

		if err := templates.ExerciseForm(&exercise).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// ExerciseDetailHandler handles the GET /exercises/:id route
func ExerciseDetailHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
			return
		}

		exercise, err := db.GetExercise(id)
		if err != nil {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}

		oneRepMax, err := db.GetLatestOneRepMax(id)
		if err != nil {
			// No OneRepMax found, that's okay
			if err := templates.ExerciseDetail(exercise, nil).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		if err := templates.ExerciseDetail(exercise, &oneRepMax).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}



// CreateExerciseHandler handles the POST /api/exercises route
func CreateExerciseHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			if err := templates.FormError("Failed to parse form data").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")

		if name == "" {
			if err := templates.FormError("Exercise name is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		_, err = db.AddExercise(name, description)
		if err != nil {
			if err := templates.FormError("Failed to create exercise: " + err.Error()).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Add("HX-Redirect", "/exercises")
		if err := templates.FormSuccess("Exercise created successfully! Redirecting...").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

// UpdateExerciseHandler handles the PUT /api/exercises/:id route
func UpdateExerciseHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			if err := templates.FormError("Invalid exercise ID").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Set headers before processing the form to avoid "Exercise name is required" flash
		w.Header().Set("HX-Redirect", "/exercises/"+idStr)

		err = r.ParseForm()
		if err != nil {
			if err := templates.FormError("Failed to parse form data").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")

		if name == "" {
			if err := templates.FormError("Exercise name is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Update the exercise in the database
		err = db.UpdateExercise(id, name, description)
		if err != nil {
			if err := templates.FormError("Failed to update exercise: " + err.Error()).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Check if one_rep_max was provided and save it
		oneRepMaxStr := r.FormValue("one_rep_max")
		if oneRepMaxStr != "" {
			oneRepMax, err := strconv.ParseFloat(oneRepMaxStr, 64)
			if err != nil {
				if err := templates.FormError("Invalid one rep max value").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}

			if oneRepMax > 0 {
				_, err = db.SaveOneRepMax(id, oneRepMax)
				if err != nil {
					if err := templates.FormError("Failed to save one rep max: " + err.Error()).Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}
			}
		}
		
		// Instead of just setting HX-Redirect header, use a full page refresh
		w.Header().Set("HX-Refresh", "true")
		
		// Render success message
		if err := templates.FormSuccess("Exercise updated successfully! Page will refresh...").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		
		// Add JavaScript fallback for browsers without HTMX
		fmt.Fprintf(w, "<script>setTimeout(function() { window.location.href = '/exercises/%s'; }, 800);</script>", idStr)
	}
}

// DeleteExerciseHandler handles the DELETE /api/exercises/:id route
func DeleteExerciseHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
			return
		}

		err = db.DeleteExercise(id)
		if err != nil {
			http.Error(w, "Failed to delete exercise: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Return empty response for the HTMX swap
		w.WriteHeader(http.StatusOK)
	}
}



// GetOneRepMaxHandler handles the GET /api/exercises/:id/1rm route
func GetOneRepMaxHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
			return
		}

		oneRepMax, err := db.GetLatestOneRepMax(id)
		if err != nil {
			if err := templates.OneRepMaxValue(0, "").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		formattedDate := oneRepMax.Date.Format("Jan 2, 2006")
		if err := templates.OneRepMaxValue(oneRepMax.Weight, formattedDate).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}