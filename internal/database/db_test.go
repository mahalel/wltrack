package database

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	// Create a unique in-memory SQLite database for each test that doesn't leave files
	// Each test gets a unique connection name to ensure isolation
	dbName := fmt.Sprintf("file:memdb_%s.db?mode=memory&cache=shared", t.Name())
	db, err := New(dbName, "")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Set up the database schema
	if err := db.SetupDB(); err != nil {
		t.Fatalf("Failed to set up database tables: %v", err)
	}

	// Register cleanup to remove the test database file
	t.Cleanup(func() {
		// Close the database connection
		if err := db.CloseDB(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}

		// Extract filename from the connection string
		filename := strings.TrimPrefix(dbName, "file:")
		filename = strings.Split(filename, "?")[0]

		// Try to delete the file if it exists
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				t.Logf("Warning: Failed to remove test database file %s: %v", filename, err)
			}
		}
	})

	return db
}

func TestNewDatabaseConnection(t *testing.T) {
	// Use a unique in-memory database that doesn't leave files
	dbName := fmt.Sprintf("file:memdb_%s.db?mode=memory", t.Name())
	db, err := New(dbName, "")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Register cleanup to remove the test database file
	t.Cleanup(func() {
		if err := db.CloseDB(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}

		// Extract filename from the connection string
		filename := strings.TrimPrefix(dbName, "file:")
		filename = strings.Split(filename, "?")[0]

		// Try to delete the file if it exists
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				t.Logf("Warning: Failed to remove test database file %s: %v", filename, err)
			}
		}
	})

	if db == nil {
		t.Fatal("Expected a database connection but got nil")
	}
}

func TestExerciseCRUD(t *testing.T) {
	db := setupTestDB(t)

	// Test AddExercise
	name := "Squat"
	description := "Barbell back squat"
	id, err := db.AddExercise(name, description)
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}
	if id <= 0 {
		t.Fatalf("Expected positive ID for new exercise, got: %d", id)
	}

	// Test GetExercise
	exercise, err := db.GetExercise(id)
	if err != nil {
		t.Fatalf("Failed to get exercise: %v", err)
	}
	if exercise.Name != name {
		t.Errorf("Expected exercise name to be %q, got %q", name, exercise.Name)
	}
	if exercise.Description != description {
		t.Errorf("Expected exercise description to be %q, got %q", description, exercise.Description)
	}

	// Test UpdateExercise
	newName := "Front Squat"
	newDescription := "Barbell front squat"
	err = db.UpdateExercise(id, newName, newDescription)
	if err != nil {
		t.Fatalf("Failed to update exercise: %v", err)
	}

	// Verify update
	updatedExercise, err := db.GetExercise(id)
	if err != nil {
		t.Fatalf("Failed to get updated exercise: %v", err)
	}
	if updatedExercise.Name != newName {
		t.Errorf("Expected updated exercise name to be %q, got %q", newName, updatedExercise.Name)
	}
	if updatedExercise.Description != newDescription {
		t.Errorf("Expected updated exercise description to be %q, got %q", newDescription, updatedExercise.Description)
	}

	// Test GetAllExercises
	exercises, err := db.GetAllExercises()
	if err != nil {
		t.Fatalf("Failed to get all exercises: %v", err)
	}
	if len(exercises) != 1 {
		t.Errorf("Expected 1 exercise, got %d", len(exercises))
	}

	// Test DeleteExercise
	err = db.DeleteExercise(id)
	if err != nil {
		t.Fatalf("Failed to delete exercise: %v", err)
	}

	// Verify deletion
	exercises, err = db.GetAllExercises()
	if err != nil {
		t.Fatalf("Failed to get exercises after deletion: %v", err)
	}
	if len(exercises) != 0 {
		t.Errorf("Expected 0 exercises after deletion, got %d", len(exercises))
	}
}

func TestOneRepMaxFunctions(t *testing.T) {
	db := setupTestDB(t)

	// Create an exercise
	exerciseID, err := db.AddExercise("Deadlift", "Conventional deadlift")
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}

	// Test SaveOneRepMax
	weight := 225.5
	_, err = db.SaveOneRepMax(exerciseID, weight)
	if err != nil {
		t.Fatalf("Failed to save one rep max: %v", err)
	}

	// Test GetLatestOneRepMax
	orm, err := db.GetLatestOneRepMax(exerciseID)
	if err != nil {
		t.Fatalf("Failed to get latest one rep max: %v", err)
	}
	if orm.ExerciseID != exerciseID {
		t.Errorf("Expected exercise ID %d, got %d", exerciseID, orm.ExerciseID)
	}
	if orm.Weight != weight {
		t.Errorf("Expected weight %.1f, got %.1f", weight, orm.Weight)
	}
}

