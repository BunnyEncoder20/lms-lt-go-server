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

-- name: GetUserByID :one
SELECT
    id,
    pes_number,
    password,
    first_name,
    last_name,
    email,
    role,
    cluster,
    title,
    gender,
    band,
    grade,
    ic,
    sbg,
    bu,
    segment,
    department,
    base_location,
    is_id,
    ns_id,
    dh_id
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT
    id,
    pes_number,
    password,
    first_name,
    last_name,
    email,
    role,
    cluster,
    title,
    gender,
    band,
    grade,
    ic,
    sbg,
    bu,
    segment,
    department,
    base_location,
    is_id,
    ns_id,
    dh_id
FROM users
WHERE email = ?;

-- name: ListUsers :many
SELECT
    id,
    pes_number,
    password,
    first_name,
    last_name,
    email,
    role,
    cluster,
    title,
    gender,
    band,
    grade,
    ic,
    sbg,
    bu,
    segment,
    department,
    base_location,
    is_id,
    ns_id,
    dh_id
FROM users
ORDER BY first_name, last_name;

-- name: UpdateUserStatus :exec
UPDATE users
SET is_active = ?
WHERE id = ?;

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

-- name: ListActiveTrainings :many
SELECT
    id,
    title,
    description,
    category,
    start_date,
    end_date,
    location,
    virtual_link,
    pre_read_uri,
    created_by_id,
    deadline_days,
    hr_program_id,
    mapped_category,
    mode_of_delivery,
    instructor_name,
    institute_partner_name,
    process_owner_name,
    process_owner_email,
    duration_manhours,
    training_mandays,
    facility_id,
    is_active
FROM trainings
WHERE is_active = 1
ORDER BY start_date ASC;

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
