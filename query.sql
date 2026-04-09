-- USERS
-- name: CreateUser :one
INSERT INTO users (
    id, pes_number, password, first_name, last_name, full_name,
    email, role, cluster, location, title, gender, band, grade,
    employment_status, is_psn, is_name, ns_psn, ns_name, dh_psn, dh_name,
    ic, sbg, bu, segment, department, base_location,
    manager_id, skip_manager_id, is_id, ns_id, dh_id
) VALUES (
    sqlc.arg('id'), sqlc.arg('pes_number'), sqlc.arg('password'), sqlc.arg('first_name'), sqlc.arg('last_name'), sqlc.arg('full_name'),
    sqlc.arg('email'), sqlc.arg('role'), sqlc.arg('cluster'), sqlc.arg('location'), sqlc.arg('title'), sqlc.arg('gender'), sqlc.arg('band'), sqlc.arg('grade'),
    sqlc.arg('employment_status'), sqlc.arg('is_psn'), sqlc.arg('is_name'), sqlc.arg('ns_psn'), sqlc.arg('ns_name'), sqlc.arg('dh_psn'), sqlc.arg('dh_name'),
    sqlc.arg('ic'), sqlc.arg('sbg'), sqlc.arg('bu'), sqlc.arg('segment'), sqlc.arg('department'), sqlc.arg('base_location'),
    NULLIF(sqlc.arg('manager_id'), '00000000-0000-0000-0000-000000000000'), 
    NULLIF(sqlc.arg('skip_manager_id'), '00000000-0000-0000-0000-000000000000'), 
    sqlc.arg('is_id'), sqlc.arg('ns_id'), sqlc.arg('dh_id')
)
RETURNING *;

-- name: UpsertUser :one
INSERT INTO users (
    id, pes_number, password, first_name, last_name, full_name,
    email, role, cluster, location, title, gender, band, grade,
    employment_status, is_psn, is_name, ns_psn, ns_name, dh_psn, dh_name,
    ic, sbg, bu, segment, department, base_location,
    manager_id, skip_manager_id, is_id, ns_id, dh_id
) VALUES (
    sqlc.arg('id'), sqlc.arg('pes_number'), sqlc.arg('password'), sqlc.arg('first_name'), sqlc.arg('last_name'), sqlc.arg('full_name'),
    sqlc.arg('email'), sqlc.arg('role'), sqlc.arg('cluster'), sqlc.arg('location'), sqlc.arg('title'), sqlc.arg('gender'), sqlc.arg('band'), sqlc.arg('grade'),
    sqlc.arg('employment_status'), sqlc.arg('is_psn'), sqlc.arg('is_name'), sqlc.arg('ns_psn'), sqlc.arg('ns_name'), sqlc.arg('dh_psn'), sqlc.arg('dh_name'),
    sqlc.arg('ic'), sqlc.arg('sbg'), sqlc.arg('bu'), sqlc.arg('segment'), sqlc.arg('department'), sqlc.arg('base_location'),
    NULLIF(sqlc.arg('manager_id'), '00000000-0000-0000-0000-000000000000'), 
    NULLIF(sqlc.arg('skip_manager_id'), '00000000-0000-0000-0000-000000000000'), 
    sqlc.arg('is_id'), sqlc.arg('ns_id'), sqlc.arg('dh_id')
)
ON CONFLICT (pes_number) DO UPDATE SET
    password = excluded.password,
    first_name = excluded.first_name,
    last_name = excluded.last_name,
    full_name = excluded.full_name,
    email = excluded.email,
    role = excluded.role,
    cluster = excluded.cluster,
    location = excluded.location,
    title = excluded.title,
    band = excluded.band,
    grade = excluded.grade,
    employment_status = excluded.employment_status,
    is_psn = excluded.is_psn,
    is_name = excluded.is_name,
    ns_psn = excluded.ns_psn,
    ns_name = excluded.ns_name,
    dh_psn = excluded.dh_psn,
    dh_name = excluded.dh_name,
    ic = excluded.ic,
    sbg = excluded.sbg,
    bu = excluded.bu,
    segment = excluded.segment,
    department = excluded.department,
    base_location = excluded.base_location,
    manager_id = NULLIF(excluded.manager_id, '00000000-0000-0000-0000-000000000000'),
    skip_manager_id = NULLIF(excluded.skip_manager_id, '00000000-0000-0000-0000-000000000000'),
    is_id = excluded.is_id,
    ns_id = excluded.ns_id,
    dh_id = excluded.dh_id,
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
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    title = COALESCE(sqlc.narg('title'), title),
    department = COALESCE(sqlc.narg('department'), department),
    base_location = COALESCE(sqlc.narg('base_location'), base_location),
    location = COALESCE(sqlc.narg('location'), location),
    employment_status = COALESCE(sqlc.narg('employment_status'), employment_status),
    manager_id = COALESCE(sqlc.narg('manager_id'), manager_id),
    skip_manager_id = COALESCE(sqlc.narg('skip_manager_id'), skip_manager_id),
    is_id = COALESCE(sqlc.narg('is_id'), is_id),
    ns_id = COALESCE(sqlc.narg('ns_id'), ns_id),
    dh_id = COALESCE(sqlc.narg('dh_id'), dh_id)
