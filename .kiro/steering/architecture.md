---
inclusion: always
---

# StudentHub — Architecture

This document defines the architectural rules for StudentHub. Every file added to the project must fit within the structure described here. When in doubt about where code belongs, this document is the authority.

---

## Three-Tier Architecture

StudentHub is organized as three distinct tiers. Each tier has a single, well-defined responsibility. Code must not reach across tiers — every interaction crosses the boundary through the defined interface only.

```
┌─────────────────────────────────────────┐
│            Tier 1 — Frontend            │
│   HTML + CSS + Vanilla JavaScript       │
│   Runs in the browser                   │
│   Communicates via REST API only        │
└────────────────────┬────────────────────┘
                     │ HTTP JSON  (REST API at /api/v1)
┌────────────────────▼────────────────────┐
│            Tier 2 — Go Server           │
│   Handles HTTP, validates input,        │
│   applies business rules,               │
│   reads and writes data                 │
└────────────────────┬────────────────────┘
                     │ database/sql  (parameterized SQL)
┌────────────────────▼────────────────────┐
│            Tier 3 — SQLite              │
│   Single-file relational database       │
│   studenthub.db                         │
└─────────────────────────────────────────┘
```

The Go server also serves the static frontend files (`/`, `/css/...`, `/js/...`). This is a deployment convenience — it does not blur the tier boundary. The frontend still communicates through the API, not by calling Go functions directly.

---

## Backend Package Organization

The backend is divided into six packages. Each package has a strict responsibility. These boundaries are enforced by convention — follow them exactly when adding new files.

### `main`

**File:** `main.go`  
**Responsibility:** Entry point only. Wire everything together and start the server.

`main.go` does exactly four things:

1. Load configuration via `config.Load()`.
2. Open the database via `db.Open()` and run migrations via `db.RunMigrations()`.
3. Register all HTTP routes using `net/http.ServeMux`, applying middleware.
4. Call `http.ListenAndServe`.

`main.go` must not contain any business logic, validation, SQL, or HTML. If it grows beyond ~60 lines, something is in the wrong place.

---

### `config`

**File:** `config/config.go`  
**Responsibility:** Read runtime configuration from environment variables and expose it as a typed struct.

```
config.Config{
    Port                string   // from PORT, default "8080"
    DBPath              string   // from DB_PATH, default "./studenthub.db"
    AttendanceThreshold float64  // from ATTENDANCE_THRESHOLD, default 75.0
}
```

`config` has no dependencies on any other package in this project. It only calls `os.Getenv` and `strconv` from the standard library.

---

### `db`

**Files:** `db/db.go`, `db/migrations/*.sql`  
**Responsibility:** Database connection management and schema migration only.

`db.Open(path string) (*sql.DB, error)` must:
- Open the SQLite connection.
- Set `PRAGMA foreign_keys = ON` immediately after every connection open. This pragma is not persisted in the file — it must be set on every new `*sql.DB`.
- Set `db.SetMaxOpenConns(1)`. SQLite supports only one concurrent writer. This prevents "database is locked" errors.

`db.RunMigrations(db *sql.DB, dir string) error` reads numbered `.sql` files from `db/migrations/` in filename order and applies any not yet recorded in the `schema_migrations` table.

The `db` package must not define any domain structs, handle HTTP, or contain validation logic.

---

### `models`

**Files:** `models/student.go`, `models/course.go`, `models/enrollment.go`, `models/marks.go`, `models/attendance.go`, `models/errors.go`  
**Responsibility:** Domain structs and all SQL query functions.

Rules for the `models` package:

- Every function that reads or writes the database lives here. No SQL anywhere else.
- Functions accept a `*sql.DB` as their first argument. They do not hold a reference to a global DB.
- All queries use `?` placeholders. String-interpolated SQL is never acceptable.
- Sentinel errors `ErrNotFound` and `ErrDuplicate` are defined in `models/errors.go` and returned by model functions when appropriate. Handlers translate these to HTTP status codes.
- Computed values — grade and attendance percentage — are calculated inside model functions on read, never stored in the database. See the "Computed Values" section below.
- The `models` package does not import `handlers`, `validators`, or `middleware`. It only imports `database/sql`, `math`, and the standard library.

---

### `validators`

**Files:** `validators/student_validator.go`, `validators/course_validator.go`, `validators/marks_validator.go`, `validators/attendance_validator.go`  
**Responsibility:** Input validation only. Pure functions — no I/O, no database, no HTTP.

Every validator function has the signature:

```go
func ValidateXxx(x models.Xxx) []ValidationError
```

`ValidationError` is defined once:

```go
type ValidationError struct {
    Field   string
    Message string
}
```

