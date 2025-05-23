-- ================================
-- DATABASE RESET SCRIPT
-- ================================
-- This script drops all tables and recreates the schema from scratch
-- WARNING: All data will be lost when running this script

PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

-- Drop existing tables if they exist
DROP TABLE IF EXISTS personal_records;
DROP TABLE IF EXISTS onerm_history;
DROP TABLE IF EXISTS template_exercises;
DROP TABLE IF EXISTS workout_templates;
DROP TABLE IF EXISTS sets;
DROP TABLE IF EXISTS workout_exercises;
DROP TABLE IF EXISTS workouts;
DROP TABLE IF EXISTS exercises;

-- Drop old tables if they exist (from previous schema)
DROP TABLE IF EXISTS one_rep_max;
DROP TABLE IF EXISTS sets_old;
DROP TABLE IF EXISTS workout_exercises_old;
DROP TABLE IF EXISTS workouts_old;
DROP TABLE IF EXISTS exercises_old;

-- Create new tables

-- Exercises: Master list of all exercises
CREATE TABLE exercises (
    exercise_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    current_1rm REAL,  -- Current one-rep maximum in kg
    muscle_group TEXT, -- e.g., "Chest", "Back", "Legs"
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Workouts: Individual workout sessions
CREATE TABLE workouts (
    workout_id INTEGER PRIMARY KEY,
    date TEXT NOT NULL,  -- YYYY-MM-DD format
    name TEXT,
    template_id INTEGER, -- Reference to template if used
    status TEXT DEFAULT 'completed', -- 'planned', 'in_progress', 'completed'
    duration_minutes INTEGER, -- Total workout duration
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES workout_templates (template_id)
);

-- Junction table linking workouts to exercises
CREATE TABLE workout_exercises (
    workout_exercise_id INTEGER PRIMARY KEY,
    workout_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    order_index INTEGER DEFAULT 0, -- Order of exercise in workout
    notes TEXT,
    FOREIGN KEY (workout_id) REFERENCES workouts (workout_id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
);

-- Individual sets within workout exercises
CREATE TABLE sets (
    set_id INTEGER PRIMARY KEY,
    workout_exercise_id INTEGER NOT NULL,
    set_number INTEGER NOT NULL, -- 1, 2, 3, etc.
    set_range TEXT, -- e.g. "1-3", "4-5" (optional grouping)
    reps_planned INTEGER, -- Planned reps
    reps_completed INTEGER, -- Actual reps performed
    percentage_1rm INTEGER, -- Stored as integer (e.g., 70 for 70%)
    weight_used REAL NOT NULL, -- Weight in kg
    rpe INTEGER, -- Rate of Perceived Exertion (1-10)
    completed BOOLEAN DEFAULT 1,
    rest_seconds INTEGER, -- Rest time after this set
    notes TEXT,
    FOREIGN KEY (workout_exercise_id) REFERENCES workout_exercises (workout_exercise_id) ON DELETE CASCADE
);

-- Workout templates for reusable workout plans
CREATE TABLE workout_templates (
    template_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT, -- e.g., "Push", "Pull", "Legs", "Full Body"
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Exercises within workout templates
CREATE TABLE template_exercises (
    template_exercise_id INTEGER PRIMARY KEY,
    template_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    order_index INTEGER DEFAULT 0,
    planned_sets INTEGER,
    planned_reps TEXT, -- e.g., "8-12", "5x5"
    target_percentage INTEGER, -- Target % of 1RM
    notes TEXT,
    FOREIGN KEY (template_id) REFERENCES workout_templates (template_id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
);

-- Track 1RM changes over time
CREATE TABLE onerm_history (
    history_id INTEGER PRIMARY KEY,
    exercise_id INTEGER NOT NULL,
    date TEXT NOT NULL, -- YYYY-MM-DD format
    onerm_value REAL NOT NULL,
    method TEXT, -- How 1RM was determined: 'calculated', 'tested', 'estimated'
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
);

-- Personal records for various rep ranges and metrics
CREATE TABLE personal_records (
    pr_id INTEGER PRIMARY KEY,
    exercise_id INTEGER NOT NULL,
    record_type TEXT NOT NULL, -- e.g., '1RM', '3RM', '5RM', 'volume', 'max_reps'
    record_value REAL NOT NULL, -- The record value
    reps INTEGER, -- Number of reps (for rep-based records)
    weight REAL, -- Weight used (for weight-based records)
    date_achieved TEXT NOT NULL, -- YYYY-MM-DD format
    workout_id INTEGER, -- Link to the workout where PR was achieved
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id),
    FOREIGN KEY (workout_id) REFERENCES workouts (workout_id)
);

-- Create indices
CREATE INDEX idx_workout_date ON workouts(date);
CREATE INDEX idx_workout_template ON workouts(template_id);
CREATE INDEX idx_workout_status ON workouts(status);
CREATE INDEX idx_workout_exercises_workout ON workout_exercises(workout_id);
CREATE INDEX idx_workout_exercises_exercise ON workout_exercises(exercise_id);
CREATE INDEX idx_sets_workout_exercise ON sets(workout_exercise_id);
CREATE INDEX idx_sets_completed ON sets(completed);
CREATE INDEX idx_template_exercises_template ON template_exercises(template_id);
CREATE INDEX idx_template_exercises_exercise ON template_exercises(exercise_id);
CREATE INDEX idx_onerm_exercise ON onerm_history(exercise_id);
CREATE INDEX idx_onerm_date ON onerm_history(date);
CREATE INDEX idx_pr_exercise ON personal_records(exercise_id);
CREATE INDEX idx_pr_type ON personal_records(record_type);
CREATE INDEX idx_pr_date ON personal_records(date_achieved);
CREATE INDEX idx_exercise_muscle_group ON exercises(muscle_group);
CREATE UNIQUE INDEX idx_unique_exercise_name ON exercises(name);

COMMIT;

PRAGMA foreign_keys = ON;

-- Final message to display to the user
SELECT 'Database reset successfully with new schema' as message;