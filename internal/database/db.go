package database

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/mahalel/wltrack/internal/models"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// DB represents the database connection
type DB struct {
	db *sql.DB
}

// New creates a new database connection
func New(url, authToken string) (*DB, error) {
	var driver, connStr string

	if strings.HasPrefix(url, "file:") {
		// Local SQLite database
		driver = "sqlite3"
		connStr = strings.TrimPrefix(url, "file:")
		log.Println("Using local SQLite database")
	} else {
		// Turso database
		driver = "libsql"
		connStr = url
		if authToken != "" {
			connStr += "?authToken=" + authToken
		}
		log.Println("Using Turso database")
	}

	conn, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}

	return &DB{db: conn}, nil
}

// SetupDB initializes the database with tables if they don't exist
func (db *DB) SetupDB() error {
	ctx := context.Background()

	// Create exercises table
	_, err := db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS exercises (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create one_rep_max table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS one_rep_max (
			id INTEGER PRIMARY KEY,
			exercise_id INTEGER NOT NULL,
			weight REAL NOT NULL,
			date TIMESTAMP NOT NULL,
			FOREIGN KEY (exercise_id) REFERENCES exercises(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create workouts table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY,
			date TIMESTAMP NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create workout_exercises table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workout_exercises (
			id INTEGER PRIMARY KEY,
			workout_id INTEGER NOT NULL,
			exercise_id INTEGER NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workout_id) REFERENCES workouts(id),
			FOREIGN KEY (exercise_id) REFERENCES exercises(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create sets table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sets (
			id INTEGER PRIMARY KEY,
			workout_exercise_id INTEGER NOT NULL,
			reps INTEGER NOT NULL,
			weight REAL NOT NULL,
			percentage_of_max REAL,
			set_order INTEGER NOT NULL,
			FOREIGN KEY (workout_exercise_id) REFERENCES workout_exercises(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// CloseDB closes the database connection
func (db *DB) CloseDB() error {
	return db.db.Close()
}

// GetAllExercises retrieves all exercises from the database
func (db *DB) GetAllExercises() ([]models.Exercise, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `SELECT id, name, description, created_at FROM exercises ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		if err := rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.CreatedAt); err != nil {
			return nil, err
		}
		exercises = append(exercises, ex)
	}

	return exercises, nil
}

// GetExercise retrieves a single exercise by ID
func (db *DB) GetExercise(id int64) (models.Exercise, error) {
	ctx := context.Background()
	var ex models.Exercise
	err := db.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at FROM exercises WHERE id = ?`,
		id,
	).Scan(&ex.ID, &ex.Name, &ex.Description, &ex.CreatedAt)
	return ex, err
}

// AddExercise adds a new exercise
func (db *DB) AddExercise(name, description string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO exercises (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// SaveOneRepMax saves a one rep max for an exercise
func (db *DB) SaveOneRepMax(exerciseID int64, weight float64) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO one_rep_max (exercise_id, weight, date) VALUES (?, ?, CURRENT_TIMESTAMP)`,
		exerciseID, weight,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetLatestOneRepMax gets the most recent 1RM for an exercise
func (db *DB) GetLatestOneRepMax(exerciseID int64) (models.OneRepMax, error) {
	ctx := context.Background()
	var orm models.OneRepMax
	err := db.db.QueryRowContext(ctx,
		`SELECT id, exercise_id, weight, date FROM one_rep_max 
		 WHERE exercise_id = ? ORDER BY date DESC LIMIT 1`,
		exerciseID,
	).Scan(&orm.ID, &orm.ExerciseID, &orm.Weight, &orm.Date)
	if err != nil {
		return models.OneRepMax{}, err
	}
	return orm, nil
}

// UpdateExercise updates an existing exercise
func (db *DB) UpdateExercise(id int64, name, description string) error {
	ctx := context.Background()
	_, err := db.db.ExecContext(ctx,
		`UPDATE exercises SET name = ?, description = ? WHERE id = ?`,
		name, description, id,
	)
	return err
}

// AddWorkout adds a new workout
func (db *DB) AddWorkout(date string, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO workouts (date, notes) VALUES (?, ?)`,
		date, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateWorkout updates an existing workout
func (db *DB) UpdateWorkout(id int64, date string, notes string) error {
	ctx := context.Background()
	_, err := db.db.ExecContext(ctx,
		`UPDATE workouts SET date = ?, notes = ? WHERE id = ?`,
		date, notes, id,
	)
	return err
}

// AddWorkoutExercise adds an exercise to a workout
func (db *DB) AddWorkoutExercise(workoutID, exerciseID int64, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO workout_exercises (workout_id, exercise_id, notes) VALUES (?, ?, ?)`,
		workoutID, exerciseID, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// AddSet adds a set to a workout exercise
func (db *DB) AddSet(workoutExerciseID int64, reps int, weight float64, percentageOfMax float64, setOrder int) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO sets (workout_exercise_id, reps, weight, percentage_of_max, set_order) 
		 VALUES (?, ?, ?, ?, ?)`,
		workoutExerciseID, reps, weight, percentageOfMax, setOrder,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetWorkouts gets all workouts, ordered by date
func (db *DB) GetWorkouts() ([]models.Workout, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `SELECT id, date, notes, created_at FROM workouts ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []models.Workout
	for rows.Next() {
		var w models.Workout
		if err := rows.Scan(&w.ID, &w.Date, &w.Notes, &w.CreatedAt); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}

	return workouts, nil
}

// GetWorkoutDetails gets a workout with all exercises and sets
func (db *DB) GetWorkoutDetails(workoutID int64) (models.WorkoutWithExercises, error) {
	ctx := context.Background()
	var workout models.Workout
	var result models.WorkoutWithExercises

	// Get workout
	err := db.db.QueryRowContext(ctx,
		`SELECT id, date, notes, created_at FROM workouts WHERE id = ?`,
		workoutID,
	).Scan(&workout.ID, &workout.Date, &workout.Notes, &workout.CreatedAt)
	if err != nil {
		return result, err
	}

	result.Workout = workout

	// Get workout exercises
	exercisesRows, err := db.db.QueryContext(ctx, `
		SELECT we.id, we.exercise_id, we.notes, e.name, e.description, e.created_at
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.id
		WHERE we.workout_id = ?
		ORDER BY we.id
	`, workoutID)
	if err != nil {
		return result, err
	}
	defer exercisesRows.Close()

	for exercisesRows.Next() {
		var ex models.ExerciseWithSets
		var we models.WorkoutExercise
		var e models.Exercise

		if err := exercisesRows.Scan(
			&we.ID, &we.ExerciseID, &we.Notes,
			&e.Name, &e.Description, &e.CreatedAt,
		); err != nil {
			return result, err
		}

		we.WorkoutID = workoutID
		e.ID = we.ExerciseID
		ex.Exercise = e
		ex.WorkoutExercise = we

		// Get sets for this exercise
		setsRows, err := db.db.QueryContext(ctx, `
			SELECT id, reps, weight, percentage_of_max, set_order
			FROM sets
			WHERE workout_exercise_id = ?
			ORDER BY set_order
		`, we.ID)
		if err != nil {
			return result, err
		}

		var sets []models.Set
		for setsRows.Next() {
			var s models.Set
			if err := setsRows.Scan(&s.ID, &s.Reps, &s.Weight, &s.PercentageOfMax, &s.SetOrder); err != nil {
				setsRows.Close()
				return result, err
			}
			s.WorkoutExerciseID = we.ID
			sets = append(sets, s)
		}
		setsRows.Close()

		ex.Sets = sets
		result.Exercises = append(result.Exercises, ex)
	}

	return result, nil
}

// GetExerciseHistory gets the history of an exercise
func (db *DB) GetExerciseHistory(exerciseID int64) ([]models.WorkoutExercise, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `
		SELECT we.id, we.workout_id, w.date, we.notes, we.created_at
		FROM workout_exercises we
		JOIN workouts w ON we.workout_id = w.id
		WHERE we.exercise_id = ?
		ORDER BY w.date DESC
	`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.WorkoutExercise
	for rows.Next() {
		var we models.WorkoutExercise
		var workoutDate time.Time
		if err := rows.Scan(&we.ID, &we.WorkoutID, &workoutDate, &we.Notes, &we.CreatedAt); err != nil {
			return nil, err
		}
		we.ExerciseID = exerciseID
		history = append(history, we)
	}

	return history, nil
}

// GetExerciseSetsForWorkout gets all sets for a specific exercise in a workout
func (db *DB) GetExerciseSetsForWorkout(workoutExerciseID int64) ([]models.Set, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `
		SELECT id, reps, weight, percentage_of_max, set_order
		FROM sets
		WHERE workout_exercise_id = ?
		ORDER BY set_order
	`, workoutExerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []models.Set
	for rows.Next() {
		var s models.Set
		if err := rows.Scan(&s.ID, &s.Reps, &s.Weight, &s.PercentageOfMax, &s.SetOrder); err != nil {
			return nil, err
		}
		s.WorkoutExerciseID = workoutExerciseID
		sets = append(sets, s)
	}

	return sets, nil
}

// DeleteExercise deletes an exercise and all related data
func (db *DB) DeleteExercise(exerciseID int64) error {
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Find all workout_exercises for this exercise
	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM workout_exercises WHERE exercise_id = ?
	`, exerciseID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Delete all sets for each workout_exercise
	for rows.Next() {
		var workoutExerciseID int64
		if err := rows.Scan(&workoutExerciseID); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM sets WHERE workout_exercise_id = ?`, workoutExerciseID)
		if err != nil {
			return err
		}
	}
	rows.Close()

	// Delete all workout_exercises for this exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_exercises WHERE exercise_id = ?`, exerciseID)
	if err != nil {
		return err
	}

	// Delete all one_rep_max entries for this exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM one_rep_max WHERE exercise_id = ?`, exerciseID)
	if err != nil {
		return err
	}

	// Delete the exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM exercises WHERE id = ?`, exerciseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteWorkout deletes a workout and all related data
func (db *DB) DeleteWorkout(workoutID int64) error {
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Find all workout_exercises for this workout
	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM workout_exercises WHERE workout_id = ?
	`, workoutID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Delete all sets for each workout_exercise
	for rows.Next() {
		var workoutExerciseID int64
		if err := rows.Scan(&workoutExerciseID); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM sets WHERE workout_exercise_id = ?`, workoutExerciseID)
		if err != nil {
			return err
		}
	}
	rows.Close()

	// Delete all workout_exercises for this workout
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_exercises WHERE workout_id = ?`, workoutID)
	if err != nil {
		return err
	}

	// Delete the workout
	_, err = tx.ExecContext(ctx, `DELETE FROM workouts WHERE id = ?`, workoutID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteWorkoutExercisesAndSets deletes all exercises and sets for a workout without deleting the workout itself
func (db *DB) DeleteWorkoutExercisesAndSets(workoutID int64) error {
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Find all workout_exercises for this workout
	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM workout_exercises WHERE workout_id = ?
	`, workoutID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Delete all sets for each workout_exercise
	for rows.Next() {
		var workoutExerciseID int64
		if err := rows.Scan(&workoutExerciseID); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM sets WHERE workout_exercise_id = ?`, workoutExerciseID)
		if err != nil {
			return err
		}
	}
	rows.Close()

	// Delete all workout_exercises for this workout
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_exercises WHERE workout_id = ?`, workoutID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
