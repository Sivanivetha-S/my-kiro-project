# StudentHub — API and Database Conventions

**Source of truth:** `.kiro/specs/studenthub/design.md`
**Load this file when:** writing HTTP handlers, SQL queries, building request/response types, or working with the SQLite database.

---

## REST API Conventions

### Base path and versioning

All API endpoints are prefixed `/api/v1`. The frontend calls these via `api.js`. No endpoint exists outside this prefix except the static file routes (`/`, `/css/...`, `/js/...`).

### HTTP method semantics

| Operation | Method | Success status |
|---|---|---|
| Create a resource | POST | **201 Created** |
| Read a list | GET | **200 OK** |
| Read a single resource | GET | **200 OK** |
| Full update | PUT | **200 OK** |
| Delete | DELETE | **204 No Content** (empty body) |

Never return 200 for a create. Never return a body for a delete. Never use PATCH — all updates are full PUT in this project.

### JSON field naming

All JSON fields use `snake_case`. This must match:
- The `json` struct tag on every Go struct field
- The property names used in JavaScript fetch calls

```go
// Correct
type Student struct {
    StudentID  string `json:"student_id"`
    FullName   string `json:"full_name"`
    DOB        string `json:"dob"`
    CreatedAt  string `json:"created_at"`
}
```

### Content-Type

Every response (including errors) sets `Content-Type: application/json`. The `writeJSON` and `writeError` helpers in `handlers/response.go` do this automatically — use them everywhere.

### Timestamps and dates

- Timestamps: ISO 8601 UTC string — `"2026-10-02T14:30:00Z"`
- Dates (DOB): `"YYYY-MM-DD"` string — `"2004-06-15"`
- Both are stored as `TEXT` in SQLite and returned as strings in JSON. No time.Time in API responses.

---

## Complete Routing Table

Source: `design.md § 5.3`. Do not add routes not listed here without updating the spec first.

```
POST   /api/v1/students                    → students.Create
GET    /api/v1/students                    → students.List         (?search=, ?department=, ?year=)
GET    /api/v1/students/{id}               → students.GetByID      (full profile with courses/marks/attendance)
PUT    /api/v1/students/{id}               → students.Update
DELETE /api/v1/students/{id}               → students.Delete

GET    /api/v1/students/{id}/courses       → enrollments.GetCoursesForStudent
GET    /api/v1/students/{id}/marks         → marks.GetForStudent
GET    /api/v1/students/{id}/attendance    → attendance.GetForStudent

POST   /api/v1/courses                     → courses.Create
GET    /api/v1/courses                     → courses.List
GET    /api/v1/courses/{id}                → courses.GetByID
PUT    /api/v1/courses/{id}                → courses.Update
DELETE /api/v1/courses/{id}               → courses.Delete
GET    /api/v1/courses/{id}/students       → enrollments.GetStudentsForCourse

POST   /api/v1/enrollments                 → enrollments.Create
DELETE /api/v1/enrollments/{id}            → enrollments.Delete

POST   /api/v1/marks                       → marks.Create
PUT    /api/v1/marks/{id}                  → marks.Update

POST   /api/v1/attendance                  → attendance.Create
GET    /api/v1/attendance/low              → attendance.GetLowAttendance
PUT    /api/v1/attendance/{id}             → attendance.Update

GET    /api/v1/dashboard                   → dashboard.Get
```

---

## Error Response Format

All errors — validation failures, not-found, conflicts, server errors — return JSON in this exact shape:

```json
{ "error": "<human-readable message>", "code": "<SCREAMING_SNAKE_CASE>" }
```

### HTTP status code mapping

| Situation | Status | Code field example |
|---|---|---|
| Input validation failure | 400 | `"VALIDATION_ERROR"` |
| Resource not found | 404 | `"NOT_FOUND"` |
| Duplicate unique field | 409 | `"DUPLICATE_STUDENT_ID"`, `"DUPLICATE_EMAIL"`, `"DUPLICATE_COURSE_CODE"` |
| Unexpected server error | 500 | `"INTERNAL_ERROR"` |

