-- ================================
-- DATABASE MIGRATION SCRIPT
-- ================================
-- This script migrates from the old schema to the new schema
-- It preserves existing data where possible

PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

-- Rename existing tables to keep data
ALTER TABLE exercises RENAME TO exercises_old;
ALTER TABLE one_rep_max RENAME TO one_rep_max_old;
ALTER TABLE workouts RENAME TO workouts_old;
ALTER TABLE workout_exercises RENAME TO workout_exercises_old;
ALTER TABLE sets RENAME TO sets_old;

-- Create new tables

-- Exercises table
CREATE TABLE exercises (
    exercise_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    current_1rm REAL,
    muscle_group TEXT,
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Workouts table
CREATE TABLE workouts (
    workout_id INTEGER PRIMARY KEY,
    date TEXT NOT NULL,
    name TEXT,
    template_id INTEGER,
    status TEXT DEFAULT 'completed',
    duration_minutes INTEGER,
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES workout_templates (template_id)
);

-- Workout exercises table
CREATE TABLE workout_exercises (
    workout_exercise_id INTEGER PRIMARY KEY,
    workout_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    order_index INTEGER DEFAULT 0,
    notes TEXT,
    FOREIGN KEY (workout_id) REFERENCES workouts (workout_id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
);

-- Sets table
CREATE TABLE sets (
    set_id INTEGER PRIMARY KEY,
    workout_exercise_id INTEGER NOT NULL,
    set_number INTEGER NOT NULL,
    set_range TEXT,
    reps_planned INTEGER,
    reps_completed INTEGER,
    percentage_1rm INTEGER,
    weight_used REAL NOT NULL,
    rpe INTEGER,
    completed BOOLEAN DEFAULT 1,
    rest_seconds INTEGER,
    notes TEXT,
    FOREIGN KEY (workout_exercise_id) REFERENCES workout_exercises (workout_exercise_id) ON DELETE CASCADE
);

-- New tables
CREATE TABLE workout_templates (
    template_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE template_exercises (
    template_exercise_id INTEGER PRIMARY KEY,
    template_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    order_index INTEGER DEFAULT 0,
    planned_sets INTEGER,
    planned_reps TEXT,
    target_percentage INTEGER,
    notes TEXT,
    FOREIGN KEY (template_id) REFERENCES workout_templates (template_id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
);

CREATE TABLE onerm_history (
    history_id INTEGER PRIMARY KEY,
    exercise_id INTEGER NOT NULL,
    date TEXT NOT NULL,
    onerm_value REAL NOT NULL,
    method TEXT,
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id)
);

CREATE TABLE personal_records (
    pr_id INTEGER PRIMARY KEY,
    exercise_id INTEGER NOT NULL,
    record_type TEXT NOT NULL,
    record_value REAL NOT NULL,
    reps INTEGER,
    weight REAL,
    date_achieved TEXT NOT NULL,
    workout_id INTEGER,
    notes TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercises (exercise_id),
    FOREIGN KEY (workout_id) REFERENCES workouts (workout_id)
);

-- Migrate data from old tables to new tables

-- Migrate exercises (with muscle_group as NULL initially)
INSERT INTO exercises (exercise_id, name, notes, created_at)
SELECT id, name, description, created_at FROM exercises_old;

-- Migrate one_rep_max to two places:
-- 1. Update current_1rm in exercises
-- 2. Add to onerm_history
WITH LatestOneRepMax AS (
    SELECT exercise_id, MAX(date) as latest_date
    FROM one_rep_max_old
    GROUP BY exercise_id
)
UPDATE exercises
SET current_1rm = (
    SELECT orm.weight
    FROM one_rep_max_old orm
    JOIN LatestOneRepMax lrm ON orm.exercise_id = lrm.exercise_id AND orm.date = lrm.latest_date
    WHERE orm.exercise_id = exercises.exercise_id
);

-- Migrate all 1RM history
INSERT INTO onerm_history (exercise_id, date, onerm_value, method, created_at)
SELECT exercise_id, date, weight, 'unknown', date FROM one_rep_max_old;

-- Migrate workouts
INSERT INTO workouts (workout_id, date, notes, created_at, status)
SELECT id, date, notes, created_at, 'completed' FROM workouts_old;

-- Migrate workout_exercises
INSERT INTO workout_exercises (workout_exercise_id, workout_id, exercise_id, notes, created_at)
SELECT id, workout_id, exercise_id, notes, created_at FROM workout_exercises_old;

-- Migrate sets
INSERT INTO sets (set_id, workout_exercise_id, set_number, reps_completed, percentage_1rm, weight_used, completed, set_range)
SELECT id, workout_exercise_id, set_order, reps, percentage_of_max, weight, 1, range_id
FROM sets_old;

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

-- Drop old tables
DROP TABLE exercises_old;
DROP TABLE one_rep_max_old;
DROP TABLE workouts_old;
DROP TABLE workout_exercises_old;
DROP TABLE sets_old;

COMMIT;

PRAGMA foreign_keys = ON;

-- Final message to display to the user
SELECT 'Migration completed successfully' as message;