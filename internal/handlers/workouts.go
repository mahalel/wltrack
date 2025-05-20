package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/mahalel/wltrack/internal/database"
	"github.com/mahalel/wltrack/internal/templates"
)

// WorkoutsHandler handles the GET /workouts route
func WorkoutsHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workouts, err := db.GetWorkouts()
		if err != nil {
			http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
			return
		}

		if err := templates.WorkoutList(workouts).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// NewWorkoutHandler handles the GET /workouts/new route
func NewWorkoutHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exercises, err := db.GetAllExercises()
		if err != nil {
			http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
			return
		}

		if err := templates.WorkoutForm(exercises).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// WorkoutDetailHandler handles the GET /workouts/:id route
func WorkoutDetailHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}

		workout, err := db.GetWorkoutDetails(id)
		if err != nil {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}

		if err := templates.WorkoutDetail(workout).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// EditWorkoutHandler handles the GET /workouts/:id/edit route
func EditWorkoutHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}

		workout, err := db.GetWorkoutDetails(id)
		if err != nil {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}

		// Get all exercises for the dropdown
		exercises, err := db.GetAllExercises()
		if err != nil {
			http.Error(w, "Failed to fetch exercises", http.StatusInternalServerError)
			return
		}

		if err := templates.WorkoutEditForm(workout, exercises).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}