Rules for the `validators` package:

- Validators return all failing fields at once, not just the first. This allows the UI to show all inline errors in one response.
- Validators never query the database. Uniqueness checks (duplicate Student ID, duplicate email) are the responsibility of the model layer, which returns `ErrDuplicate`.
- Validators are pure functions: given the same input, they always return the same output. They have no side effects.
- The `validators` package imports only `models` and the standard library (`regexp`, `strings`, `time`, `unicode`).

---

### `handlers`

**Files:** `handlers/students.go`, `handlers/courses.go`, `handlers/enrollments.go`, `handlers/marks.go`, `handlers/attendance.go`, `handlers/dashboard.go`, `handlers/response.go`  
**Responsibility:** HTTP request/response lifecycle only. Decode → validate → call model → encode.

Every handler function follows this exact sequence:

```
1. Decode the request body (json.Decoder) or read path/query parameters.
2. Call the appropriate validator. If errors → return 400.
3. Call the appropriate model function.
4. Map model errors to HTTP status codes (see "Error Flow" below).
5. Encode the response (json.Encoder) and write the status code.
```

Rules for the `handlers` package:

- No SQL in handlers. All data access goes through `models`.
- No business logic in handlers. Grade calculation, attendance percentage, threshold comparisons — all of these live in `models`.
- No validation logic in handlers. All validation goes through `validators`.
- Handlers receive the `*sql.DB` via closure or struct injection from `main.go`. They do not open or close database connections.
- `handlers/response.go` contains the shared `writeJSON` and `writeError` helpers used by all handlers. Do not duplicate response-writing logic.

---

### `middleware`

**File:** `middleware/cors.go`  
**Responsibility:** HTTP middleware only.

Middleware functions accept and return `http.Handler`. They must not contain business logic, validation, or database access.

---

## Frontend Module Organization

The frontend is organized as ES Modules. The responsibilities mirror the backend's separation of concerns.

```
frontend/js/
├── app.js              — hash router + page loader. No business logic.
├── api.js              — all fetch() calls. The only file that knows the API base URL.
├── utils.js            — pure helper functions (date formatting, grade badge class, etc.)
├── pages/              — one file per page. Orchestrates api.js + components.
│   ├── dashboard.js
│   ├── students.js
│   ├── student-detail.js
│   ├── courses.js
│   ├── marks.js
│   └── attendance.js
└── components/         — reusable UI pieces. Receive data, render HTML.
    ├── modal.js
    ├── toast.js
    ├── table.js
    └── form.js
```

Rules for the frontend:

- **`api.js` is the only file that calls `fetch()`**. Page modules import `api.js` and call its methods. No other file reaches the network directly.
- **Page modules orchestrate; components render.** A page module calls the API, processes the response, then hands data to components for rendering.
- **`utils.js` contains only pure functions.** No DOM manipulation, no API calls, no side effects.
- The hash-based router in `app.js` maps URL patterns to page modules. No routing logic appears in page modules themselves.
- CSS classes come from the stylesheets. No inline styles in JavaScript (`element.style.xxx = ...` is not acceptable except for dynamic values that cannot be expressed as CSS classes, such as a computed width percentage).

---

## API Boundaries

The API is the contract between the frontend and backend. These rules maintain its integrity.

**URL structure:**
- All API endpoints are prefixed `/api/v1`.
- Resource paths follow the pattern: `/api/v1/{resource}` (collection) and `/api/v1/{resource}/{id}` (single item).
- Sub-resource paths: `/api/v1/students/{id}/courses`, `/api/v1/students/{id}/marks`, `/api/v1/students/{id}/attendance`.
- The full routing table is defined in `design.md § 5.3`. Do not add routes that are not in the spec without updating the spec first.

**HTTP methods:** Use HTTP methods according to their semantics.

| Operation | Method | Success status |
|---|---|---|
| Create | POST | 201 Created |
| Read (list or single) | GET | 200 OK |
| Full update | PUT | 200 OK |
| Delete | DELETE | 204 No Content |

**JSON field names:** All JSON fields use `snake_case` (e.g., `student_id`, `full_name`, `course_code`). This must match the `json` struct tags in Go and the property names used in frontend JavaScript.

**Content type:** Every API response sets `Content-Type: application/json`. This applies to error responses too.

---

## Database Interaction Boundaries

These rules prevent SQL from escaping the `models` package and ensure data integrity.

**Parameterized statements only.** Every SQL statement that incorporates user input uses `?` placeholders. No string concatenation or `fmt.Sprintf` to build SQL. This applies without exception.

**Foreign key enforcement.** `PRAGMA foreign_keys = ON` must be executed on every new database connection in `db.Open()`. Without it, SQLite silently ignores foreign key constraints.

