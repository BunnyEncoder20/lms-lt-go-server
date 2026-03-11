-- schema.sql
-- USERS TABLES --
CREATE TABLE users (
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
    is_active BOOLEAN NOT NULL DEFAULT 1, -- SQLite uses 0/1 for boolean values
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    manager_id TEXT REFERENCES users (id) ON DELETE SET NULL
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
CREATE TABLE trainings (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    location TEXT,
    virtual_link TEXT,
    pre_read_url TEXT,
    deadline_days INTEGER NOT NULL DEFAULT 2,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