// CreateWorkoutHandler handles the POST /api/workouts route
func CreateWorkoutHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			if err := templates.FormError("Failed to parse form data").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		date := r.FormValue("date")
		if date == "" {
			if err := templates.FormError("Date is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		notes := r.FormValue("notes")

		// Create the workout
		workoutID, err := db.AddWorkout(date, notes)
		if err != nil {
			if err := templates.FormError("Failed to create workout: "+err.Error()).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Get exercise IDs
		exerciseIDs := r.Form["exercise_id[]"]
		exerciseNotes := r.Form["exercise_notes[]"]

		// Get set range values
		setStarts := r.Form["set_start[]"]
		setEnds := r.Form["set_end[]"]

		// Get set details
		reps := r.Form["reps[]"]
		weights := r.Form["weight[]"]
		percentages := r.Form["percentage[]"]

		// Validate we have at least one exercise
		if len(exerciseIDs) == 0 {
			if err := templates.FormError("At least one exercise is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Process each exercise
		for i, exerciseIDStr := range exerciseIDs {
			// Skip empty exercise selections
			if exerciseIDStr == "" {
				continue
			}

			exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
			if err != nil {
				if err := templates.FormError("Invalid exercise ID").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}

			// Get the notes for this exercise (if available)
			var exerciseNote string
			if i < len(exerciseNotes) {
				exerciseNote = exerciseNotes[i]
			}

			// Add the exercise to the workout
			workoutExerciseID, err := db.AddWorkoutExercise(workoutID, exerciseID, exerciseNote)
			if err != nil {
				if err := templates.FormError("Failed to add exercise to workout: "+err.Error()).Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}

			// Process each set range for this exercise
			totalSets := 0
			var maxWeight float64 = 0
			var repsAtMaxWeight = 0

			// For each set range in the current exercise
			for rangeIdx := 0; rangeIdx < len(setStarts) && rangeIdx < len(setEnds) && rangeIdx < len(reps); rangeIdx++ {
				// Parse set range
				startSet, err := strconv.Atoi(setStarts[rangeIdx])
				if err != nil || startSet < 1 {
					if err := templates.FormError("Invalid set range start").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				endSet, err := strconv.Atoi(setEnds[rangeIdx])
				if err != nil || endSet < startSet {
					if err := templates.FormError("Invalid set range end").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				// Parse rep count for this range
				repCount, err := strconv.Atoi(reps[rangeIdx])
				if err != nil || repCount < 1 {
					if err := templates.FormError("Invalid rep count").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				// Parse weight for this range
				weight, err := strconv.ParseFloat(weights[rangeIdx], 64)
				if err != nil || weight < 0 {
					if err := templates.FormError("Invalid weight").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				// Parse percentage (optional) for this range
				var percentage float64
				if rangeIdx < len(percentages) && percentages[rangeIdx] != "" {
					percentage, err = strconv.ParseFloat(percentages[rangeIdx], 64)
					if err != nil || percentage < 0 || percentage > 100 {
						if err := templates.FormError("Invalid percentage of 1RM").Render(r.Context(), w); err != nil {
							http.Error(w, "Error rendering template", http.StatusInternalServerError)
						}
						return
					}
				}

				// Create each set in the range
				for setOrder := startSet; setOrder <= endSet; setOrder++ {
					_, err = db.AddSet(workoutExerciseID, repCount, weight, percentage, setOrder)
					if err != nil {
						if err := templates.FormError("Failed to add set: "+err.Error()).Render(r.Context(), w); err != nil {
							http.Error(w, "Error rendering template", http.StatusInternalServerError)
						}
						return
					}
					totalSets++
				}

				// Track the heaviest weight used in this workout for 1RM calculation
				// Lower reps at the same weight is better for 1RM calculation since it's closer to max strength
				if weight > maxWeight || (weight == maxWeight && repCount < repsAtMaxWeight) {
					maxWeight = weight
					repsAtMaxWeight = repCount
				}
			}

			// Make sure at least one set was added
			if totalSets == 0 {
				if err := templates.FormError("Each exercise must have at least one set").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}
		}

		w.Header().Add("HX-Redirect", "/workouts/"+strconv.FormatInt(workoutID, 10))
		if err := templates.FormSuccess("Workout recorded successfully! Redirecting...").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

// DeleteWorkoutHandler handles the DELETE /api/workouts/:id route
func DeleteWorkoutHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}

		err = db.DeleteWorkout(id)
		if err != nil {
			http.Error(w, "Failed to delete workout: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return OK for now
		w.Header().Add("HX-Redirect", "/workouts")
		w.WriteHeader(http.StatusOK)
	}
}

// UpdateWorkoutHandler handles the PUT /api/workouts/:id route
func UpdateWorkoutHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			if err := templates.FormError("Invalid workout ID").Render(r.Context(), w); err != nil {
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

		// Set HTMX headers early to prevent form error flashes
		w.Header().Set("HX-Refresh", "true")

		date := r.FormValue("date")
		if date == "" {
			if err := templates.FormError("Date is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		notes := r.FormValue("notes")

		// Update the workout
		err = db.UpdateWorkout(id, date, notes)
		if err != nil {
			if err := templates.FormError("Failed to update workout: "+err.Error()).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Delete existing workout exercises and sets
		err = db.DeleteWorkoutExercisesAndSets(id)
		if err != nil {
			if err := templates.FormError("Failed to update workout exercises: "+err.Error()).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Get exercise IDs
		exerciseIDs := r.Form["exercise_id[]"]
		exerciseNotes := r.Form["exercise_notes[]"]

		// Get set range values
		setStarts := r.Form["set_start[]"]
		setEnds := r.Form["set_end[]"]

		// Get set details
		reps := r.Form["reps[]"]
		weights := r.Form["weight[]"]
		percentages := r.Form["percentage[]"]

		// Validate we have at least one exercise
		if len(exerciseIDs) == 0 {
			if err := templates.FormError("At least one exercise is required").Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// Process each exercise
		for i, exerciseIDStr := range exerciseIDs {
			// Skip empty exercise selections
			if exerciseIDStr == "" {
				continue
			}

			exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
			if err != nil {
				if err := templates.FormError("Invalid exercise ID").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}

			// Get the notes for this exercise (if available)
			var exerciseNote string
			if i < len(exerciseNotes) {
				exerciseNote = exerciseNotes[i]
			}

			// Add the exercise to the workout
			workoutExerciseID, err := db.AddWorkoutExercise(id, exerciseID, exerciseNote)
			if err != nil {
				if err := templates.FormError("Failed to add exercise to workout: "+err.Error()).Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}

			// Process each set range for this exercise
			totalSets := 0
			var maxWeight float64 = 0
			var repsAtMaxWeight = 0

			// For each set range in the current exercise
			for rangeIdx := 0; rangeIdx < len(setStarts) && rangeIdx < len(setEnds) && rangeIdx < len(reps); rangeIdx++ {
				// Parse set range
				startSet, err := strconv.Atoi(setStarts[rangeIdx])
				if err != nil || startSet < 1 {
					if err := templates.FormError("Invalid set range start").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				endSet, err := strconv.Atoi(setEnds[rangeIdx])
				if err != nil || endSet < startSet {
					if err := templates.FormError("Invalid set range end").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				// Parse rep count for this range
				repCount, err := strconv.Atoi(reps[rangeIdx])
				if err != nil || repCount < 1 {
					if err := templates.FormError("Invalid rep count").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				// Parse weight for this range
				weight, err := strconv.ParseFloat(weights[rangeIdx], 64)
				if err != nil || weight < 0 {
					if err := templates.FormError("Invalid weight").Render(r.Context(), w); err != nil {
						http.Error(w, "Error rendering template", http.StatusInternalServerError)
					}
					return
				}

				// Parse percentage (optional) for this range
				var percentage float64
				if rangeIdx < len(percentages) && percentages[rangeIdx] != "" {
					percentage, err = strconv.ParseFloat(percentages[rangeIdx], 64)
					if err != nil || percentage < 0 || percentage > 100 {
						if err := templates.FormError("Invalid percentage of 1RM").Render(r.Context(), w); err != nil {
							http.Error(w, "Error rendering template", http.StatusInternalServerError)
						}
						return
					}
				}

				// Create each set in the range
				for setOrder := startSet; setOrder <= endSet; setOrder++ {
					_, err = db.AddSet(workoutExerciseID, repCount, weight, percentage, setOrder)
					if err != nil {
						if err := templates.FormError("Failed to add set: "+err.Error()).Render(r.Context(), w); err != nil {
							http.Error(w, "Error rendering template", http.StatusInternalServerError)
						}
						return
					}
					totalSets++
				}

				// Track the heaviest weight used in this workout for 1RM calculation
				// Lower reps at the same weight is better for 1RM calculation since it's closer to max strength
				if weight > maxWeight || (weight == maxWeight && repCount < repsAtMaxWeight) {
					maxWeight = weight
					repsAtMaxWeight = repCount
				}
			}

			// After adding all sets, calculate 1RM based on the heaviest set
			if maxWeight > 0 {
				updateEstimated1RM(db, exerciseID, repsAtMaxWeight, maxWeight)
			}

			// Make sure at least one set was added
			if totalSets == 0 {
				if err := templates.FormError("Each exercise must have at least one set").Render(r.Context(), w); err != nil {
					http.Error(w, "Error rendering template", http.StatusInternalServerError)
				}
				return
			}
		}

		if err := templates.FormSuccess("Workout updated successfully! Page will refresh...").Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
		if _, err := fmt.Fprintf(w, "<script>setTimeout(function() { window.location.href = '/workouts/%d'; }, 800);</script>", id); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

// GetExerciseCountHandler handles the GET /api/workouts/:id/exercise-count route
func GetExerciseCountHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}

		workout, err := db.GetWorkoutDetails(id)
		if err != nil {
			if err := templates.ExerciseCount(0).Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		if err := templates.ExerciseCount(len(workout.Exercises)).Render(r.Context(), w); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

// GetExerciseHistoryHandler handles the GET /api/exercises/:id/history route
func GetExerciseHistoryHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode([]map[string]interface{}{}) // Return empty array instead of error
			return
		}

		history, err := db.GetExerciseHistory(id)
		if err != nil {
			// Return empty array instead of error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}

		// For each workout exercise, fetch the sets
		type ChartData struct {
			Date   string    `json:"date"`
			Weight []float64 `json:"weights"`
			Reps   []int     `json:"reps"`
		}

		var chartData []ChartData

		for _, workoutExercise := range history {
			sets, err := db.GetExerciseSetsForWorkout(workoutExercise.ID)
			if err != nil {
				// Skip this workout but log the error
				log.Printf("Error fetching sets for workout exercise %d: %v", workoutExercise.ID, err)
				continue
			}

			// Initialize with empty arrays (not nil)
			weights := []float64{}
			reps := []int{}

			for _, set := range sets {
				weights = append(weights, set.Weight)
				reps = append(reps, set.Reps)
			}

			// Parse the date from the query result (we already get the date in GetExerciseHistory)
			// Format date as "Week W, MMM" (e.g., "Week 32, Aug")
			var dateStr string

			// Get all workouts to find the one for this exercise
			allWorkouts, err := db.GetWorkouts()
			if err != nil {
				// Fallback if we can't get the workouts
				dateStr = strconv.FormatInt(workoutExercise.WorkoutID, 10)
			} else {
				// Find the matching workout
				var workoutDate time.Time
				for _, w := range allWorkouts {
					if w.ID == workoutExercise.WorkoutID {
						workoutDate = w.Date
						break
					}
				}

				if !workoutDate.IsZero() {
					// Format as week and month
					_, week := workoutDate.ISOWeek()
					monthName := workoutDate.Format("Jan")
					dateStr = fmt.Sprintf("Week %d, %s", week, monthName)
				} else {
					// Fallback to ID if we couldn't find the workout
					dateStr = strconv.FormatInt(workoutExercise.WorkoutID, 10)
				}
			}

			chartData = append(chartData, ChartData{
				Date:   dateStr,
				Weight: weights,
				Reps:   reps,
			})
		}

		// Ensure we always send a valid array, even if empty
		if chartData == nil {
			chartData = []ChartData{}
		}

		// Respond with JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(chartData); err != nil {
			// If encoding fails, return an empty array
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]map[string]interface{}{})
			return
		}
	}
}

// Helper function to calculate and possibly update the estimated 1RM
func updateEstimated1RM(db *database.DB, exerciseID int64, reps int, weight float64) {
	// Don't calculate for very high reps (less accurate)
	if reps > 12 {
		return
	}

	// Several formulas for calculating 1RM:
	var estimated1RM float64

	if reps == 1 {
		// If it's already a 1RM attempt, just use the weight directly
		estimated1RM = weight
	} else {
		// Brzycki formula: weight * (36 / (37 - reps)) - more accurate for lower reps
		brzeycki := weight * (36.0 / (37.0 - float64(reps)))

		// Epley formula: weight * (1 + 0.0333 * reps) - good middle-ground formula
		epley := weight * (1.0 + 0.0333*float64(reps))

		// Lombardi formula: weight * (reps ^ 0.1) - more conservative estimate
		lombardi := weight * math.Pow(float64(reps), 0.1)

		// Weight the formulas differently based on rep range
		if reps <= 3 {
			// For low reps (1-3), Brzycki tends to be more accurate
			estimated1RM = (brzeycki * 0.6) + (epley * 0.3) + (lombardi * 0.1)
		} else if reps <= 6 {
			// For medium reps (4-6), balanced approach
			estimated1RM = (brzeycki * 0.4) + (epley * 0.4) + (lombardi * 0.2)
		} else {
			// For higher reps (7-12), Epley tends to be more accurate
			estimated1RM = (brzeycki * 0.3) + (epley * 0.5) + (lombardi * 0.2)
		}
	}

	// Round to the nearest 0.5kg
	estimated1RM = math.Round(estimated1RM*2) / 2

	// Check if this is higher than the current 1RM
	currentOneRM, err := db.GetLatestOneRepMax(exerciseID)
	if err != nil {
		// If there's no current 1RM, just save this one
		_, _ = db.SaveOneRepMax(exerciseID, estimated1RM)
		return
	}

	// Always save a new 1RM in the following cases:
	if estimated1RM > currentOneRM.Weight {
		// The calculated 1RM is higher than the current record
		_, _ = db.SaveOneRepMax(exerciseID, estimated1RM)
	} else if reps == 1 && weight > currentOneRM.Weight*0.9 {
		// This is a heavy single at 90%+ of current 1RM, worth recording
		_, _ = db.SaveOneRepMax(exerciseID, estimated1RM)
	} else if reps <= 3 && weight > currentOneRM.Weight*0.85 {
		// Using 85%+ of 1RM for low reps often indicates progress
		_, _ = db.SaveOneRepMax(exerciseID, estimated1RM)
	} else if reps <= 5 && weight > currentOneRM.Weight*0.8 {
		// Decent volume at 80%+ can indicate strength improvements
		_, _ = db.SaveOneRepMax(exerciseID, estimated1RM)
	}
}