For 500 responses: log the full error with `log.Printf` server-side. The client response body must contain only the generic message `"An unexpected error occurred"`. **Never send Go error strings, stack traces, or SQL error text to the client.**

### Handler error mapping pattern

Every handler must use this exact switch for model errors:

```go
switch {
case errors.Is(err, models.ErrNotFound):
    writeError(w, http.StatusNotFound, "Resource not found", "NOT_FOUND")
case errors.Is(err, models.ErrDuplicate):
    writeError(w, http.StatusConflict, "Record already exists", "DUPLICATE")
default:
    log.Printf("handler error: %v", err)
    writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
}
```

`models.ErrNotFound` and `models.ErrDuplicate` are sentinel errors defined in `models/errors.go`.

---

## SQLite Database Rules

Source: `design.md § 7`, `architecture.md`

### Schema — five tables

```
students      — student records, student_id and email are UNIQUE
courses       — course records, course_code is UNIQUE
enrollments   — many-to-many join, UNIQUE(student_id, course_id)
marks         — one per (student_id, course_id), UNIQUE(student_id, course_id)
attendance    — one per (student_id, course_id), UNIQUE(student_id, course_id)
```

All foreign keys use `ON DELETE CASCADE`. Deleting a student removes their enrollments, marks, and attendance automatically. Deleting a course does the same.

### Connection setup — two mandatory pragmas

Both of these must be applied in `db.Open()` immediately after opening the connection:

```go
// 1. Foreign keys are OFF by default in SQLite. Must be set on every connection.
db.Exec("PRAGMA foreign_keys = ON")

// 2. SQLite allows only one concurrent writer. One connection prevents locking errors.
db.SetMaxOpenConns(1)
```

Omitting either of these causes silent data integrity failures or runtime panics under concurrent load.

### Parameterised statements — no exceptions

All SQL that incorporates user input uses `?` placeholders:

```go
// Always correct
db.QueryRow("SELECT * FROM students WHERE id = ?", id)
db.QueryRow("SELECT * FROM students WHERE student_id = ?", studentID)

// Never acceptable — SQL injection risk
db.QueryRow("SELECT * FROM students WHERE id = " + id)
db.QueryRow(fmt.Sprintf("SELECT * FROM students WHERE student_id = '%s'", studentID))
```

This rule has no exceptions, including for integer parameters.

### Migrations

Schema lives in `db/migrations/001_init.sql`. Applied by `db.RunMigrations()` on startup. A `schema_migrations` table tracks which migrations have run, making migrations idempotent. New schema changes go in a new numbered file (e.g., `002_add_index.sql`), never by editing the existing file.

### Indexes

The following indexes are defined in `001_init.sql` and must be preserved:

```sql
CREATE INDEX idx_students_department ON students(department);
CREATE INDEX idx_students_year       ON students(year);
CREATE INDEX idx_students_student_id ON students(student_id);
CREATE INDEX idx_enrollments_student ON enrollments(student_id);
CREATE INDEX idx_enrollments_course  ON enrollments(course_id);
CREATE INDEX idx_marks_student       ON marks(student_id);
CREATE INDEX idx_attendance_student  ON attendance(student_id);
```

---

## Frontend API Module

All `fetch()` calls in the frontend go through `frontend/js/api.js` only. No other file calls `fetch()` directly.

```javascript
// Every page module imports api and calls its methods
import { api } from '../api.js';
const students = await api.get('/students');
await api.post('/students', formData);
await api.put(`/students/${id}`, formData);
await api.delete(`/students/${id}`);
```

HTTP errors from the server (non-2xx) are thrown as structured objects `{ status, error, code }` by `api.js`. Page modules catch these and either call `form.displayErrors()` for 400/409 responses or `toast.show('error', ...)` for 500/network failures. Raw error strings are never shown to the user.
