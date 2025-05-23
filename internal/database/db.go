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
			exercise_id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			current_1rm REAL,
			notes TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create workouts table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workouts (
			workout_id INTEGER PRIMARY KEY,
			date TEXT NOT NULL,
			template_id INTEGER,
			notes TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (template_id) REFERENCES workout_templates (template_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create workout_exercises table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workout_exercises (
			workout_exercise_id INTEGER PRIMARY KEY,
			workout_id INTEGER NOT NULL,
			exercise_id INTEGER NOT NULL,
			order_index INTEGER DEFAULT 0,
			notes TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workout_id) REFERENCES workouts (workout_id) ON DELETE CASCADE,
			FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create sets table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sets (
			set_id INTEGER PRIMARY KEY,
			workout_exercise_id INTEGER NOT NULL,
			set_number INTEGER NOT NULL,
			set_range TEXT,
			reps INTEGER,
			percentage_1rm INTEGER,
			weight REAL NOT NULL,
			notes TEXT,
			FOREIGN KEY (workout_exercise_id) REFERENCES workout_exercises (workout_exercise_id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create workout_templates table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workout_templates (
			template_id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create template_exercises table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS template_exercises (
			template_exercise_id INTEGER PRIMARY KEY,
			template_id INTEGER NOT NULL,
			exercise_id INTEGER NOT NULL,
			order_index INTEGER DEFAULT 0,
			notes TEXT,
			FOREIGN KEY (template_id) REFERENCES workout_templates (template_id) ON DELETE CASCADE,
			FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create onerm_history table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS onerm_history (
			history_id INTEGER PRIMARY KEY,
			exercise_id INTEGER NOT NULL,
			onerm_value REAL NOT NULL,
			notes TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create personal_records table
	_, err = db.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS personal_records (
			pr_id INTEGER PRIMARY KEY,
			exercise_id INTEGER NOT NULL,
			workout_id INTEGER,
			weight REAL,
			date TEXT NOT NULL,
			notes TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id),
			FOREIGN KEY (workout_id) REFERENCES workouts (workout_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create performance indices
	// Workout indices
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_workout_date ON workouts(date)`)
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_workout_template ON workouts(template_id)`)
	if err != nil {
		return err
	}

	// Workout exercises indices
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_workout_exercises_workout ON workout_exercises(workout_id)`)
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_workout_exercises_exercise ON workout_exercises(exercise_id)`)
	if err != nil {
		return err
	}

	// Sets indices
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_sets_workout_exercise ON sets(workout_exercise_id)`)
	if err != nil {
		return err
	}

	// Template indices
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_template_exercises_template ON template_exercises(template_id)`)
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_template_exercises_exercise ON template_exercises(exercise_id)`)
	if err != nil {
		return err
	}

	// History and records indices
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_onerm_exercise ON onerm_history(exercise_id)`)
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_pr_exercise ON personal_records(exercise_id)`)
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_pr_date ON personal_records(date)`)
	if err != nil {
		return err
	}

	// Ensure exercise names are unique
	_, err = db.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_exercise_name ON exercises(name)`)
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
	rows, err := db.db.QueryContext(ctx, `SELECT exercise_id, name, current_1rm, notes, created_at FROM exercises ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var exercises []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		if err := rows.Scan(&ex.ID, &ex.Name, &ex.Current1RM, &ex.Notes, &ex.CreatedAt); err != nil {
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
		`SELECT exercise_id, name, current_1rm, notes, created_at FROM exercises WHERE exercise_id = ?`,
		id,
	).Scan(&ex.ID, &ex.Name, &ex.Current1RM, &ex.Notes, &ex.CreatedAt)
	return ex, err
}

// AddExercise adds a new exercise
func (db *DB) AddExercise(name string, current1RM float64, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO exercises (name, current_1rm, notes) VALUES (?, ?, ?)`,
		name, current1RM, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// SaveOneRepMax saves a one rep max for an exercise
func (db *DB) SaveOneRepMax(exerciseID int64, weight float64) (int64, error) {
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Update the current_1rm in the exercises table
	_, err = tx.ExecContext(ctx,
		`UPDATE exercises SET current_1rm = ? WHERE exercise_id = ?`,
		weight, exerciseID,
	)
	if err != nil {
		return 0, err
	}

	// Insert into onerm_history
	result, err := tx.ExecContext(ctx,
		`INSERT INTO onerm_history (exercise_id, created_at, onerm_value)
		 VALUES (?, CURRENT_TIMESTAMP, ?')`,
		exerciseID, weight,
	)
	if err != nil {
		return 0, err
	}

	historyID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return historyID, nil
}

// GetLatestOneRepMax gets the most recent 1RM for an exercise
func (db *DB) GetLatestOneRepMax(exerciseID int64) (models.OneRepMax, error) {
	ctx := context.Background()
	var orm models.OneRepMax
	err := db.db.QueryRowContext(ctx,
		`SELECT history_id, exercise_id, onerm_value, created_at FROM onerm_history
		 WHERE exercise_id = ? ORDER BY date DESC LIMIT 1`,
		exerciseID,
	).Scan(&orm.ID, &orm.ExerciseID, &orm.Weight, &orm.CreatedAt)
	if err != nil {
		return models.OneRepMax{}, err
	}
	return orm, nil
}

// UpdateExercise updates an existing exercise
func (db *DB) UpdateExercise(id int64, name string, current1RM float64, notes string) error {
	ctx := context.Background()
	_, err := db.db.ExecContext(ctx,
		`UPDATE exercises SET name = ?, current_1rm = ?, notes = ? WHERE exercise_id = ?`,
		name, current1RM, notes, id,
	)
	return err
}

// AddWorkout adds a new workout
func (db *DB) AddWorkout(date string, templateID int64, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO workouts (date, template_id, notes, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		date, templateID, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateWorkout updates an existing workout
func (db *DB) UpdateWorkout(id int64, date string, templateID int64, notes string) error {
	ctx := context.Background()
	_, err := db.db.ExecContext(ctx,
		`UPDATE workouts SET date = ?, template_id = ?, notes = ? WHERE workout_id = ?`,
		date, templateID, notes, id,
	)
	return err
}

// AddWorkoutExercise adds an exercise to a workout
func (db *DB) AddWorkoutExercise(workoutID, exerciseID int64, orderIndex int, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO workout_exercises (workout_id, exercise_id, order_index, notes) VALUES (?, ?, ?, ?)`,
		workoutID, exerciseID, orderIndex, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// AddSet adds a set to a workout exercise
func (db *DB) AddSet(workoutExerciseID int64, setNumber int, setRange string, reps int,
	percentage1RM int, weight float64, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO sets (workout_exercise_id, set_number, set_range, reps,
		                   percentage_1rm, weight, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		workoutExerciseID, setNumber, setRange, reps,
		percentage1RM, weight, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetWorkouts gets all workouts, ordered by date
func (db *DB) GetWorkouts() ([]models.Workout, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `SELECT workout_id, date, template_id, notes, created_at FROM workouts ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var workouts []models.Workout
	for rows.Next() {
		var w models.Workout
		if err := rows.Scan(&w.ID, &w.Date, &w.TemplateID, &w.Notes, &w.CreatedAt); err != nil {
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
		`SELECT workout_id, date, template_id, notes, created_at FROM workouts WHERE workout_id = ?`,
		workoutID,
	).Scan(&workout.ID, &workout.Date, &workout.TemplateID, &workout.Notes, &workout.CreatedAt)
	if err != nil {
		return result, err
	}

	result.Workout = workout

	// Get workout exercises
	exercisesRows, err := db.db.QueryContext(ctx, `
		SELECT we.workout_exercise_id, we.exercise_id, we.order_index, we.notes,
		       e.name, e.current_1rm, e.notes, e.created_at
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.exercise_id
		WHERE we.workout_id = ?
		ORDER BY we.order_index
	`, workoutID)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := exercisesRows.Close(); err != nil {
			log.Printf("Error closing exercise rows: %v", err)
		}
	}()

	for exercisesRows.Next() {
		var ex models.ExerciseWithSets
		var we models.WorkoutExercise
		var e models.Exercise

		if err := exercisesRows.Scan(
			&we.ID, &we.ExerciseID, &we.OrderIndex, &we.Notes,
			&e.Name, &e.Current1RM, &e.Notes, &e.CreatedAt,
		); err != nil {
			return result, err
		}

		we.WorkoutID = workoutID
		e.ID = we.ExerciseID
		ex.Exercise = e
		ex.WorkoutExercise = we

		// Get sets for this exercise
		setsRows, err := db.db.QueryContext(ctx, `
			SELECT set_id, set_number, set_range, reps,
			       percentage_1rm, weight, notes
			FROM sets
			WHERE workout_exercise_id = ?
			ORDER BY set_number
		`, we.ID)
		if err != nil {
			return result, err
		}

		var sets []models.Set
		for setsRows.Next() {
			var s models.Set
			if err := setsRows.Scan(&s.ID, &s.SetNumber, &s.SetRange, &s.Reps,
				&s.Percentage1RM, &s.Weight, &s.Notes); err != nil {
				if err := setsRows.Close(); err != nil {
					log.Printf("Error closing sets rows: %v", err)
				}
				return result, err
			}
			s.WorkoutExerciseID = we.ID
			sets = append(sets, s)
		}
		if err := setsRows.Close(); err != nil {
			log.Printf("Error closing sets rows: %v", err)
		}

		ex.Sets = sets
		result.Exercises = append(result.Exercises, ex)
	}

	return result, nil
}

// GetExerciseHistory gets the history of an exercise
func (db *DB) GetExerciseHistory(exerciseID int64) ([]models.WorkoutExercise, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `
		SELECT we.workout_exercise_id, we.workout_id, w.date, we.order_index, we.notes, we.created_at
		FROM workout_exercises we
		JOIN workouts w ON we.workout_id = w.workout_id
		WHERE we.exercise_id = ?
		ORDER BY w.date DESC
	`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var history []models.WorkoutExercise
	for rows.Next() {
		var we models.WorkoutExercise
		var workoutDate time.Time
		if err := rows.Scan(&we.ID, &we.WorkoutID, &workoutDate, &we.OrderIndex, &we.Notes); err != nil {
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
		SELECT set_id, set_number, set_range, reps,
		       percentage_1rm, weight, notes
		FROM sets
		WHERE workout_exercise_id = ?
		ORDER BY set_number
	`, workoutExerciseID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var sets []models.Set
	for rows.Next() {
		var s models.Set
		if err := rows.Scan(&s.ID, &s.SetNumber, &s.SetRange, &s.Reps,
			&s.Percentage1RM, &s.Weight, &s.Notes); err != nil {
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
		SELECT workout_exercise_id FROM workout_exercises WHERE exercise_id = ?
	`, exerciseID)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

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
	if err := rows.Close(); err != nil {
		log.Printf("Error closing rows: %v", err)
	}

	// Delete all workout_exercises for this exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_exercises WHERE exercise_id = ?`, exerciseID)
	if err != nil {
		return err
	}

	// Delete all onerm_history entries for this exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM onerm_history WHERE exercise_id = ?`, exerciseID)
	if err != nil {
		return err
	}

	// Delete all personal_records entries for this exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM personal_records WHERE exercise_id = ?`, exerciseID)
	if err != nil {
		return err
	}

	// Delete all template_exercises entries for this exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM template_exercises WHERE exercise_id = ?`, exerciseID)
	if err != nil {
		return err
	}

	// Delete the exercise
	_, err = tx.ExecContext(ctx, `DELETE FROM exercises WHERE exercise_id = ?`, exerciseID)
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
		SELECT workout_exercise_id FROM workout_exercises WHERE workout_id = ?
	`, workoutID)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

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
	if err := rows.Close(); err != nil {
		log.Printf("Error closing rows: %v", err)
	}

	// Delete all workout_exercises for this workout
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_exercises WHERE workout_id = ?`, workoutID)
	if err != nil {
		return err
	}

	// Delete any personal records associated with this workout
	_, err = tx.ExecContext(ctx, `DELETE FROM personal_records WHERE workout_id = ?`, workoutID)
	if err != nil {
		return err
	}

	// Delete the workout
	_, err = tx.ExecContext(ctx, `DELETE FROM workouts WHERE workout_id = ?`, workoutID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AddWorkoutTemplate adds a new workout template
func (db *DB) AddWorkoutTemplate(name, description, category string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO workout_templates (name, description, category) VALUES (?, ?, ?)`,
		name, description, category,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateWorkoutTemplate updates an existing workout template
func (db *DB) UpdateWorkoutTemplate(templateID int64, name, description, category string) error {
	ctx := context.Background()
	_, err := db.db.ExecContext(ctx,
		`UPDATE workout_templates SET name = ?, description = ?, category = ? WHERE template_id = ?`,
		name, description, category, templateID,
	)
	return err
}

// GetWorkoutTemplates gets all workout templates
func (db *DB) GetWorkoutTemplates() ([]models.WorkoutTemplate, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `SELECT template_id, name, description, created_at FROM workout_templates ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var templates []models.WorkoutTemplate
	for rows.Next() {
		var t models.WorkoutTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	return templates, nil
}

// GetWorkoutTemplate gets a single workout template by ID
func (db *DB) GetWorkoutTemplate(templateID int64) (models.WorkoutTemplate, error) {
	ctx := context.Background()
	var template models.WorkoutTemplate
	err := db.db.QueryRowContext(ctx,
		`SELECT template_id, name, description, created_at FROM workout_templates WHERE template_id = ?`,
		templateID,
	).Scan(&template.ID, &template.Name, &template.Description, &template.CreatedAt)
	return template, err
}

// AddTemplateExercise adds an exercise to a workout template
func (db *DB) AddTemplateExercise(templateID, exerciseID int64, orderIndex, plannedSets int, plannedReps string, targetPercentage int, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO template_exercises (template_id, exercise_id, order_index, planned_sets, planned_reps, target_percentage, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		templateID, exerciseID, orderIndex, plannedSets, plannedReps, targetPercentage, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetTemplateWithExercises gets a workout template with all its exercises
func (db *DB) GetTemplateWithExercises(templateID int64) (models.TemplateWithExercises, error) {
	ctx := context.Background()
	var template models.WorkoutTemplate
	var result models.TemplateWithExercises

	// Get template
	err := db.db.QueryRowContext(ctx,
		`SELECT template_id, name, description, created_at FROM workout_templates WHERE template_id = ?`,
		templateID,
	).Scan(&template.ID, &template.Name, &template.Description, &template.CreatedAt)
	if err != nil {
		return result, err
	}

	result.Template = template

	// Get template exercises
	rows, err := db.db.QueryContext(ctx, `
		SELECT template_exercise_id, exercise_id, order_index, notes
		FROM template_exercises
		WHERE template_id = ?
		ORDER BY order_index
	`, templateID)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var exercises []models.TemplateExercise
	for rows.Next() {
		var ex models.TemplateExercise
		if err := rows.Scan(&ex.ID, &ex.ExerciseID, &ex.OrderIndex, &ex.Notes); err != nil {
			return result, err
		}
		ex.TemplateID = templateID
		exercises = append(exercises, ex)
	}

	result.Exercises = exercises
	return result, nil
}

// DeleteWorkoutTemplate deletes a workout template and all related data
func (db *DB) DeleteWorkoutTemplate(templateID int64) error {
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete all template exercises
	_, err = tx.ExecContext(ctx, `DELETE FROM template_exercises WHERE template_id = ?`, templateID)
	if err != nil {
		return err
	}

	// Delete the template
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_templates WHERE template_id = ?`, templateID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AddPersonalRecord adds a new personal record
func (db *DB) AddPersonalRecord(exerciseID int64, recordType string, recordValue float64, reps int, weight float64, dateAchieved string, workoutID int64, notes string) (int64, error) {
	ctx := context.Background()
	result, err := db.db.ExecContext(ctx,
		`INSERT INTO personal_records (exercise_id, record_type, record_value, reps, weight, date_achieved, workout_id, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		exerciseID, recordType, recordValue, reps, weight, dateAchieved, workoutID, notes,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetPersonalRecordsForExercise gets all personal records for a specific exercise
func (db *DB) GetPersonalRecordsForExercise(exerciseID int64) ([]models.PersonalRecord, error) {
	ctx := context.Background()
	rows, err := db.db.QueryContext(ctx, `
		SELECT pr_id, workout_id, weight, date, notes, created_at
		FROM personal_records
		WHERE exercise_id = ?
		ORDER BY record_type, date_achieved DESC
	`, exerciseID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var records []models.PersonalRecord
	for rows.Next() {
		var pr models.PersonalRecord
		if err := rows.Scan(&pr.ID, &pr.WorkoutID, &pr.Weight, &pr.Date, &pr.Notes, &pr.CreatedAt); err != nil {
			return nil, err
		}
		pr.ExerciseID = exerciseID
		records = append(records, pr)
	}

	return records, nil
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
		SELECT workout_exercise_id FROM workout_exercises WHERE workout_id = ?
	`, workoutID)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

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
	if err := rows.Close(); err != nil {
		log.Printf("Error closing rows: %v", err)
	}

	// Delete all workout_exercises for this workout
	_, err = tx.ExecContext(ctx, `DELETE FROM workout_exercises WHERE workout_id = ?`, workoutID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
