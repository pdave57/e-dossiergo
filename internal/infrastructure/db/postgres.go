// Package db provides PostgreSQL connectivity and schema management.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Open creates and validates a *sql.DB connection pool.
func Open(dsn string, maxOpen, maxIdle int, maxLife time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(maxLife)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	return db, nil
}

// Migrate runs all DDL statements to create or update the schema.
// Idempotent — safe to run on every startup.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

// schema contains the complete DDL for e-Dossier.
const schema = `
-- ─────────────────────────────────────────────
-- EXTENSIONS
-- ─────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─────────────────────────────────────────────
-- GEO / ADMINISTRATIVE
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS states (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    name        TEXT NOT NULL,
    code        TEXT NOT NULL UNIQUE,
    country     TEXT NOT NULL DEFAULT 'Nigeria',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT
);

CREATE TABLE IF NOT EXISTS zones (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, code)
);

CREATE TABLE IF NOT EXISTS lgas (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    zone_id     TEXT NOT NULL REFERENCES zones(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, code)
);

-- ─────────────────────────────────────────────
-- AUTH / RBAC
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id      TEXT NOT NULL REFERENCES states(id),
    school_id     TEXT,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'ACTIVE',
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    created_by    TEXT,
    updated_by    TEXT
);

CREATE INDEX IF NOT EXISTS idx_users_state_id ON users(state_id);
CREATE INDEX IF NOT EXISTS idx_users_school_id ON users(school_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TABLE IF NOT EXISTS roles (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, code)
);

CREATE TABLE IF NOT EXISTS permissions (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    resource    TEXT NOT NULL,
    action      TEXT NOT NULL,
    description TEXT,
    UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by    TEXT,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    school_id   TEXT,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by TEXT,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (user_id)
);

-- ─────────────────────────────────────────────
-- SCHOOL
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS schools (
    id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id     TEXT NOT NULL REFERENCES states(id),
    zone_id      TEXT NOT NULL REFERENCES zones(id),
    lga_id       TEXT NOT NULL REFERENCES lgas(id),
    name         TEXT NOT NULL,
    code         TEXT NOT NULL UNIQUE,
    category     TEXT NOT NULL,
    ownership    TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'ACTIVE',
    address      TEXT,
    email        TEXT,
    phone        TEXT,
    head_teacher TEXT,
    founded      INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   TEXT,
    updated_by   TEXT
);

CREATE INDEX IF NOT EXISTS idx_schools_state_id  ON schools(state_id);
CREATE INDEX IF NOT EXISTS idx_schools_zone_id   ON schools(zone_id);
CREATE INDEX IF NOT EXISTS idx_schools_lga_id    ON schools(lga_id);
CREATE INDEX IF NOT EXISTS idx_schools_status    ON schools(status);

CREATE TABLE IF NOT EXISTS school_facilities (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    school_id   TEXT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    quantity    INTEGER NOT NULL DEFAULT 1,
    condition   TEXT NOT NULL DEFAULT 'GOOD',
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT
);

-- ─────────────────────────────────────────────
-- ACADEMIC SESSION & TERMS
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS academic_sessions (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    name        TEXT NOT NULL,
    start_year  INTEGER NOT NULL,
    end_year    INTEGER NOT NULL,
    status      TEXT NOT NULL DEFAULT 'DRAFT',
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, name)
);

CREATE TABLE IF NOT EXISTS terms (
    id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    session_id   TEXT NOT NULL REFERENCES academic_sessions(id) ON DELETE CASCADE,
    term_number  INTEGER NOT NULL CHECK (term_number IN (1,2,3)),
    name         TEXT NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    created_by   TEXT,
    updated_by   TEXT,
    UNIQUE (session_id, term_number)
);

-- ─────────────────────────────────────────────
-- LEVEL SETUP
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS levels (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    type        TEXT NOT NULL,
    ord         INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, code)
);

CREATE TABLE IF NOT EXISTS sub_levels (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    school_id   TEXT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    level_id    TEXT NOT NULL REFERENCES levels(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    capacity    INTEGER NOT NULL DEFAULT 40,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (school_id, level_id, code)
);

CREATE TABLE IF NOT EXISTS school_levels (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    school_id   TEXT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    level_id    TEXT NOT NULL REFERENCES levels(id),
    session_id  TEXT NOT NULL REFERENCES academic_sessions(id),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (school_id, level_id, session_id)
);

-- ─────────────────────────────────────────────
-- SUBJECTS
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subjects (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT 'CORE',
    level_type  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, code)
);

CREATE TABLE IF NOT EXISTS school_subjects (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    school_id   TEXT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    subject_id  TEXT NOT NULL REFERENCES subjects(id),
    level_id    TEXT NOT NULL REFERENCES levels(id),
    session_id  TEXT NOT NULL REFERENCES academic_sessions(id),
    teacher_id  TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (school_id, subject_id, level_id, session_id)
);

-- ─────────────────────────────────────────────
-- PERSONNEL
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS personnel (
    id                  TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id            TEXT NOT NULL REFERENCES states(id),
    school_id           TEXT NOT NULL REFERENCES schools(id),
    staff_id            TEXT NOT NULL UNIQUE,
    first_name          TEXT NOT NULL,
    middle_name         TEXT,
    last_name           TEXT NOT NULL,
    gender              TEXT NOT NULL,
    date_of_birth       DATE,
    email               TEXT,
    phone               TEXT,
    address             TEXT,
    role                TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'ACTIVE',
    qualification       TEXT,
    specialization      TEXT,
    date_of_employment  DATE,
    lga_id              TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    created_by          TEXT,
    updated_by          TEXT
);

CREATE INDEX IF NOT EXISTS idx_personnel_school_id ON personnel(school_id);
CREATE INDEX IF NOT EXISTS idx_personnel_state_id  ON personnel(state_id);
CREATE INDEX IF NOT EXISTS idx_personnel_staff_id  ON personnel(staff_id);

CREATE TABLE IF NOT EXISTS personnel_transfers (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    personnel_id    TEXT NOT NULL REFERENCES personnel(id),
    from_school_id  TEXT NOT NULL REFERENCES schools(id),
    to_school_id    TEXT NOT NULL REFERENCES schools(id),
    transfer_date   DATE NOT NULL,
    reason          TEXT,
    approved_by     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      TEXT,
    updated_by      TEXT
);

-- ─────────────────────────────────────────────
-- STUDENTS
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS students (
    id                TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id          TEXT NOT NULL REFERENCES states(id),
    admission_no      TEXT NOT NULL UNIQUE,
    first_name        TEXT NOT NULL,
    middle_name       TEXT,
    last_name         TEXT NOT NULL,
    gender            TEXT NOT NULL,
    date_of_birth     DATE NOT NULL,
    state_of_origin   TEXT,
    lga_id            TEXT,
    religion          TEXT,
    phone             TEXT,
    email             TEXT,
    address           TEXT,
    guardian_name     TEXT NOT NULL,
    guardian_phone    TEXT NOT NULL,
    guardian_relation TEXT,
    status            TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    created_by        TEXT,
    updated_by        TEXT
);

CREATE INDEX IF NOT EXISTS idx_students_state_id     ON students(state_id);
CREATE INDEX IF NOT EXISTS idx_students_admission_no ON students(admission_no);

CREATE TABLE IF NOT EXISTS enrollments (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    student_id    TEXT NOT NULL REFERENCES students(id),
    school_id     TEXT NOT NULL REFERENCES schools(id),
    session_id    TEXT NOT NULL REFERENCES academic_sessions(id),
    level_id      TEXT NOT NULL REFERENCES levels(id),
    sub_level_id  TEXT NOT NULL REFERENCES sub_levels(id),
    status        TEXT NOT NULL DEFAULT 'ACTIVE',
    enrolled_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    TEXT,
    updated_by    TEXT,
    UNIQUE (student_id, session_id)
);

CREATE INDEX IF NOT EXISTS idx_enrollments_school_id   ON enrollments(school_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_session_id  ON enrollments(session_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_student_id  ON enrollments(student_id);

CREATE TABLE IF NOT EXISTS level_progressions (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    student_id      TEXT NOT NULL REFERENCES students(id),
    school_id       TEXT NOT NULL REFERENCES schools(id),
    from_session_id TEXT NOT NULL REFERENCES academic_sessions(id),
    to_session_id   TEXT REFERENCES academic_sessions(id),
    from_level_id   TEXT NOT NULL REFERENCES levels(id),
    to_level_id     TEXT REFERENCES levels(id),
    decision        TEXT NOT NULL,
    decided_by      TEXT NOT NULL,
    decision_date   DATE NOT NULL,
    remarks         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      TEXT,
    updated_by      TEXT
);

-- ─────────────────────────────────────────────
-- EXAM & RESULTS
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS score_configs (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    school_id   TEXT REFERENCES schools(id),
    ca1_max     NUMERIC(5,2) NOT NULL DEFAULT 10,
    ca2_max     NUMERIC(5,2) NOT NULL DEFAULT 10,
    ca3_max     NUMERIC(5,2) NOT NULL DEFAULT 10,
    exam_max    NUMERIC(5,2) NOT NULL DEFAULT 70,
    total_max   NUMERIC(5,2) NOT NULL DEFAULT 100,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, school_id)
);

CREATE TABLE IF NOT EXISTS grade_configs (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    state_id    TEXT NOT NULL REFERENCES states(id),
    school_id   TEXT REFERENCES schools(id),
    grade       TEXT NOT NULL,
    min_score   NUMERIC(5,2) NOT NULL,
    max_score   NUMERIC(5,2) NOT NULL,
    remark      TEXT NOT NULL,
    points      NUMERIC(4,2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  TEXT,
    updated_by  TEXT,
    UNIQUE (state_id, school_id, grade)
);

CREATE TABLE IF NOT EXISTS score_sheets (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    enrollment_id TEXT NOT NULL REFERENCES enrollments(id),
    student_id    TEXT NOT NULL REFERENCES students(id),
    school_id     TEXT NOT NULL REFERENCES schools(id),
    session_id    TEXT NOT NULL REFERENCES academic_sessions(id),
    term_id       TEXT NOT NULL REFERENCES terms(id),
    subject_id    TEXT NOT NULL REFERENCES subjects(id),
    ca1_score     NUMERIC(5,2) NOT NULL DEFAULT 0,
    ca2_score     NUMERIC(5,2) NOT NULL DEFAULT 0,
    ca3_score     NUMERIC(5,2) NOT NULL DEFAULT 0,
    exam_score    NUMERIC(5,2) NOT NULL DEFAULT 0,
    total_score   NUMERIC(5,2) NOT NULL DEFAULT 0,
    grade         TEXT,
    remark        TEXT,
    position      INTEGER NOT NULL DEFAULT 0,
    recorded_by   TEXT,
    recorded_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    TEXT,
    updated_by    TEXT,
    UNIQUE (student_id, subject_id, term_id)
);

CREATE INDEX IF NOT EXISTS idx_score_sheets_student  ON score_sheets(student_id);
CREATE INDEX IF NOT EXISTS idx_score_sheets_term     ON score_sheets(term_id);
CREATE INDEX IF NOT EXISTS idx_score_sheets_school   ON score_sheets(school_id);

CREATE TABLE IF NOT EXISTS report_cards (
    id                TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    student_id        TEXT NOT NULL REFERENCES students(id),
    school_id         TEXT NOT NULL REFERENCES schools(id),
    session_id        TEXT NOT NULL REFERENCES academic_sessions(id),
    term_id           TEXT NOT NULL REFERENCES terms(id),
    level_id          TEXT NOT NULL REFERENCES levels(id),
    sub_level_id      TEXT NOT NULL REFERENCES sub_levels(id),
    total_score       NUMERIC(8,2) NOT NULL DEFAULT 0,
    average_score     NUMERIC(5,2) NOT NULL DEFAULT 0,
    overall_grade     TEXT,
    class_position    INTEGER NOT NULL DEFAULT 0,
    subject_count     INTEGER NOT NULL DEFAULT 0,
    attendance        INTEGER NOT NULL DEFAULT 0,
    total_school_days INTEGER NOT NULL DEFAULT 0,
    principal_remark  TEXT,
    teacher_remark    TEXT,
    published_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by        TEXT,
    updated_by        TEXT,
    UNIQUE (student_id, term_id)
);

-- ─────────────────────────────────────────────
-- SEED: DEFAULT PERMISSIONS
-- ─────────────────────────────────────────────
INSERT INTO permissions (id, resource, action, description) VALUES
  (gen_random_uuid()::TEXT, 'schools',    'create',  'Create a school')            ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'schools',    'read',    'View school details')        ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'schools',    'update',  'Update school details')      ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'schools',    'delete',  'Delete a school')            ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'personnel',  'create',  'Add staff member')           ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'personnel',  'read',    'View staff records')         ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'personnel',  'update',  'Update staff records')       ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'personnel',  'delete',  'Remove staff member')        ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'students',   'create',  'Register a student')         ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'students',   'read',    'View student records')       ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'students',   'update',  'Update student records')     ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'students',   'delete',  'Delete student record')      ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'enrollments','create',  'Enrol a student')            ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'enrollments','read',    'View enrolments')            ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'enrollments','update',  'Update enrolment')           ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'results',    'create',  'Enter exam scores')          ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'results',    'read',    'View results')               ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'results',    'update',  'Modify scores')              ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'results',    'publish', 'Publish report cards')       ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'sessions',   'create',  'Create academic session')    ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'sessions',   'update',  'Update academic session')    ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'users',      'create',  'Create user accounts')       ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'users',      'read',    'View user accounts')         ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'users',      'update',  'Update user accounts')       ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'users',      'delete',  'Delete user accounts')       ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'roles',      'create',  'Create roles')               ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'roles',      'read',    'View roles')                 ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'roles',      'update',  'Update roles')               ON CONFLICT (resource, action) DO NOTHING,
  (gen_random_uuid()::TEXT, 'roles',      'delete',  'Delete roles')               ON CONFLICT (resource, action) DO NOTHING;
`