func TestWorkoutCRUD(t *testing.T) {
	db := setupTestDB(t)

	// Test AddWorkout
	today := time.Now().Format("2006-01-02")
	notes := "Test workout"

	workoutID, err := db.AddWorkout(today, notes)
	if err != nil {
		t.Fatalf("Failed to add workout: %v", err)
	}
	if workoutID <= 0 {
		t.Fatalf("Expected positive ID for new workout, got: %d", workoutID)
	}

	// Create an exercise to add to the workout
	exerciseID, err := db.AddExercise("Bench Press", "Flat barbell bench press")
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}

	// Test AddWorkoutExercise
	workoutExerciseID, err := db.AddWorkoutExercise(workoutID, exerciseID, "Testing sets")
	if err != nil {
		t.Fatalf("Failed to add workout exercise: %v", err)
	}

	// Test AddSet
	setID, err := db.AddSet(workoutExerciseID, 5, 135.0, 75.0, 1, "range1")
	if err != nil {
		t.Fatalf("Failed to add set: %v", err)
	}
	if setID <= 0 {
		t.Errorf("Expected positive ID for new set, got: %d", setID)
	}

	// Test GetWorkoutDetails
	workoutDetails, err := db.GetWorkoutDetails(workoutID)
	if err != nil {
		t.Fatalf("Failed to get workout details: %v", err)
	}
	if workoutDetails.Workout.ID != workoutID {
		t.Errorf("Expected workout ID %d, got %d", workoutID, workoutDetails.Workout.ID)
	}
	if len(workoutDetails.Exercises) != 1 {
		t.Errorf("Expected 1 exercise in workout, got %d", len(workoutDetails.Exercises))
	}
	if len(workoutDetails.Exercises[0].Sets) != 1 {
		t.Errorf("Expected 1 set in exercise, got %d", len(workoutDetails.Exercises[0].Sets))
	}

	// Test UpdateWorkout
	newNotes := "Updated workout notes"
	err = db.UpdateWorkout(workoutID, today, newNotes)
	if err != nil {
		t.Fatalf("Failed to update workout: %v", err)
	}

	// Test DeleteWorkout
	err = db.DeleteWorkout(workoutID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	// Verify deletion by getting all workouts
	workouts, err := db.GetWorkouts()
	if err != nil {
		t.Fatalf("Failed to get workouts after deletion: %v", err)
	}
	if len(workouts) != 0 {
		t.Errorf("Expected 0 workouts after deletion, got %d", len(workouts))
	}
}

func TestExerciseHistory(t *testing.T) {
	db := setupTestDB(t)

	// Create an exercise
	exerciseID, err := db.AddExercise("Pull-up", "Bodyweight pull-up")
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}

	// Create two workouts on different dates
	workout1ID, err := db.AddWorkout("2023-01-01", "First workout")
	if err != nil {
		t.Fatalf("Failed to add first workout: %v", err)
	}

	workout2ID, err := db.AddWorkout("2023-01-08", "Second workout")
	if err != nil {
		t.Fatalf("Failed to add second workout: %v", err)
	}

	// Add the exercise to both workouts
	we1ID, err := db.AddWorkoutExercise(workout1ID, exerciseID, "First time")
	if err != nil {
		t.Fatalf("Failed to add exercise to first workout: %v", err)
	}

	we2ID, err := db.AddWorkoutExercise(workout2ID, exerciseID, "Second time")
	if err != nil {
		t.Fatalf("Failed to add exercise to second workout: %v", err)
	}

	// Add sets to both workout exercises
	_, err = db.AddSet(we1ID, 5, 0.0, 0.0, 1, "range1")
	if err != nil {
		t.Fatalf("Failed to add set to first workout: %v", err)
	}

	_, err = db.AddSet(we2ID, 8, 0.0, 0.0, 1, "range1")
	if err != nil {
		t.Fatalf("Failed to add set to second workout: %v", err)
	}

	// Test GetExerciseHistory
	history, err := db.GetExerciseHistory(exerciseID)
	if err != nil {
		t.Fatalf("Failed to get exercise history: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 entries in exercise history, got %d", len(history))
	}

	// The history should be in descending order by date
	if history[0].WorkoutID != workout2ID {
		t.Errorf("Expected first history entry to be from workout %d, got %d", workout2ID, history[0].WorkoutID)
	}

	// Test GetExerciseSetsForWorkout
	sets, err := db.GetExerciseSetsForWorkout(we1ID)
	if err != nil {
		t.Fatalf("Failed to get exercise sets: %v", err)
	}

	if len(sets) != 1 {
		t.Errorf("Expected 1 set for workout exercise, got %d", len(sets))
	}
}
