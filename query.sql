-- USERS
-- name: CreateUser :one
INSERT INTO users (
    id, pes_number, password, first_name, last_name,
    email, role, cluster, title, gender, band, grade,
    ic, sbg, bu, segment, department, base_location,
    is_id, ns_id, dh_id
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?
)
RETURNING *;

-- name: UpsertUser :one
INSERT INTO users (
    id, pes_number, password, first_name, last_name,
    email, role, cluster, title, gender, band, grade,
    ic, sbg, bu, segment, department, base_location,
    is_id, ns_id, dh_id
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?
)
ON CONFLICT (pes_number) DO UPDATE SET
    password = excluded.password,
    first_name = excluded.first_name,
    last_name = excluded.last_name,
    email = excluded.email,
    role = excluded.role,
    cluster = excluded.cluster,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = ?;

-- name: GetUserByPesNumber :one
SELECT *
FROM users
WHERE pes_number = ?;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY first_name, last_name;

-- name: GetTeamMembers :many
SELECT *
FROM users
WHERE is_id = ?;

-- name: UpdateUserStatus :exec
UPDATE users SET is_active = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: UpdateUser :one
UPDATE users SET
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    title = COALESCE(sqlc.narg('title'), title),
    department = COALESCE(sqlc.narg('department'), department),
    base_location = COALESCE(sqlc.narg('base_location'), base_location)
WHERE id = ?
RETURNING *;

-- TRAININGS
-- name: CreateTraining :one
INSERT INTO trainings (
    id, title, description, category, start_date, end_date,
    location, virtual_link, pre_read_uri, created_by_id,
    deadline_days, hr_program_id, mapped_category, mode_of_delivery,
    instructor_name, institute_partner_name, process_owner_name,
    process_owner_email, duration_manhours, training_mandays,
    facility_id
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?,
    ?
)
RETURNING *;

-- name: UpsertTraining :one
INSERT INTO trainings (
    id, title, description, category, start_date, end_date,
    location, virtual_link, pre_read_uri, created_by_id,
    deadline_days, hr_program_id, mapped_category, mode_of_delivery,
    instructor_name, institute_partner_name, process_owner_name,
    process_owner_email, duration_manhours, training_mandays,
    facility_id
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?,
    ?
)
ON CONFLICT (title) DO UPDATE SET
    description = excluded.description,
    category = excluded.category,
    start_date = excluded.start_date,
    end_date = excluded.end_date,
    location = excluded.location,
    virtual_link = excluded.virtual_link,
    pre_read_uri = excluded.pre_read_uri,
    deadline_days = excluded.deadline_days,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetTrainingByTitle :one
SELECT * FROM trainings
WHERE title = ?;

-- name: ListActiveTrainings :many
SELECT * FROM trainings
WHERE is_active = 1
ORDER BY start_date ASC;

-- name: GetTrainingByID :one
SELECT * FROM trainings
WHERE id = ?;

-- name: ListTrainings :many
SELECT * FROM trainings
ORDER BY start_date ASC;

-- name: ListTrainingsByCategory :many
SELECT * FROM trainings
WHERE category = ? AND is_active = 1
ORDER BY start_date ASC;

-- name: ListUpcomingTrainings :many
SELECT * FROM trainings
WHERE start_date > CURRENT_TIMESTAMP AND is_active = 1
ORDER BY start_date ASC;

-- name: UpdateTraining :one
UPDATE trainings SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    category = COALESCE(sqlc.narg('category'), category),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date),
    location = COALESCE(sqlc.narg('location'), location),
    virtual_link = COALESCE(sqlc.narg('virtual_link'), virtual_link),
    pre_read_uri = COALESCE(sqlc.narg('pre_read_uri'), pre_read_uri),
    deadline_days = COALESCE(sqlc.narg('deadline_days'), deadline_days),
    mapped_category = COALESCE(sqlc.narg('mapped_category'), mapped_category),
    mode_of_delivery = COALESCE(sqlc.narg('mode_of_delivery'), mode_of_delivery),
    instructor_name = COALESCE(sqlc.narg('instructor_name'), instructor_name),
    institute_partner_name = COALESCE(sqlc.narg('institute_partner_name'), institute_partner_name),
    process_owner_name = COALESCE(sqlc.narg('process_owner_name'), process_owner_name),
    process_owner_email = COALESCE(sqlc.narg('process_owner_email'), process_owner_email),
    duration_manhours = COALESCE(sqlc.narg('duration_manhours'), duration_manhours),
    training_mandays = COALESCE(sqlc.narg('training_mandays'), training_mandays),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTraining :exec
