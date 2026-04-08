-- schema.sql
-- USERS TABLES --
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY, -- Use TEXT for UUID strings
    pes_number TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    full_name TEXT,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL DEFAULT 'EMPLOYEE' CHECK (
        role IN ('ADMIN', 'MANAGER', 'EMPLOYEE', 'COURSE_DIRECTOR')
    ),
    cluster TEXT,
    location TEXT,

    -- Demographics
    title TEXT NOT NULL,
    gender TEXT NOT NULL, -- CHECK (gender IN ('M', 'F'))
    band TEXT NOT NULL,
    grade TEXT NOT NULL,

    -- L&T Organizational Matrix / Status
    employment_status TEXT,
    is_psn TEXT,
    is_name TEXT,
    ns_psn TEXT,
    ns_name TEXT,
    dh_psn TEXT,
    dh_name TEXT,

    ic TEXT,
    sbg TEXT,
    bu TEXT,
    segment TEXT,
    department TEXT,
    base_location TEXT,

    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Supervisory Relationships (IDs)
    manager_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    skip_manager_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    is_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    ns_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    dh_id TEXT REFERENCES users (id) ON DELETE SET NULL
);

-- Triggers to emulate Prisma's @updateAt behavior
CREATE TRIGGER IF NOT EXISTS update_users_updated_at AFTER
UPDATE ON users FOR EACH ROW
WHEN old.updated_at = new.updated_at BEGIN
    UPDATE users
    SET
        updated_at = CURRENT_TIMESTAMP
    WHERE
        id = old.id;

END;

-- TRAINING TABLE --
CREATE TABLE IF NOT EXISTS trainings (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT NOT NULL CHECK (
        category IN (
            'TECHNICAL', 'IT_DIGITAL', 'QUALITY', 'SAFETY', 'BEHAVIORAL'
        )
    ),
    instructor_name TEXT,
    learning_outcomes TEXT, -- Store as JSON array string
    month_tag TEXT,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    start_time TEXT,
    end_time TEXT,
    timezone TEXT,
    format TEXT CHECK (
        format IN ('IN_PERSON', 'VIRTUAL', 'HYBRID')
    ),
    registration_deadline DATETIME,
    max_capacity INTEGER,
    target_clusters TEXT, -- Store as JSON array string
    prerequisites_url TEXT,
    venue_cost INTEGER DEFAULT 0,
    professional_fees INTEGER DEFAULT 0,
    stationary_cost INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (
        status IN ('DRAFT', 'SCHEDULED', 'PUBLISHED')
    ),
    location TEXT,
    virtual_link TEXT,
    pre_read_url TEXT,
    deadline_days INTEGER NOT NULL DEFAULT 2,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    created_by_id TEXT NOT NULL,

    -- Legacy / HR Mappings
    hr_program_id TEXT,
    mapped_category TEXT,
    mode_of_delivery TEXT,
    institute_partner_name TEXT,
    process_owner_name TEXT,
    process_owner_email TEXT,
    duration_manhours REAL,
    training_mandays REAL,
    facility_id TEXT,

    FOREIGN KEY (created_by_id) REFERENCES users (id) ON DELETE SET NULL
);

-- TRAINING TABLE UPDATED_AT TRIGGER
CREATE TRIGGER IF NOT EXISTS update_trainings_updated_at AFTER
UPDATE ON trainings
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE trainings SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;

-- TRAINING CALENDAR PLAN TABLE --
CREATE TABLE IF NOT EXISTS training_calendar_plans (
    id TEXT PRIMARY KEY,
    program_name TEXT NOT NULL,
    mapped_category TEXT NOT NULL,
    target_month TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PLANNED' CHECK (
        status IN ('PLANNED', 'FINALIZED', 'CANCELLED')
    ),
    actual_training_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (actual_training_id) REFERENCES trainings (
        id
    ) ON DELETE SET NULL
);

-- TRAINING CALENDAR PLAN TABLE UPDATED_AT TRIGGER
CREATE TRIGGER IF NOT EXISTS update_training_calendar_plans_updated_at AFTER
UPDATE ON training_calendar_plans
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE training_calendar_plans SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;


