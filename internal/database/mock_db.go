package database

import (
	"database/sql"
	"time"

	"github.com/mahalel/wltrack/internal/models"
)

// MockDB is a mock implementation of the database functionality for testing
type MockDB struct {
	exercises        map[int64]models.Exercise
	workouts         map[int64]models.Workout
	workoutExercises map[int64]models.WorkoutExercise
	sets             map[int64]models.Set
	oneRepMaxes      map[int64]models.OneRepMax

	// Auto-increment counters
	exerciseID        int64
	workoutID         int64
	workoutExerciseID int64
	setID             int64
	oneRepMaxID       int64
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{
		exercises:        make(map[int64]models.Exercise),
		workouts:         make(map[int64]models.Workout),
		workoutExercises: make(map[int64]models.WorkoutExercise),
		sets:             make(map[int64]models.Set),
		oneRepMaxes:      make(map[int64]models.OneRepMax),

		exerciseID:        1,
		workoutID:         1,
		workoutExerciseID: 1,
		setID:             1,
		oneRepMaxID:       1,
	}
}

// SetupDB initializes the mock database
func (db *MockDB) SetupDB() error {
	return nil
}

// CloseDB closes the database connection
func (db *MockDB) CloseDB() error {
	return nil
}

// GetAllExercises retrieves all exercises
func (db *MockDB) GetAllExercises() ([]models.Exercise, error) {
	exercises := make([]models.Exercise, 0, len(db.exercises))
	for _, ex := range db.exercises {
		exercises = append(exercises, ex)
	}
	return exercises, nil
}

// GetExercise retrieves a single exercise by ID
func (db *MockDB) GetExercise(id int64) (models.Exercise, error) {
	if ex, ok := db.exercises[id]; ok {
		return ex, nil
	}
	return models.Exercise{}, sql.ErrNoRows
}

// AddExercise adds a new exercise
func (db *MockDB) AddExercise(name, description string) (int64, error) {
	id := db.exerciseID
	db.exercises[id] = models.Exercise{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}
	db.exerciseID++
	return id, nil
}

// UpdateExercise updates an existing exercise
func (db *MockDB) UpdateExercise(id int64, name, description string) error {
	if _, ok := db.exercises[id]; !ok {
		return sql.ErrNoRows
	}
	ex := db.exercises[id]
	ex.Name = name
	ex.Description = description
	db.exercises[id] = ex
	return nil
}

// DeleteExercise deletes an exercise
func (db *MockDB) DeleteExercise(id int64) error {
	if _, ok := db.exercises[id]; !ok {
		return sql.ErrNoRows
	}
	delete(db.exercises, id)
	return nil
}

// SaveOneRepMax saves a one rep max for an exercise
func (db *MockDB) SaveOneRepMax(exerciseID int64, weight float64) (int64, error) {
	if _, ok := db.exercises[exerciseID]; !ok {
		return 0, sql.ErrNoRows
	}

	id := db.oneRepMaxID
	db.oneRepMaxes[id] = models.OneRepMax{
		ID:         id,
		ExerciseID: exerciseID,
		Weight:     weight,
		Date:       time.Now(),
	}
	db.oneRepMaxID++
	return id, nil
}

// GetLatestOneRepMax gets the most recent 1RM for an exercise
func (db *MockDB) GetLatestOneRepMax(exerciseID int64) (models.OneRepMax, error) {
	var latest models.OneRepMax
	var latestTime time.Time

	for _, orm := range db.oneRepMaxes {
		if orm.ExerciseID == exerciseID {
			ormTime := orm.Date
			if latest.ID == 0 || ormTime.After(latestTime) {
				latest = orm
				latestTime = ormTime
			}
		}
	}

	if latest.ID == 0 {
		return models.OneRepMax{}, sql.ErrNoRows
	}

	return latest, nil
}

// AddWorkout adds a new workout
func (db *MockDB) AddWorkout(date string, notes string) (int64, error) {
	id := db.workoutID
	// Convert date string to time.Time
	dateTime, _ := time.Parse("2006-01-02", date)
	db.workouts[id] = models.Workout{
		ID:        id,
		Date:      dateTime,
		Notes:     notes,
		CreatedAt: time.Now(),
	}
	db.workoutID++
	return id, nil
}

// UpdateWorkout updates an existing workout
func (db *MockDB) UpdateWorkout(id int64, date string, notes string) error {
	if _, ok := db.workouts[id]; !ok {
		return sql.ErrNoRows
	}
	dateTime, _ := time.Parse("2006-01-02", date)
	w := db.workouts[id]
	w.Date = dateTime
	w.Notes = notes
	db.workouts[id] = w
	return nil
}

