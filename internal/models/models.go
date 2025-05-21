package models

import "time"

// Exercise represents a type of weightlifting exercise
type Exercise struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	OneRepMax   float64   `json:"one_rep_max,omitempty"` // This field is not stored in DB, used for form display only
}

// OneRepMax represents a user's one rep maximum for a specific exercise
type OneRepMax struct {
	ID         int64     `json:"id"`
	ExerciseID int64     `json:"exercise_id"`
	Weight     float64   `json:"weight"` // in kg
	Date       time.Time `json:"date"`
}

// Workout represents a single workout session
type Workout struct {
	ID        int64     `json:"id"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// WorkoutExercise represents an exercise performed during a workout
type WorkoutExercise struct {
	ID         int64     `json:"id"`
	WorkoutID  int64     `json:"workout_id"`
	ExerciseID int64     `json:"exercise_id"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Set represents a single set of an exercise
type Set struct {
	ID                int64   `json:"id"`
	WorkoutExerciseID int64   `json:"workout_exercise_id"`
	Reps              int     `json:"reps"`
	Weight            float64 `json:"weight"` // in kg
	PercentageOfMax   float64 `json:"percentage_of_max,omitempty"`
	SetOrder          int     `json:"set_order"` // To track the order of sets
	RangeID           string  `json:"range_id,omitempty"` // To group sets into ranges
}

// ExerciseWithSets is a helper struct for returning an exercise with all its sets
type ExerciseWithSets struct {
	Exercise        Exercise        `json:"exercise"`
	WorkoutExercise WorkoutExercise `json:"workout_exercise"`
	Sets            []Set           `json:"sets"`
}

// WorkoutWithExercises is a helper struct for returning a workout with all its exercises and sets
type WorkoutWithExercises struct {
	Workout   Workout            `json:"workout"`
	Exercises []ExerciseWithSets `json:"exercises"`
}
