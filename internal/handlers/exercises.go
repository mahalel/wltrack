package handlers

import (
	"net/http"
	"strconv"

	"github.com/mahalel/wltrack/internal/database"
	"github.com/mahalel/wltrack/internal/templates"
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

		// Get 1RM history
		oneRepMaxHistory, err := db.GetLatestOneRepMax(id)
		if err != nil {
			// No OneRepMax history found, that's okay - we'll use the current_1rm from the exercise
			if err := templates.ExerciseDetail(exercise, nil).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		if err := templates.ExerciseDetail(exercise, &oneRepMaxHistory).Render(r.Context(), w); err != nil {
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
		notes := r.FormValue("notes")

		// Parse 1RM if provided
		var current1RM float64
		current1RMStr := r.FormValue("current_1rm")
		if current1RMStr != "" {
			current1RM, err = strconv.ParseFloat(current1RMStr, 64)
			if err != nil {
				if err := templates.FormError("Invalid one rep max value").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}
		}

		if name == "" {
			if err := templates.FormError("Exercise name is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		_, err = db.AddExercise(name, current1RM, notes)
		if err != nil {
			if err := templates.FormError("Failed to create exercise: "+err.Error()).Render(r.Context(), w); err != nil {
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

		err = r.ParseForm()
		if err != nil {
			if err := templates.FormError("Failed to parse form data").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		name := r.FormValue("name")
		notes := r.FormValue("notes")

		// Parse 1RM if provided
		var current1RM float64
		current1RMStr := r.FormValue("current_1rm")
		if current1RMStr != "" {
			current1RM, err = strconv.ParseFloat(current1RMStr, 64)
			if err != nil {
				if err := templates.FormError("Invalid one rep max value").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}
		}

		if name == "" {
			if err := templates.FormError("Exercise name is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Update the exercise in the database
		err = db.UpdateExercise(id, name, current1RM, notes)
		if err != nil {
			if err := templates.FormError("Failed to update exercise: "+err.Error()).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// If current1RM was provided, save it to the 1RM history
		if current1RMStr != "" && current1RM > 0 {
			_, err = db.SaveOneRepMax(id, current1RM)
			if err != nil {
				if err := templates.FormError("Failed to save one rep max: "+err.Error()).Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}
		}

		w.Header().Set("HX-Redirect", "/exercises/"+idStr)
		if err := templates.FormSuccess("Exercise updated successfully! Redirecting...").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
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

		exercise, err := db.GetExercise(id)
		if err != nil {
			http.Error(w, "Exercise not found", http.StatusBadRequest)
			return
		}

		// If the exercise has a current_1rm value, use it
		if exercise.Current1RM > 0 {
			// Try to get the latest history record for date information
			oneRepMax, err := db.GetLatestOneRepMax(id)
			var formattedDate string
			if err == nil {
				formattedDate = oneRepMax.CreatedAt.Format("Jan 2, 2006")
			}

			if err := templates.OneRepMaxValue(exercise.Current1RM, formattedDate).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// No current 1RM
		if err := templates.OneRepMaxValue(0, "").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}