WHERE id = ?
RETURNING *;

-- TRAININGS
-- name: CreateTraining :one
INSERT INTO trainings (
    id, title, description, category, instructor_name,
    learning_outcomes, month_tag, start_date, end_date,
    start_time, end_time, timezone, format,
    registration_deadline, max_capacity, target_clusters,
    prerequisites_url, venue_cost, professional_fees, stationary_cost,
    status, location, virtual_link, pre_read_url,
    deadline_days, created_by_id, hr_program_id, mapped_category,
    mode_of_delivery, institute_partner_name, process_owner_name,
    process_owner_email, duration_manhours, training_mandays,
    facility_id
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?,
    ?
)
RETURNING *;

-- name: UpsertTraining :one
INSERT INTO trainings (
    id, title, description, category, instructor_name,
    learning_outcomes, month_tag, start_date, end_date,
    start_time, end_time, timezone, format,
    registration_deadline, max_capacity, target_clusters,
    prerequisites_url, venue_cost, professional_fees, stationary_cost,
    status, location, virtual_link, pre_read_url,
    deadline_days, created_by_id, hr_program_id, mapped_category,
    mode_of_delivery, institute_partner_name, process_owner_name,
    process_owner_email, duration_manhours, training_mandays,
    facility_id
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?,
    ?
)
ON CONFLICT (title) DO UPDATE SET
    description = excluded.description,
    category = excluded.category,
    instructor_name = excluded.instructor_name,
    learning_outcomes = excluded.learning_outcomes,
    month_tag = excluded.month_tag,
    start_date = excluded.start_date,
    end_date = excluded.end_date,
    start_time = excluded.start_time,
    end_time = excluded.end_time,
    timezone = excluded.timezone,
    format = excluded.format,
    registration_deadline = excluded.registration_deadline,
    max_capacity = excluded.max_capacity,
    target_clusters = excluded.target_clusters,
    prerequisites_url = excluded.prerequisites_url,
    venue_cost = excluded.venue_cost,
    professional_fees = excluded.professional_fees,
    stationary_cost = excluded.stationary_cost,
    status = excluded.status,
    location = excluded.location,
    virtual_link = excluded.virtual_link,
    pre_read_url = excluded.pre_read_url,
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