**Single writer.** `db.SetMaxOpenConns(1)` must be set. SQLite allows multiple concurrent readers but only one writer. Using a single connection avoids write contention and "database is locked" errors.

**Cascade deletes.** The `enrollments`, `marks`, and `attendance` tables use `ON DELETE CASCADE` referencing both `students` and `courses`. When a student or course is deleted, their related rows are removed automatically by the database. Do not replicate this logic in application code.

**No stored computed values.** The following values must never be stored in the database:

| Value | Stored as | Computed from |
|---|---|---|
| Letter grade | never stored | `marks` column in `marks` table, via `CalculateGrade()` |
| Attendance percentage | never stored | `attended` / `total_classes` in `attendance` table, via `CalculateAttendancePercentage()` |

Computed values are derived on every read. This ensures they are always consistent with the raw data and eliminates the possibility of stale cached values.

---

## Validation and Business Logic Placement

This table defines the single correct location for each type of logic. Do not place logic in a layer other than the one listed.

| Logic type | Lives in | Example |
|---|---|---|
| Field format validation | `validators` package | Email format, student ID character rules |
| Range validation | `validators` package | Year 1–6, marks 0–100, credits 1–6 |
| Business rule validation | `validators` package | Age ≥ 15, attended ≤ total_classes |
| Uniqueness checks | `models` package | Duplicate student_id, duplicate course_code |
| Referential integrity | `models` package + SQLite | Enrollment requires existing student + course |
| Grade calculation | `models/marks.go` | `CalculateGrade(float64) string` |
| Attendance % calculation | `models/attendance.go` | `CalculateAttendancePercentage(int, int) float64` |
| Low-attendance threshold | `models/attendance.go` | Comparison against `config.AttendanceThreshold` |
| HTTP status codes | `handlers` package | `ErrNotFound` → 404, `ErrDuplicate` → 409 |
| Response serialization | `handlers` package | `writeJSON`, `writeError` |
| DOM rendering | `frontend/js/pages/` | Rendering API responses into HTML |
| Client-side validation | `frontend/js/components/form.js` | Mirror of server-side rules, runs before fetch |

The key principle: **validation happens in `validators`, data access happens in `models`, HTTP happens in `handlers`. No layer does another layer's job.**

---

## Error Flow: Backend to Frontend

Errors travel through a defined path. Each layer transforms errors into the format appropriate for the next layer.

```
SQLite constraint error
        │
        ▼
models layer
  - sql.ErrNoRows         → return models.ErrNotFound
  - UNIQUE constraint err → return models.ErrDuplicate
  - other DB error        → return wrapped error (log server-side)
        │
        ▼
handlers layer
  - models.ErrNotFound    → HTTP 404 + {"error": "...", "code": "NOT_FOUND"}
  - models.ErrDuplicate   → HTTP 409 + {"error": "...", "code": "DUPLICATE_..."}
  - []ValidationError     → HTTP 400 + {"error": "...", "code": "VALIDATION_ERROR", "fields": [...]}
  - any other error       → HTTP 500 + {"error": "An unexpected error occurred", "code": "INTERNAL_ERROR"}
                            (full error is logged to server stdout only — never sent to client)
        │
        ▼ HTTP JSON response
frontend api.js
  - non-2xx response      → throw { status, error, code }
        │
        ▼
frontend page module
  - status 400            → pass field errors to form.displayErrors()
  - status 409            → display conflict message near the relevant field
  - status 404            → display "not found" state in the UI
  - status 500 / network  → call toast.show('error', 'Something went wrong. Please try again.')
```

**The critical rule:** Internal error details (Go error strings, stack traces, SQL errors, file paths) must never reach the HTTP response body. The handler logs the full error server-side and returns a generic message to the client.

---

## What Goes Where — Quick Reference

| I need to... | File to edit |
|---|---|
| Add a new API endpoint | `main.go` (route), `handlers/xxx.go` (handler), `models/xxx.go` (query) |
| Add a new validation rule | `validators/xxx_validator.go` |
| Change the grading scale | `models/marks.go` — `CalculateGrade()` only |
| Change the attendance threshold | `config/config.go` default + `ATTENDANCE_THRESHOLD` env var |
| Add a new database table | `db/migrations/002_xxx.sql` (new migration file) |
| Add a new frontend page | `frontend/js/pages/xxx.js` + route entry in `frontend/js/app.js` |
| Add a reusable UI component | `frontend/js/components/xxx.js` |
| Change how errors are displayed | `frontend/js/components/form.js` (field errors), `frontend/js/components/toast.js` (global errors) |
| Change a CSS design token | `frontend/css/main.css` (`:root` custom properties) |