// AddWorkoutExercise adds an exercise to a workout
func (db *MockDB) AddWorkoutExercise(workoutID, exerciseID int64, notes string) (int64, error) {
	if _, ok := db.workouts[workoutID]; !ok {
		return 0, sql.ErrNoRows
	}
	if _, ok := db.exercises[exerciseID]; !ok {
		return 0, sql.ErrNoRows
	}

	id := db.workoutExerciseID
	db.workoutExercises[id] = models.WorkoutExercise{
		ID:         id,
		WorkoutID:  workoutID,
		ExerciseID: exerciseID,
		Notes:      notes,
		CreatedAt:  time.Now(),
	}
	db.workoutExerciseID++
	return id, nil
}

// AddSet adds a set to a workout exercise
func (db *MockDB) AddSet(workoutExerciseID int64, reps int, weight float64, percentageOfMax float64, setOrder int, rangeID string) (int64, error) {
	if _, ok := db.workoutExercises[workoutExerciseID]; !ok {
		return 0, sql.ErrNoRows
	}

	id := db.setID
	db.sets[id] = models.Set{
		ID:                id,
		WorkoutExerciseID: workoutExerciseID,
		Reps:              reps,
		Weight:            weight,
		PercentageOfMax:   percentageOfMax,
		SetOrder:          setOrder,
		RangeID:           rangeID,
	}
	db.setID++
	return id, nil
}

// GetWorkouts gets all workouts
func (db *MockDB) GetWorkouts() ([]models.Workout, error) {
	workouts := make([]models.Workout, 0, len(db.workouts))
	for _, w := range db.workouts {
		workouts = append(workouts, w)
	}
	return workouts, nil
}

// GetWorkoutDetails gets a workout with all exercises and sets
func (db *MockDB) GetWorkoutDetails(workoutID int64) (models.WorkoutWithExercises, error) {
	result := models.WorkoutWithExercises{}

	workout, ok := db.workouts[workoutID]
	if !ok {
		return result, sql.ErrNoRows
	}

	result.Workout = workout

	// Find all workout exercises for this workout
	for weID, we := range db.workoutExercises {
		if we.WorkoutID == workoutID {
			exercise, ok := db.exercises[we.ExerciseID]
			if !ok {
				continue // Skip if exercise not found
			}

			exWithSets := models.ExerciseWithSets{
				Exercise:        exercise,
				WorkoutExercise: we,
			}

			// Find all sets for this workout exercise
			for _, set := range db.sets {
				if set.WorkoutExerciseID == weID {
					exWithSets.Sets = append(exWithSets.Sets, set)
				}
			}

			result.Exercises = append(result.Exercises, exWithSets)
		}
	}

	return result, nil
}

// GetExerciseHistory gets the history of an exercise
func (db *MockDB) GetExerciseHistory(exerciseID int64) ([]models.WorkoutExercise, error) {
	if _, ok := db.exercises[exerciseID]; !ok {
		return nil, sql.ErrNoRows
	}

	var history []models.WorkoutExercise
	for _, we := range db.workoutExercises {
		if we.ExerciseID == exerciseID {
			// Make sure CreatedAt is initialized
			if we.CreatedAt.IsZero() {
				we.CreatedAt = time.Now()
			}
			history = append(history, we)
		}
	}

	return history, nil
}

// GetExerciseSetsForWorkout gets all sets for a specific exercise in a workout
func (db *MockDB) GetExerciseSetsForWorkout(workoutExerciseID int64) ([]models.Set, error) {
	if _, ok := db.workoutExercises[workoutExerciseID]; !ok {
		return nil, sql.ErrNoRows
	}

	var sets []models.Set
	for _, set := range db.sets {
		if set.WorkoutExerciseID == workoutExerciseID {
			sets = append(sets, set)
		}
	}

	return sets, nil
}

// DeleteWorkout deletes a workout and all related data
func (db *MockDB) DeleteWorkout(workoutID int64) error {
	if _, ok := db.workouts[workoutID]; !ok {
		return sql.ErrNoRows
	}

	// Delete all workout exercises and their sets
	for weID, we := range db.workoutExercises {
		if we.WorkoutID == workoutID {
			for setID, set := range db.sets {
				if set.WorkoutExerciseID == weID {
					delete(db.sets, setID)
				}
			}
			delete(db.workoutExercises, weID)
		}
	}

	delete(db.workouts, workoutID)
	return nil
}

// DeleteWorkoutExercisesAndSets deletes all exercises and sets for a workout
func (db *MockDB) DeleteWorkoutExercisesAndSets(workoutID int64) error {
	if _, ok := db.workouts[workoutID]; !ok {
		return sql.ErrNoRows
	}

	// Delete all workout exercises and their sets
	for weID, we := range db.workoutExercises {
		if we.WorkoutID == workoutID {
			for setID, set := range db.sets {
				if set.WorkoutExerciseID == weID {
					delete(db.sets, setID)
				}
			}
			delete(db.workoutExercises, weID)
		}
	}

	return nil
}