-- name: GetAllActiveAndUpcomingTrainings :many
SELECT * FROM trainings
WHERE start_date > CURRENT_TIMESTAMP OR is_active = 1
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
    category = COALESCE(NULLIF(sqlc.narg('category'), ''), category),
    instructor_name = COALESCE(sqlc.narg('instructor_name'), instructor_name),
    learning_outcomes = COALESCE(sqlc.narg('learning_outcomes'), learning_outcomes),
    month_tag = COALESCE(sqlc.narg('month_tag'), month_tag),
    start_date = COALESCE(sqlc.narg('start_date'), start_date),
    end_date = COALESCE(sqlc.narg('end_date'), end_date),
    start_time = COALESCE(sqlc.narg('start_time'), start_time),
    end_time = COALESCE(sqlc.narg('end_time'), end_time),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    format = COALESCE(sqlc.narg('format'), format),
    registration_deadline = COALESCE(sqlc.narg('registration_deadline'), registration_deadline),
    max_capacity = COALESCE(sqlc.narg('max_capacity'), max_capacity),
    target_clusters = COALESCE(sqlc.narg('target_clusters'), target_clusters),
    prerequisites_url = COALESCE(sqlc.narg('prerequisites_url'), prerequisites_url),
    venue_cost = COALESCE(sqlc.narg('venue_cost'), venue_cost),
    professional_fees = COALESCE(sqlc.narg('professional_fees'), professional_fees),
    stationary_cost = COALESCE(sqlc.narg('stationary_cost'), stationary_cost),
    status = COALESCE(sqlc.narg('status'), status),
    location = COALESCE(sqlc.narg('location'), location),
    virtual_link = COALESCE(sqlc.narg('virtual_link'), virtual_link),
    pre_read_url = COALESCE(sqlc.narg('pre_read_url'), pre_read_url),
    deadline_days = COALESCE(sqlc.narg('deadline_days'), deadline_days),
    mapped_category = COALESCE(sqlc.narg('mapped_category'), mapped_category),
    mode_of_delivery
    = COALESCE(NULLIF(sqlc.narg('mode_of_delivery'), ''), mode_of_delivery),
    institute_partner_name
    = COALESCE(sqlc.narg('institute_partner_name'), institute_partner_name),
    process_owner_name
    = COALESCE(sqlc.narg('process_owner_name'), process_owner_name),
    process_owner_email
    = COALESCE(sqlc.narg('process_owner_email'), process_owner_email),
    duration_manhours
    = COALESCE(sqlc.narg('duration_manhours'), duration_manhours),
    training_mandays
    = COALESCE(sqlc.narg('training_mandays'), training_mandays),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteTraining :exec
DELETE FROM trainings
WHERE id = ?;

-- NOMINATIONS
-- name: CreateNomination :one
INSERT INTO nominations (
    id, status, user_id, training_id, course_id, nominated_by_id,
    hr_completion_status, prof_fees, venue_cost, other_cost,
    non_tems_travel, non_tems_accommodation, total_cost
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?
)
RETURNING *;

-- name: GetNominationsByUserID :many
SELECT n.* FROM nominations n
LEFT JOIN trainings t ON n.training_id = t.id
WHERE n.user_id = ?
ORDER BY COALESCE(t.start_date, n.created_at) ASC;

-- name: GetNominationsByTrainingID :many
SELECT * FROM nominations
WHERE training_id = ?;

-- name: GetNominationByID :one
SELECT * FROM nominations
WHERE id = ?;

-- name: GetNominationsByManagerID :many
SELECT n.* FROM nominations n
JOIN users u ON n.user_id = u.id
WHERE u.is_id = ?
ORDER BY n.created_at DESC;

-- name: UpdateNominationStatus :one
UPDATE nominations SET
    status = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: ListNominationsByFilters :many
SELECT n.* FROM nominations n
WHERE
    (COALESCE(sqlc.narg('status'), '') = '' OR n.status = sqlc.narg('status'))
    AND (
        COALESCE(sqlc.narg('training_id'), '') = ''
        OR n.training_id = sqlc.narg('training_id')
    )
    AND (
        COALESCE(sqlc.narg('user_id'), '') = ''
        OR n.user_id = sqlc.narg('user_id')
    )
    AND (
        COALESCE(sqlc.narg('nominated_by_id'), '') = ''
        OR n.nominated_by_id = sqlc.narg('nominated_by_id')
    )
ORDER BY n.created_at DESC;

