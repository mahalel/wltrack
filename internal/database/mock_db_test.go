package database

import (
	"testing"
)

func TestMockDBExerciseCRUD(t *testing.T) {
	db := NewMockDB()

	// Test AddExercise
	name := "Overhead Press"
	description := "Barbell overhead press"
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
	newName := "Push Press"
	newDescription := "Barbell push press"
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

func TestMockDBWorkflowWithRelationships(t *testing.T) {
	db := NewMockDB()

	// 1. Add exercise
	exerciseID, err := db.AddExercise("Squat", "Barbell back squat")
	if err != nil {
		t.Fatalf("Failed to add exercise: %v", err)
	}

	// 2. Save 1RM for exercise
	weight := 225.0
	_, err = db.SaveOneRepMax(exerciseID, weight)
	if err != nil {
		t.Fatalf("Failed to save one rep max: %v", err)
	}

	// 3. Get latest 1RM
	orm, err := db.GetLatestOneRepMax(exerciseID)
	if err != nil {
		t.Fatalf("Failed to get latest one rep max: %v", err)
	}
	if orm.Weight != weight {
		t.Errorf("Expected 1RM weight to be %.1f, got %.1f", weight, orm.Weight)
	}

	// 4. Add workout
	workoutID, err := db.AddWorkout("2023-05-01", "Leg day")
	if err != nil {
		t.Fatalf("Failed to add workout: %v", err)
	}

	// 5. Add exercise to workout
	workoutExerciseID, err := db.AddWorkoutExercise(workoutID, exerciseID, "Heavy day")
	if err != nil {
		t.Fatalf("Failed to add workout exercise: %v", err)
	}

	// 6. Add sets to workout exercise
	_, err = db.AddSet(workoutExerciseID, 5, 185.0, 80.0, 1)
	if err != nil {
		t.Fatalf("Failed to add first set: %v", err)
	}
	_, err = db.AddSet(workoutExerciseID, 5, 205.0, 90.0, 2)
	if err != nil {
		t.Fatalf("Failed to add second set: %v", err)
	}

	// 7. Get workout details
	workoutDetails, err := db.GetWorkoutDetails(workoutID)
	if err != nil {
		t.Fatalf("Failed to get workout details: %v", err)
	}
	if len(workoutDetails.Exercises) != 1 {
		t.Errorf("Expected 1 exercise in workout, got %d", len(workoutDetails.Exercises))
	}
	if len(workoutDetails.Exercises[0].Sets) != 2 {
		t.Errorf("Expected 2 sets in exercise, got %d", len(workoutDetails.Exercises[0].Sets))
	}

	// 8. Get exercise history
	history, err := db.GetExerciseHistory(exerciseID)
	if err != nil {
		t.Fatalf("Failed to get exercise history: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}

	// 9. Get exercise sets for workout
	sets, err := db.GetExerciseSetsForWorkout(workoutExerciseID)
	if err != nil {
		t.Fatalf("Failed to get exercise sets: %v", err)
	}
	if len(sets) != 2 {
		t.Errorf("Expected 2 sets, got %d", len(sets))
	}

	// 10. Test deleting workout exercises and sets
	err = db.DeleteWorkoutExercisesAndSets(workoutID)
	if err != nil {
		t.Fatalf("Failed to delete workout exercises and sets: %v", err)
	}

	// Verify deletion
	workoutDetails, err = db.GetWorkoutDetails(workoutID)
	if err != nil {
		t.Fatalf("Failed to get workout details after deletion: %v", err)
	}
	if len(workoutDetails.Exercises) != 0 {
		t.Errorf("Expected 0 exercises after deletion, got %d", len(workoutDetails.Exercises))
	}

	// 11. Test deleting workout (should still exist)
	workouts, err := db.GetWorkouts()
	if err != nil {
		t.Fatalf("Failed to get workouts: %v", err)
	}
	if len(workouts) != 1 {
		t.Errorf("Expected 1 workout to still exist, got %d", len(workouts))
	}

	// 12. Delete workout
	err = db.DeleteWorkout(workoutID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	// Verify deletion
	workouts, err = db.GetWorkouts()
	if err != nil {
		t.Fatalf("Failed to get workouts after deletion: %v", err)
	}
	if len(workouts) != 0 {
		t.Errorf("Expected 0 workouts after deletion, got %d", len(workouts))
	}
}
