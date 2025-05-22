# WLTrack Database Schema Documentation

This document provides a detailed overview of the WLTrack application's database schema, tables, and relationships.

## Schema Overview

The WLTrack database is designed to store comprehensive weightlifting data, including:

- Exercises
- Workout sessions
- Sets and repetitions
- One-rep maximum (1RM) history
- Personal records (PRs)
- Workout templates

## Core Tables

### Exercises Table

Stores the master list of all exercises.

| Column | Type | Description |
|--------|------|-------------|
| `exercise_id` | INTEGER | Primary key |
| `name` | TEXT | Exercise name (unique) |
| `current_1rm` | REAL | Current one-rep maximum in kg |
| `notes` | TEXT | Additional notes about the exercise |
| `created_at` | TEXT | Timestamp when the exercise was created |

### Workouts Table

Records individual workout sessions.

| Column | Type | Description |
|--------|------|-------------|
| `workout_id` | INTEGER | Primary key |
| `date` | TEXT | Date of the workout (YYYY-MM-DD format) |
| `template_id` | INTEGER | Reference to template if used (foreign key) |
| `notes` | TEXT | Workout notes or commentary |
| `created_at` | TEXT | Timestamp when the workout was created |

### Workout Exercises Table

Junction table linking workouts to exercises.

| Column | Type | Description |
|--------|------|-------------|
| `workout_exercise_id` | INTEGER | Primary key |
| `workout_id` | INTEGER | Foreign key referencing workouts |
| `exercise_id` | INTEGER | Foreign key referencing exercises |
| `order_index` | INTEGER | Order of exercise in workout |
| `notes` | TEXT | Notes specific to this exercise in this workout |
| `created_at` | TEXT | Timestamp when the workout exercise was created |

### Sets Table

Records individual sets within workout exercises.

| Column | Type | Description |
|--------|------|-------------|
| `set_id` | INTEGER | Primary key |
| `workout_exercise_id` | INTEGER | Foreign key referencing workout_exercises |
| `set_number` | INTEGER | Order of set within the exercise (1, 2, 3, etc.) |
| `set_range` | TEXT | Optional grouping (e.g., "1-3", "4-5") |
| `reps` | INTEGER | Number of repetitions performed |
| `percentage_1rm` | INTEGER | Percentage of one-rep max (70 for 70%) |
| `weight` | REAL | Weight used in kg |
| `notes` | TEXT | Notes specific to this set |

## Template Tables

### Workout Templates Table

Stores reusable workout plans.

| Column | Type | Description |
|--------|------|-------------|
| `template_id` | INTEGER | Primary key |
| `name` | TEXT | Template name |
| `description` | TEXT | Template description |
| `created_at` | TEXT | Timestamp when template was created |

### Template Exercises Table

Stores exercises within workout templates.

| Column | Type | Description |
|--------|------|-------------|
| `template_exercise_id` | INTEGER | Primary key |
| `template_id` | INTEGER | Foreign key referencing workout_templates |
| `exercise_id` | INTEGER | Foreign key referencing exercises |
| `order_index` | INTEGER | Order of exercise in template |
| `notes` | TEXT | Exercise-specific notes within this template |

## History and Records Tables

### 1RM History Table

Tracks changes in one-rep maximum over time.

| Column | Type | Description |
|--------|------|-------------|
| `history_id` | INTEGER | Primary key |
| `exercise_id` | INTEGER | Foreign key referencing exercises |
| `onerm_value` | REAL | The 1RM value in kg |
| `notes` | TEXT | Additional notes |
| `created_at` | TEXT | Timestamp when record was created |

### Personal Records Table

Stores personal records for various rep ranges and metrics.

| Column | Type | Description |
|--------|------|-------------|
| `pr_id` | INTEGER | Primary key |
| `exercise_id` | INTEGER | Foreign key referencing exercises |
| `workout_id` | INTEGER | Foreign key referencing the workout where PR was achieved |
| `weight` | REAL | Weight used in kg |
| `date` | TEXT | Date record was achieved (YYYY-MM-DD format) |
| `notes` | TEXT | Additional notes |
| `created_at` | TEXT | Timestamp when record was created |

## Performance Indices

The schema includes multiple indices to optimize query performance:

### Workout Indices
- `idx_workout_date`: Index on workouts(date)
- `idx_workout_template`: Index on workouts(template_id)

### Workout Exercises Indices
- `idx_workout_exercises_workout`: Index on workout_exercises(workout_id)
- `idx_workout_exercises_exercise`: Index on workout_exercises(exercise_id)

### Sets Indices
- `idx_sets_workout_exercise`: Index on sets(workout_exercise_id)

### Template Indices
- `idx_template_exercises_template`: Index on template_exercises(template_id)
- `idx_template_exercises_exercise`: Index on template_exercises(exercise_id)

### History and Records Indices
- `idx_onerm_exercise`: Index on onerm_history(exercise_id)
- `idx_pr_exercise`: Index on personal_records(exercise_id)
- `idx_pr_date`: Index on personal_records(date)

### Exercise Indices
- `idx_unique_exercise_name`: Unique index on exercises(name)

## Entity Relationships

The database uses foreign keys to enforce relationships between tables:

- **workouts → workout_templates**: A workout can be based on a template
- **workout_exercises → workouts**: Exercises belong to a specific workout (cascade delete)
- **workout_exercises → exercises**: Links exercises to workouts
- **sets → workout_exercises**: Sets belong to a specific exercise in a workout (cascade delete)
- **template_exercises → workout_templates**: Template exercises belong to a specific template (cascade delete)
- **template_exercises → exercises**: Links exercises to templates
- **onerm_history → exercises**: 1RM history belongs to a specific exercise
- **personal_records → exercises**: PRs belong to a specific exercise
- **personal_records → workouts**: PRs can be linked to a specific workout

## Helper Structures

The application uses several helper structures to combine related data:

- **ExerciseWithSets**: Combines an exercise with its workout-specific information and sets
- **WorkoutWithExercises**: Combines a workout with all its exercises and sets
- **TemplateWithExercises**: Combines a template with all its exercises

## Examples

### Example Query: Find a User's Recent Workouts

```sql
SELECT workout_id, date, template_id, notes, created_at
FROM workouts
ORDER BY date DESC
LIMIT 5;
```

### Example Query: Get Exercises in a Workout with Sets

```sql
SELECT e.name, we.order_index, s.set_number, s.reps, s.weight
FROM workout_exercises we
JOIN exercises e ON we.exercise_id = e.exercise_id
JOIN sets s ON s.workout_exercise_id = we.workout_exercise_id
WHERE we.workout_id = 1
ORDER BY we.order_index, s.set_number;
```

### Example Query: Find 1RM Progress for an Exercise

```sql
SELECT created_at, onerm_value
FROM onerm_history
WHERE exercise_id = 1
ORDER BY created_at;
```
