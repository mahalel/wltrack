package models

import "time"

// Exercise represents a type of weightlifting exercise
type Exercise struct {
	ID         int64     `json:"exercise_id"`
	Name       string    `json:"name"`
	Current1RM float64   `json:"current_1rm,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// OneRepMax represents a user's one rep maximum history for a specific exercise
type OneRepMax struct {
	ID         int64     `json:"history_id"`
	ExerciseID int64     `json:"exercise_id"`
	Weight     float64   `json:"onerm_value"` // in kg
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// PersonalRecord represents a personal record for an exercise
type PersonalRecord struct {
	ID         int64     `json:"pr_id"`
	ExerciseID int64     `json:"exercise_id"`
	WorkoutID  int64     `json:"workout_id,omitempty"`
	Weight     float64   `json:"weight,omitempty"`
	Date       time.Time `json:"date"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Workout represents a single workout session
type Workout struct {
	ID         int64     `json:"workout_id"`
	Date       time.Time `json:"date"`
	TemplateID int64     `json:"template_id,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// WorkoutTemplate represents a reusable workout plan
type WorkoutTemplate struct {
	ID          int64     `json:"template_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// TemplateExercise represents an exercise within a workout template
type TemplateExercise struct {
	ID         int64  `json:"template_exercise_id"`
	TemplateID int64  `json:"template_id"`
	ExerciseID int64  `json:"exercise_id"`
	OrderIndex int    `json:"order_index"`
	Notes      string `json:"notes,omitempty"`
}

// WorkoutExercise represents an exercise performed during a workout
type WorkoutExercise struct {
	ID         int64  `json:"workout_exercise_id"`
	WorkoutID  int64  `json:"workout_id"`
	ExerciseID int64  `json:"exercise_id"`
	OrderIndex int    `json:"order_index"`
	Notes      string `json:"notes,omitempty"`
}

// Set represents a single set of an exercise
type Set struct {
	ID                int64   `json:"set_id"`
	WorkoutExerciseID int64   `json:"workout_exercise_id"`
	SetNumber         int     `json:"set_number"`
	SetRange          string  `json:"set_range,omitempty"`
	Reps              int     `json:"reps,omitempty"`
	Percentage1RM     int     `json:"percentage_1rm,omitempty"`
	Weight            float64 `json:"weight"` // in kg
	Notes             string  `json:"notes,omitempty"`
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

// TemplateWithExercises is a helper struct for returning a template with all its exercises
type TemplateWithExercises struct {
	Template  WorkoutTemplate    `json:"template"`
	Exercises []TemplateExercise `json:"exercises"`
}
