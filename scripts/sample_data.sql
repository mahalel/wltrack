-- ================================
-- WEIGHTLIFTING APP SAMPLE DATA
-- ================================

-- Insert sample exercises
INSERT INTO exercises (name, muscle_group, current_1rm, notes) VALUES 
('Bench Press', 'Chest', 100.0, 'Barbell bench press'),
('Squat', 'Legs', 140.0, 'Barbell back squat'),
('Deadlift', 'Back', 160.0, 'Conventional deadlift'),
('Overhead Press', 'Shoulders', 70.0, 'Standing barbell press'),
('Pull-up', 'Back', NULL, 'Bodyweight exercise'),
('Romanian Deadlift', 'Legs', 120.0, 'Great hamstring exercise'),
('Dumbbell Row', 'Back', NULL, 'Use proper form'),
('Incline Bench Press', 'Chest', 85.0, 'Target upper chest'),
('Front Squat', 'Legs', 110.0, 'More quad-focused than back squat'),
('Lateral Raise', 'Shoulders', NULL, 'Isolate lateral deltoids');

-- Insert sample workout templates
INSERT INTO workout_templates (name, description, category) VALUES 
('Push Day', 'Chest, shoulders, and triceps focused workout', 'Push'),
('Pull Day', 'Back and biceps focused workout', 'Pull'),
('Leg Day', 'Lower body focused workout', 'Legs'),
('Full Body A', 'First day of full body split', 'Full Body'),
('Full Body B', 'Second day of full body split', 'Full Body');

-- Link exercises to templates
-- Push Day
INSERT INTO template_exercises (template_id, exercise_id, order_index, planned_sets, planned_reps, target_percentage, notes) VALUES 
(1, 1, 1, 4, '5', 80, 'Focus on chest contraction'),  -- Bench Press
(1, 4, 2, 3, '8', 75, 'Maintain tight core'), -- Overhead Press
(1, 8, 3, 3, '10', 70, 'Keep shoulders retracted'); -- Incline Bench

-- Pull Day
INSERT INTO template_exercises (template_id, exercise_id, order_index, planned_sets, planned_reps, target_percentage, notes) VALUES 
(2, 3, 1, 4, '5', 80, 'Focus on bracing core'),  -- Deadlift
(2, 5, 2, 3, '8-12', NULL, 'Full range of motion'), -- Pull-up
(2, 7, 3, 3, '10-12', NULL, 'Retract scapula at top'); -- Dumbbell Row

-- Leg Day
INSERT INTO template_exercises (template_id, exercise_id, order_index, planned_sets, planned_reps, target_percentage, notes) VALUES 
(3, 2, 1, 4, '5', 80, 'Hit proper depth'),  -- Squat
(3, 6, 2, 3, '10', 70, 'Feel the hamstrings'), -- Romanian Deadlift
(3, 9, 3, 3, '8', 75, 'Keep elbows high'); -- Front Squat

-- Insert sample workouts
INSERT INTO workouts (date, name, template_id, status, duration_minutes, notes) VALUES 
('2024-01-15', 'Morning Push', 1, 'completed', 65, 'Good energy today'),
('2024-01-17', 'Back and Bis', 2, 'completed', 75, 'New PR on deadlift'),
('2024-01-19', 'Leg Day', 3, 'completed', 80, 'Quads still sore from last week'),
('2024-01-22', 'Push Session', 1, 'completed', 60, 'Felt strong on bench today'),
('2024-01-24', 'Pull Session', 2, 'completed', 70, 'Focused on form'),
('2024-01-26', 'Lower Body', 3, 'completed', 85, 'Increased weight on all exercises'),
('2024-01-29', 'Push Day', 1, 'planned', NULL, 'Try to increase bench weight');

-- Add exercises to completed workouts
-- Workout 1 (Push - Jan 15)
INSERT INTO workout_exercises (workout_id, exercise_id, order_index, notes) VALUES 
(1, 1, 1, 'Felt strong today'),  -- Bench Press
(1, 4, 2, 'Shoulders were fatigued'), -- Overhead Press
(1, 8, 3, 'Used dumbbells instead of barbell'); -- Incline Bench Press

-- Workout 2 (Pull - Jan 17)
INSERT INTO workout_exercises (workout_id, exercise_id, order_index, notes) VALUES 
(2, 3, 1, 'New PR!'),  -- Deadlift
(2, 5, 2, 'Did 3 extra reps on last set'), -- Pull-up
(2, 7, 3, 'Increased weight from last week'); -- Dumbbell Row

