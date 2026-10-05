-- StudentHub initial schema
-- Migration: 001_init.sql
-- Creates all domain tables, the schema_migrations tracking table, and all indexes.
-- This file must be idempotent when applied through RunMigrations, which checks
-- schema_migrations before executing.

-- ============================================================
-- Migration tracking
-- ============================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    filename   TEXT NOT NULL PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- ============================================================
-- Domain tables  (design.md § 3.2 – 3.6)
-- ============================================================

CREATE TABLE IF NOT EXISTS students (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id  TEXT    NOT NULL UNIQUE,
    full_name   TEXT    NOT NULL,
    email       TEXT    NOT NULL UNIQUE,
    phone       TEXT,
    department  TEXT    NOT NULL,
    year        INTEGER NOT NULL CHECK(year BETWEEN 1 AND 6),
    section     TEXT    NOT NULL,
    dob         TEXT    NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS courses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    course_code TEXT    NOT NULL UNIQUE,
    course_name TEXT    NOT NULL,
    credits     INTEGER NOT NULL CHECK(credits BETWEEN 1 AND 6),
    department  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS enrollments (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id   INTEGER NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    enrolled_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(student_id, course_id)
);

CREATE TABLE IF NOT EXISTS marks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id   INTEGER NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    marks       REAL    NOT NULL CHECK(marks >= 0 AND marks <= 100),
    recorded_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(student_id, course_id)
);

CREATE TABLE IF NOT EXISTS attendance (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id       INTEGER NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    total_classes   INTEGER NOT NULL CHECK(total_classes >= 1),
    attended        INTEGER NOT NULL CHECK(attended >= 0),
    recorded_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(student_id, course_id),
    CHECK(attended <= total_classes)
);

-- ============================================================
-- Indexes  (design.md § 7.1)
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_students_department ON students(department);
CREATE INDEX IF NOT EXISTS idx_students_year       ON students(year);
CREATE INDEX IF NOT EXISTS idx_students_student_id ON students(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course  ON enrollments(course_id);
CREATE INDEX IF NOT EXISTS idx_marks_student       ON marks(student_id);
CREATE INDEX IF NOT EXISTS idx_attendance_student  ON attendance(student_id);