-- name: CountNominationsByStatus :one
SELECT
    COUNT(*) AS total_count,
    SUM(CASE WHEN status = 'PENDING_MANAGER_APPROVAL' THEN 1 ELSE 0 END)
        AS pending_count,
    SUM(CASE WHEN status = 'ENROLLED' THEN 1 ELSE 0 END) AS approved_count,
    SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_count,
    SUM(CASE WHEN status = 'ATTENDED' THEN 1 ELSE 0 END) AS attended_count
FROM nominations;

-- name: CountNominationsByUserID :one
SELECT
    COUNT(*) AS total_count,
    SUM(CASE WHEN status = 'PENDING_MANAGER_APPROVAL' THEN 1 ELSE 0 END)
        AS pending_count,
    SUM(CASE WHEN status = 'ENROLLED' THEN 1 ELSE 0 END) AS approved_count,
    SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_count,
    SUM(CASE WHEN status = 'ATTENDED' THEN 1 ELSE 0 END) AS attended_count
FROM nominations
WHERE user_id = ?;

-- name: CountTeamNominationsByManager :one
SELECT
    COUNT(*) AS total_count,
    SUM(CASE WHEN status = 'PENDING_MANAGER_APPROVAL' THEN 1 ELSE 0 END)
        AS pending_count,
    SUM(CASE WHEN status = 'ENROLLED' THEN 1 ELSE 0 END) AS approved_count,
    SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END) AS completed_count,
    SUM(CASE WHEN status = 'ATTENDED' THEN 1 ELSE 0 END) AS attended_count
FROM nominations n
JOIN users u ON n.user_id = u.id
WHERE u.is_id = ?;

-- name: CountTeamMembers :one
SELECT COUNT(*) FROM users
WHERE is_id = ?;

-- name: UpsertNomination :one
INSERT INTO nominations (
    id, status, user_id, training_id, course_id, nominated_by_id,
    hr_completion_status, prof_fees, venue_cost, other_cost,
    non_tems_travel, non_tems_accommodation, total_cost
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?
)
ON CONFLICT (user_id, training_id) DO UPDATE SET
    status = excluded.status,
    nominated_by_id = excluded.nominated_by_id,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- COURSE MODULE
