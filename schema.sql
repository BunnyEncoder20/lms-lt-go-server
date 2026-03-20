-- schema.sql
-- USERS TABLES --
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY, -- Use TEXT for UUID strings
    pes_number TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL DEFAULT 'EMPLOYEE' CHECK (
        role IN ('ADMIN', 'MANAGER', 'EMPLOYEE', 'COURSE_DIRECTOR')
    ),
    cluster TEXT,

    -- Demographics
    title TEXT NOT NULL,
    gender TEXT NOT NULL CHECK (gender IN ('M', 'F')),
    band TEXT NOT NULL,
    grade TEXT NOT NULL,

    -- L&T Organizational Matrix
    ic TEXT NOT NULL,
    sbg TEXT NOT NULL,
    bu TEXT NOT NULL,
    segment TEXT NOT NULL,
    department TEXT NOT NULL,
    base_location TEXT NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT 1, -- SQLite uses 0/1 for boolean values
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Supervisory Relationships
    is_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    ns_id TEXT REFERENCES users (id) ON DELETE SET NULL,  -- Next supervisor
    dh_id TEXT REFERENCES users (id) ON DELETE SET NULL  -- Dept Head
);

-- Triggers to emulate Prisma's @updateAt behavior
CREATE TRIGGER update_users_updated_at AFTER
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
        category IN ('TECHNICAL', 'BEHAVIORAL')
    ),
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    location TEXT,
    virtual_link TEXT,
    pre_read_uri TEXT,
    created_by_id TEXT NOT NULL,
    deadline_days INTEGER NOT NULL DEFAULT 2,

    -- HR mappings & Category
    hr_program_id TEXT NOT NULL,
    mapped_category TEXT NOT NULL,
    mode_of_delivery TEXT NOT NULL CHECK (
        mode_of_delivery IN (
            'IN_PERSON', 'VIRTUAL_LINK', 'HYBRID', 'E_LEARNING'
        )
    ),

    -- Vendor & Instructor details
    instructor_name TEXT NOT NULL,
    institute_partner_name TEXT,
    process_owner_name TEXT,
    process_owner_email TEXT,

    -- Logistics & Metric
    duration_manhours REAL,
    training_mandays REAL,
    facility_id TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (created_by_id) REFERENCES users (id) ON DELETE SET NULL
);

-- TRAINING TABLE UPDATED_AT TRIGGER
CREATE TRIGGER update_trainings_updated_at AFTER
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
CREATE TRIGGER update_training_calendar_plans_updated_at AFTER
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
    status TEXT NOT NULL DEFAULT 'PENDING_MANAGER' CHECK (
        status IN (
            'PENDING_MANAGER', 'APPROVED', 'REJECTED', 'COMPLETED', 'ATTENDED'
        )
    ),
    user_id TEXT NOT NULL,
    training_id TEXT NOT NULL,
    nominated_by_id TEXT NOT NULL,

    -- Add HR completion tracking
    hr_completion_status TEXT,

    -- Cost tracking (using real for floating point currency)
    prof_fees REAL DEFAULT 0.0,
    venue_cost REAL DEFAULT 0.0,
    other_cost REAL DEFAULT 0.0,
    non_tems_travel REAL DEFAULT 0.0,
    non_tems_accommodation REAL DEFAULT 0.0,
    total_cost REAL DEFAULT 0.0,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- FOREIGN KEYS
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (training_id) REFERENCES trainings (id) ON DELETE CASCADE,
    FOREIGN KEY (nominated_by_id) REFERENCES users (id) ON DELETE CASCADE,

    -- The @@unique(userId, trainingId)
    UNIQUE (user_id, training_id)

);

-- NOMINATIONS TABLE UPDATED_AT TRIGGER
CREATE TRIGGER update_nominations_updated_at AFTER
UPDATE ON nominations
FOR EACH ROW
WHEN old.updated_at = new.updated_at
BEGIN
    UPDATE nominations SET updated_at = CURRENT_TIMESTAMP
    WHERE id = old.id;
END;

-- TABLE COURSE --
CREATE TABLE IF NOT EXISTS courses (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    description TEXT,
    author_id TEXT,
    cover_image_uri TEXT,
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (
        status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED')
    ),
    category TEXT NOT NULL CHECK (
        category IN ('TECHNICAL', 'BEHAVIORAL')
    ),
    estimated_durations INTEGER,
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
CREATE TRIGGER update_courses_updated_at AFTER
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
CREATE TRIGGER update_course_modules_updated_at
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
        content_type IN ('VIDEO', 'AUDIO', 'PDF', 'IMAGE', 'RICH_TEXT')
    ),
    asset_uri TEXT,
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
CREATE TRIGGER update_lessons_updated_at
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
    progress_percentage FLOAT NOT NULL DEFAULT 0,
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