DELETE FROM trainings
WHERE id = ?;

-- NOMINATIONS
-- name: CreateNomination :one
INSERT INTO nominations (
    id, status, user_id, training_id, nominated_by_id,
    hr_completion_status, prof_fees, venue_cost, other_cost,
    non_tems_travel, non_tems_accommodation, total_cost
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?
)
RETURNING *;

-- name: GetNominationsByUserID :many
SELECT n.* FROM nominations n
JOIN trainings t ON n.training_id = t.id
WHERE n.user_id = ?
ORDER BY t.start_date ASC;

-- name: GetNominationsByTrainingID :many
SELECT * FROM nominations
WHERE training_id = ?;

-- name: UpsertNomination :one
INSERT INTO nominations (
    id, status, user_id, training_id, nominated_by_id,
    hr_completion_status, prof_fees, venue_cost, other_cost,
    non_tems_travel, non_tems_accommodation, total_cost
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?
)
ON CONFLICT (user_id, training_id) DO UPDATE SET
    status = excluded.status,
    nominated_by_id = excluded.nominated_by_id,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- COURSE
-- name: CreateCourse :one
INSERT INTO courses (
    id, title, description, author_id, cover_image_uri, status,
    category, estimated_durations, learning_outcomes, is_strict_sequencing,
    version, published_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpsertCourse :one
INSERT INTO courses (
    id, title, description, author_id, cover_image_uri, status,
    category, estimated_durations, learning_outcomes, is_strict_sequencing,
    version, published_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT (title) DO UPDATE SET
    description = excluded.description,
    author_id = excluded.author_id,
    status = excluded.status,
    category = excluded.category,
    estimated_durations = excluded.estimated_durations,
    learning_outcomes = excluded.learning_outcomes,
    is_strict_sequencing = excluded.is_strict_sequencing,
    version = excluded.version,
    published_at = excluded.published_at,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetCourseByTitle :one
SELECT * FROM courses
WHERE title = ?;

-- name: GetCourseWithAuthor :one
SELECT
    c.*,
    u.first_name AS author_first_name,
    u.last_name AS author_last_name
FROM courses c
JOIN users u ON c.author_id = u.id
WHERE c.id = ?
LIMIT 1;

-- COURSE MODULE
-- name: CreateCourseModule :one
INSERT INTO course_modules (
    id, title, course_id, description, sequence_order
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpsertCourseModule :one
INSERT INTO course_modules (
    id, title, course_id, description, sequence_order
) VALUES (
    ?, ?, ?, ?, ?
)
ON CONFLICT (course_id, title) DO UPDATE SET
    description = excluded.description,
    sequence_order = excluded.sequence_order,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- LESSON
-- name: CreateLesson :one
INSERT INTO lessons (
    id, title, content_type, asset_uri, rich_text_content,
    duration_minutes, sequence_order, module_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpsertLesson :one
INSERT INTO lessons (
    id, title, content_type, asset_uri, rich_text_content,
    duration_minutes, sequence_order, module_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT (module_id, title) DO UPDATE SET
    content_type = excluded.content_type,
    asset_uri = excluded.asset_uri,
    rich_text_content = excluded.rich_text_content,
    duration_minutes = excluded.duration_minutes,
    sequence_order = excluded.sequence_order,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- COURSE ASSIGNMENT
-- name: CreateCourseAssignment :one
INSERT INTO course_assignments (
    id, status, progress_percentage, course_version, due_date,
    course_id, user_id, assigned_by_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpsertCourseAssignment :one
INSERT INTO course_assignments (
    id, status, progress_percentage, course_version, due_date,
    course_id, user_id, assigned_by_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT (user_id, course_id) DO UPDATE SET
    status = excluded.status,
    progress_percentage = excluded.progress_percentage,
    course_version = excluded.course_version,
    due_date = excluded.due_date,
    assigned_by_id = excluded.assigned_by_id
RETURNING *;

-- LESSON PROGRESS
-- name: UpsertLessonProgress :one
INSERT INTO lesson_progress (
    id, is_completed, last_playback_position, completed_at,
    assignment_id, lesson_id
) VALUES (
    ?, ?, ?, ?, ?, ?
)
ON CONFLICT (assignment_id, lesson_id) DO UPDATE SET
    is_completed = excluded.is_completed,
    completed_at = excluded.completed_at,
    last_playback_position = excluded.last_playback_position
RETURNING *;

-- name: ListLessonsByCourse :many
SELECT l.* FROM lessons l
JOIN course_modules m ON l.module_id = m.id
WHERE m.course_id = ?
ORDER BY m.sequence_order, l.sequence_order;
