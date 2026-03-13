-- USERS
-- name: CreateUser :one
INSERT INTO users (
    id, pes_number, password, first_name, last_name,
    email, role, cluster, manager_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY last_name, first_name;

-- name: UpdateUserStatus :exec
UPDATE users
SET is_active = ?
WHERE id = ?;

-- TRAININGS
-- name: CreateTraining :one
INSERT INTO trainings (
    id, title, description, category, start_date, end_date,
    location, virtual_link, pre_read_uri, created_by_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;


-- name: ListActiveTrainings :many
SELECT * FROM trainings
WHERE is_active = 1
ORDER BY start_date ASC;

-- COURSE
-- name: CreateCourse :one
INSERT INTO courses (
    id, title, description, author_id, cover_image_uri, status,
    category, estimated_durations, learning_outcomes, is_strict_sequencing
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetCourseWithAuthor :one
-- Example of a join to handle that prisma 'author' relation
SELECT
    c.*,
    u.first_name AS author_first_name,
    u.last_name AS author_last_name
FROM courses c
JOIN users u ON c.author_id = u.id
WHERE c.id = ?
LIMIT 1;