-- NOMINATION TABLE --
CREATE TABLE IF NOT EXISTS nominations (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'PENDING_EMPLOYEE_APPROVAL' CHECK (
        status IN (
            'PENDING_MANAGER_ASSIGNMENT',
            'PENDING_EMPLOYEE_APPROVAL',
            'ENROLLED',
            'PENDING_MANAGER_APPROVAL',
            'DECLINED',
            'REJECTED',
            'COMPLETED',
            'ATTENDED'
        )
    ),
    user_id TEXT NOT NULL,
    training_id TEXT,
    course_id TEXT,
    nominated_by_id TEXT NOT NULL,

    -- Add HR completion tracking
    hr_completion_status TEXT,

    -- Cost tracking (using integer for cents/paise)
    prof_fees INTEGER DEFAULT 0,
    venue_cost INTEGER DEFAULT 0,
    other_cost INTEGER DEFAULT 0,
    non_tems_travel INTEGER DEFAULT 0,
    non_tems_accommodation INTEGER DEFAULT 0,
    total_cost INTEGER DEFAULT 0,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (training_id) REFERENCES trainings (id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
    FOREIGN KEY (nominated_by_id) REFERENCES users (id) ON DELETE CASCADE,

    -- Unique constraints
    UNIQUE (user_id, training_id),
    UNIQUE (user_id, course_id)
);

-- NOMINATIONS TABLE UPDATED_AT TRIGGER
CREATE TRIGGER IF NOT EXISTS update_nominations_updated_at AFTER
UPDATE ON nominations
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE nominations SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;

-- COURSE TABLE --
CREATE TABLE IF NOT EXISTS courses (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    description TEXT,
    author_id TEXT,
    cover_image_url TEXT,
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (
        status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED')
    ),
    category TEXT NOT NULL CHECK (
        category IN (
            'TECHNICAL', 'IT_DIGITAL', 'QUALITY', 'SAFETY', 'BEHAVIORAL'
        )
    ),
    estimated_duration INTEGER,
    -- Store the array as json string for sqlite
    learning_outcomes TEXT NOT NULL,
    is_strict_sequencing BOOLEAN NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    published_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE SET NULL
);

-- COURSE TABLE UPDATED_AT TRIGGER
CREATE TRIGGER IF NOT EXISTS update_courses_updated_at AFTER
UPDATE ON courses
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE courses SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;

-- TABLE COURSE MODULE --
CREATE TABLE IF NOT EXISTS course_modules (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    course_id TEXT NOT NULL,
    description TEXT,
    sequence_order INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,

    UNIQUE (course_id, title)
);

-- COURSE MODULE TABLE UPDATED_AT TRIGGER
CREATE TRIGGER IF NOT EXISTS update_course_modules_updated_at
AFTER UPDATE ON course_modules
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE course_modules SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;


-- LESSON TABLE --
CREATE TABLE IF NOT EXISTS lessons (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    content_type TEXT NOT NULL CHECK (
        content_type IN (
            'VIDEO', 'AUDIO', 'PDF', 'IMAGE', 'RICH_TEXT', 'PRESENTATION'
        )
    ),
    asset_url TEXT,
    rich_text_content TEXT,
    duration_minutes INTEGER,
    sequence_order INTEGER NOT NULL,
    module_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (module_id) REFERENCES course_modules (id) ON DELETE CASCADE,

    UNIQUE (module_id, title)
);

-- LESSON TABLE UPDATED_AT TRIGGER
CREATE TRIGGER IF NOT EXISTS update_lessons_updated_at
AFTER UPDATE ON lessons
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE lessons SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;

-- COURSE ASSIGNMENTS TABLE --
CREATE TABLE IF NOT EXISTS course_assignments (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'NOT_STARTED' CHECK (
        status IN ('NOT_STARTED', 'IN_PROGRESS', 'COMPLETED')
    ),
    progress_percentage REAL NOT NULL DEFAULT 0,
    course_version INTEGER NOT NULL DEFAULT 1,
    due_date DATETIME,
    enrolled_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    course_id TEXT NOT NULL,
    user_id TEXT,
    assigned_by_id TEXT,

    -- FOREIGN KEYS
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL,
    FOREIGN KEY (assigned_by_id) REFERENCES users (id) ON DELETE SET NULL,

    UNIQUE (user_id, course_id)
);

-- LESSON PROGRESS TABLE --
CREATE TABLE IF NOT EXISTS lesson_progress (
    id TEXT PRIMARY KEY,
    is_completed BOOLEAN NOT NULL DEFAULT 0,
    last_playback_position INTEGER NOT NULL DEFAULT 0,
    completed_at DATETIME,
    assignment_id TEXT NOT NULL,
    lesson_id TEXT NOT NULL,

    -- FOREIGN KEYS
    FOREIGN KEY (assignment_id)
    REFERENCES course_assignments (id)
    ON DELETE CASCADE,

    FOREIGN KEY (lesson_id)
    REFERENCES lessons (id)
    ON DELETE CASCADE,

    UNIQUE (assignment_id, lesson_id)
);

-- MANAGER ALLOCATIONS TABLE --
CREATE TABLE IF NOT EXISTS manager_allocations (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    training_id TEXT,
    course_id TEXT,
    manager_id TEXT NOT NULL,
    assigned_by_id TEXT NOT NULL,

    FOREIGN KEY (training_id) REFERENCES trainings (id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
    FOREIGN KEY (manager_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by_id) REFERENCES users (id) ON DELETE CASCADE,

    UNIQUE (manager_id, training_id),
    UNIQUE (manager_id, course_id)
);

-- HISTORICAL TRAINING RECORDS TABLE --
CREATE TABLE IF NOT EXISTS historical_training_records (
    id TEXT PRIMARY KEY,
    program_id TEXT,
    program_title TEXT NOT NULL,
    mapped_category TEXT,
    cluster TEXT,
    employee_pes_no TEXT,
    employee_name TEXT,
    completion_status TEXT,
    mode_of_delivery TEXT,
    from_date DATETIME,
    to_date DATETIME,
    month_key TEXT NOT NULL,
    man_days REAL,
    man_hours REAL,
    total_cost_inr INTEGER,
    source_file TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_htr_month_key ON historical_training_records (
    month_key
);
CREATE INDEX IF NOT EXISTS idx_htr_category ON historical_training_records (
    mapped_category
);
CREATE INDEX IF NOT EXISTS idx_htr_cluster ON historical_training_records (
    cluster
);

-- ATTENDANCE DISPATCHES TABLE --
CREATE TABLE IF NOT EXISTS attendance_dispatches (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('TRAINING', 'COURSE')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    closed_at DATETIME,
    training_id TEXT,
    course_id TEXT,
    created_by_id TEXT NOT NULL,

    FOREIGN KEY (training_id) REFERENCES trainings (id) ON DELETE CASCADE,
    FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
    FOREIGN KEY (created_by_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_dispatch_entity_created ON attendance_dispatches (
    entity_type, created_at
);
CREATE INDEX IF NOT EXISTS idx_dispatch_training ON attendance_dispatches (
    training_id
);
CREATE INDEX IF NOT EXISTS idx_dispatch_course ON attendance_dispatches (
    course_id
);

-- ATTENDANCE REQUESTS TABLE --
CREATE TABLE IF NOT EXISTS attendance_requests (
    id TEXT PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'SENT' CHECK (
        status IN ('SENT', 'CONFIRMED', 'EXPIRED', 'VOID')
    ),
    sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at DATETIME,
    consumed_at DATETIME,
    confirmed_ip TEXT,
    confirmed_user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    dispatch_id TEXT NOT NULL,
    nomination_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    consumed_by_user_id TEXT,

    FOREIGN KEY (dispatch_id) REFERENCES attendance_dispatches (
        id
    ) ON DELETE CASCADE,
    FOREIGN KEY (nomination_id) REFERENCES nominations (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (consumed_by_user_id) REFERENCES users (id) ON DELETE SET NULL,

    UNIQUE (dispatch_id, nomination_id)
);

CREATE INDEX IF NOT EXISTS idx_req_dispatch ON attendance_requests (
    dispatch_id
);
CREATE INDEX IF NOT EXISTS idx_req_nomination ON attendance_requests (
    nomination_id
);
CREATE INDEX IF NOT EXISTS idx_req_user ON attendance_requests (user_id);