-- name: CreateCourse :one
INSERT INTO courses (
    id, title, description, author_id, cover_image_url, status,
    category, estimated_duration, learning_outcomes, is_strict_sequencing,
    version, published_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpsertCourse :one
INSERT INTO courses (
    id, title, description, author_id, cover_image_url, status,
    category, estimated_duration, learning_outcomes, is_strict_sequencing,
    version, published_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT (title) DO UPDATE SET
    description = excluded.description,
    author_id = excluded.author_id,
    status = excluded.status,
    category = excluded.category,
    estimated_duration = excluded.estimated_duration,
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
LEFT JOIN users u ON c.author_id = u.id
WHERE c.id = ?
LIMIT 1;

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
    id, title, content_type, asset_url, rich_text_content,
    duration_minutes, sequence_order, module_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpsertLesson :one
INSERT INTO lessons (
    id, title, content_type, asset_url, rich_text_content,
    duration_minutes, sequence_order, module_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT (module_id, title) DO UPDATE SET
    content_type = excluded.content_type,
    asset_url = excluded.asset_url,
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

-- ADMIN STATS
-- name: GetAdminKpis :one
WITH
training_counts AS (
    SELECT COUNT(*) AS total_trainings
    FROM trainings
    WHERE is_active = 1
),

nomination_stats AS (
    SELECT
        COUNT(*) FILTER (WHERE status IN ('ENROLLED', 'COMPLETED', 'ATTENDED'))
            AS total_participants,
        COUNT(*) FILTER (WHERE status = 'COMPLETED') AS completed_count,
        COUNT(*) FILTER (WHERE status IN ('ENROLLED', 'COMPLETED', 'ATTENDED'))
            AS enrolled_count
    FROM nominations
),

mandays_calc AS (
    SELECT
        COALESCE(
            SUM((JULIANDAY(t.end_date) - JULIANDAY(t.start_date) + 1)), 0.0
        )
            AS total_man_days
    FROM trainings t
    JOIN nominations n ON t.id = n.training_id
    WHERE t.is_active = 1 AND n.status IN ('ENROLLED', 'COMPLETED', 'ATTENDED')
)

SELECT
    tc.total_trainings,
    ns.total_participants,
    ns.completed_count,
    ns.enrolled_count,
    mc.total_man_days
FROM training_counts tc, nomination_stats ns, mandays_calc mc;

-- name: GetMonthlyStats :many
SELECT
    STRFTIME('%Y-%m', t.start_date) AS month_key,
    STRFTIME('%b %y', t.start_date) AS month_label,
    COUNT(n.id) AS participant_count,
    COUNT(DISTINCT t.id) AS training_count
FROM nominations n
JOIN trainings t ON n.training_id = t.id
WHERE n.status IN ('ENROLLED', 'COMPLETED', 'ATTENDED')
GROUP BY month_key
ORDER BY month_key ASC;

-- name: GetCategoryDistribution :many
SELECT
    t.category AS name,
    COUNT(n.id) AS value
FROM nominations n
JOIN trainings t ON n.training_id = t.id
WHERE n.status IN ('ENROLLED', 'COMPLETED', 'ATTENDED')
GROUP BY t.category;

-- name: GetClusterStats :many
SELECT
    COALESCE(u.cluster, 'Unassigned') AS cluster,
    COUNT(u.id) AS total_employees,
    COUNT(DISTINCT CASE WHEN n.id IS NOT NULL THEN u.id END) AS trained,
    (
        COUNT(u.id) - COUNT(DISTINCT CASE WHEN n.id IS NOT NULL THEN u.id END)
    ) AS untrained
FROM users u
LEFT JOIN
    nominations n
    ON
        u.id = n.user_id AND n.status IN ('ENROLLED', 'COMPLETED', 'ATTENDED')
WHERE u.is_active = 1 AND u.role != 'ADMIN'
GROUP BY u.cluster;

-- name: DeleteAllHistoricalRecords :exec
DELETE FROM historical_training_records;

-- name: CreateHistoricalRecord :exec
INSERT INTO historical_training_records (
    id, program_id, program_title, mapped_category, cluster,
    employee_pes_no, employee_name, completion_status,
    mode_of_delivery, from_date, to_date, month_key,
    man_days, man_hours, total_cost_inr, source_file
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?, ?,
    ?, ?, ?, ?
);

-- TRAINING CALENDAR PLANS
-- name: CreateCalendarPlan :one
INSERT INTO training_calendar_plans (
    id, program_name, mapped_category, target_month,
    status, actual_training_id
) VALUES (
    ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: ListCalendarPlans :many
SELECT * FROM training_calendar_plans
ORDER BY target_month ASC;

-- MANAGER ALLOCATIONS
-- name: CreateManagerAllocation :one
INSERT INTO manager_allocations (
    id, training_id, course_id, manager_id, assigned_by_id
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING *;

-- name: ListManagerAllocationsByManager :many
SELECT * FROM manager_allocations
WHERE manager_id = ?;

-- ATTENDANCE DISPATCHES
-- name: CreateAttendanceDispatch :one
INSERT INTO attendance_dispatches (
    id, entity_type, expires_at, training_id, course_id, created_by_id
) VALUES (
    ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- ATTENDANCE REQUESTS
-- name: CreateAttendanceRequest :one
INSERT INTO attendance_requests (
    id, token_hash, dispatch_id, nomination_id, user_id
) VALUES (
    ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetAttendanceRequestByToken :one
SELECT * FROM attendance_requests
WHERE token_hash = ?;

-- name: ConfirmAttendanceRequest :one
UPDATE attendance_requests SET
    status = 'CONFIRMED',
    confirmed_at = CURRENT_TIMESTAMP,
    confirmed_ip = ?,
    confirmed_user_agent = ?
WHERE id = ?
RETURNING *;