-- Workout 3 (Legs - Jan 19)
INSERT INTO workout_exercises (workout_id, exercise_id, order_index, notes) VALUES 
(3, 2, 1, 'Focused on depth'),  -- Squat
(3, 6, 2, 'Hamstrings felt tight'), -- Romanian Deadlift
(3, 9, 3, 'Kept elbows high'); -- Front Squat

-- Add sets for workout 1 (Push - Jan 15)
-- Bench Press sets
INSERT INTO sets (workout_exercise_id, set_number, reps_planned, reps_completed, percentage_1rm, weight_used, rpe, completed, rest_seconds) VALUES 
(1, 1, 5, 5, 70, 70.0, 7, 1, 180),
(1, 2, 5, 5, 75, 75.0, 8, 1, 180),
(1, 3, 5, 5, 80, 80.0, 8, 1, 180),
(1, 4, 5, 4, 85, 85.0, 9, 1, 180);

-- Overhead Press sets
INSERT INTO sets (workout_exercise_id, set_number, reps_planned, reps_completed, percentage_1rm, weight_used, rpe, completed, rest_seconds) VALUES 
(2, 1, 8, 8, 65, 45.0, 7, 1, 120),
(2, 2, 8, 8, 70, 50.0, 8, 1, 120),
(2, 3, 8, 7, 75, 52.5, 9, 1, 120);

-- Incline Bench Press sets
INSERT INTO sets (workout_exercise_id, set_number, reps_planned, reps_completed, percentage_1rm, weight_used, rpe, completed, rest_seconds) VALUES 
(3, 1, 10, 10, 65, 55.0, 7, 1, 90),
(3, 2, 10, 10, 70, 60.0, 8, 1, 90),
(3, 3, 10, 8, 70, 60.0, 9, 1, 90);

-- Add sets for workout 2 (Pull - Jan 17)
-- Deadlift sets (with PR)
INSERT INTO sets (workout_exercise_id, set_number, reps_planned, reps_completed, percentage_1rm, weight_used, rpe, completed, rest_seconds) VALUES 
(4, 1, 5, 5, 70, 112.0, 7, 1, 180),
(4, 2, 5, 5, 80, 128.0, 8, 1, 180),
(4, 3, 5, 5, 85, 136.0, 9, 1, 180),
(4, 4, 1, 1, 100, 160.0, 10, 1, 300);  -- PR set

-- Pull-up sets
INSERT INTO sets (workout_exercise_id, set_number, reps_planned, reps_completed, percentage_1rm, weight_used, rpe, completed, rest_seconds) VALUES 
(5, 1, 8, 8, NULL, 0.0, 7, 1, 120),  -- Bodyweight
(5, 2, 8, 8, NULL, 0.0, 8, 1, 120),
(5, 3, 8, 11, NULL, 0.0, 9, 1, 120);  -- Extra reps

-- Add 1RM history records
INSERT INTO onerm_history (exercise_id, date, onerm_value, method, notes) VALUES 
(1, '2023-12-01', 95.0, 'tested', 'Initial testing'),
(1, '2024-01-15', 100.0, 'tested', 'Tested at end of workout'),
(2, '2023-12-05', 130.0, 'tested', 'Initial testing'),
(2, '2024-01-19', 140.0, 'tested', 'Feeling stronger'),
(3, '2023-12-10', 150.0, 'tested', 'Initial testing'),
(3, '2024-01-17', 160.0, 'tested', 'New PR during workout'),
(4, '2023-12-15', 65.0, 'tested', 'Initial testing'),
(4, '2024-01-15', 70.0, 'calculated', 'Calculated from 8 reps at 55kg');

-- Add personal records
INSERT INTO personal_records (exercise_id, record_type, record_value, reps, weight, date_achieved, workout_id, notes) VALUES 
(1, '1RM', 100.0, 1, 100.0, '2024-01-15', 1, 'First time hitting 100kg'),
(1, '5RM', 85.0, 5, 85.0, '2024-01-15', 1, 'Strong set'),
(2, '1RM', 140.0, 1, 140.0, '2024-01-19', 3, 'Good depth'),
(2, '5RM', 125.0, 5, 125.0, '2024-01-19', 3, 'Felt solid'),
(3, '1RM', 160.0, 1, 160.0, '2024-01-17', 2, 'Great form'),
(3, '5RM', 136.0, 5, 136.0, '2024-01-17', 2, 'Last warmup set before PR'),
(4, '1RM', 70.0, 1, 70.0, '2024-01-15', 1, 'Strict form'),
(4, '8RM', 52.5, 8, 52.5, '2024-01-15', 1, 'Good volume PR');