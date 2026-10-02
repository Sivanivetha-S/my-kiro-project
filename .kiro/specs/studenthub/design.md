# StudentHub — Design Specification

**Project:** StudentHub — Student Management System  
**Version:** 1.0  
**Context:** Kiro University — Spec-Driven Development Demonstration  
**Date:** 2026-10-02  
**Status:** Draft

---

## 1. Architecture Overview

StudentHub follows a classic three-tier architecture:

```
┌─────────────────────────────────┐
│         Browser (Client)        │
│   HTML + CSS + Vanilla JS       │
│   Static files served by Go     │
└────────────────┬────────────────┘
                 │ HTTP / JSON REST API
┌────────────────▼────────────────┐
│         Go HTTP Server          │
│   net/http  ·  handlers  ·      │
│   validators  ·  models         │
└────────────────┬────────────────┘
                 │ database/sql
┌────────────────▼────────────────┐
│         SQLite Database         │
│         studenthub.db           │
└─────────────────────────────────┘
```

- The Go server serves both the REST API (`/api/...`) and the static frontend files (`/`).
- The frontend communicates exclusively through the REST API using `fetch()`.
- No server-side HTML rendering — all UI state is managed in JavaScript.
- CORS middleware is applied to all `/api/` routes for local development flexibility.

---

## 2. Project Directory Structure

```
studenthub/
├── main.go                    # Entry point: wires router, DB, starts server
├── go.mod
├── go.sum
├── studenthub.db              # SQLite database file (gitignored)
├── config/
│   └── config.go              # App configuration (port, DB path, threshold)
├── db/
│   └── db.go                  # DB connection, migration runner
│   └── migrations/
│       └── 001_init.sql       # Schema creation SQL
│       └── seed.sql           # Development seed data (gitignored from prod)
├── models/
│   ├── student.go             # Student struct + DB queries
│   ├── course.go              # Course struct + DB queries
│   ├── enrollment.go          # Enrollment struct + DB queries
│   ├── marks.go               # Marks struct + DB queries + grade calc
│   └── attendance.go          # Attendance struct + DB queries + % calc
├── handlers/
│   ├── students.go            # HTTP handlers for /api/students
│   ├── courses.go             # HTTP handlers for /api/courses
│   ├── enrollments.go         # HTTP handlers for /api/enrollments
│   ├── marks.go               # HTTP handlers for /api/marks
│   ├── attendance.go          # HTTP handlers for /api/attendance
│   └── dashboard.go           # HTTP handler for /api/dashboard
├── validators/
│   ├── student_validator.go
│   ├── course_validator.go
│   ├── marks_validator.go
│   └── attendance_validator.go
├── middleware/
│   └── cors.go                # CORS headers middleware
├── frontend/
│   ├── index.html             # Single-page shell
│   ├── css/
│   │   ├── main.css           # Global styles, CSS variables, layout
│   │   ├── components.css     # Reusable: buttons, cards, tables, forms
│   │   └── responsive.css     # Media queries
│   └── js/
│       ├── app.js             # Router, global state, page loader
│       ├── api.js             # Centralized fetch wrapper
│       ├── utils.js           # Shared helpers (format date, grade badge, etc.)
│       ├── pages/
│       │   ├── dashboard.js
│       │   ├── students.js
│       │   ├── student-detail.js
│       │   ├── courses.js
│       │   ├── marks.js
│       │   └── attendance.js
│       └── components/
│           ├── modal.js       # Reusable modal dialog
│           ├── toast.js       # Notification toasts
│           ├── table.js       # Reusable sortable table renderer
│           └── form.js        # Form validation helpers
└── tests/
    ├── models/
    │   ├── marks_test.go
    │   └── attendance_test.go
    ├── validators/
    │   ├── student_validator_test.go
    │   ├── course_validator_test.go
    │   └── marks_validator_test.go
    └── handlers/
        ├── students_handler_test.go
        └── courses_handler_test.go
```

---

## 3. Data Model

### 3.1 Entity-Relationship Diagram (text)

```
students ──< enrollments >── courses
    │                             │
    └──< marks                    │
    │       └── (student + course)│
    └──< attendance               │
            └── (student + course)│
```

- `students` 1:M `enrollments`
- `courses` 1:M `enrollments`
- `students` 1:M `marks` (one per enrolled course)
- `courses` 1:M `marks`
- `students` 1:M `attendance` (one per enrolled course)
- `courses` 1:M `attendance`

### 3.2 Table: `students`

```sql
CREATE TABLE students (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id  TEXT    NOT NULL UNIQUE,        -- user-defined, e.g. STU2024001
    full_name   TEXT    NOT NULL,
    email       TEXT    NOT NULL UNIQUE,
    phone       TEXT,
    department  TEXT    NOT NULL,
    year        INTEGER NOT NULL CHECK(year BETWEEN 1 AND 6),
    section     TEXT    NOT NULL,
    dob         TEXT    NOT NULL,               -- stored as YYYY-MM-DD
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
```

### 3.3 Table: `courses`

```sql
CREATE TABLE courses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    course_code TEXT    NOT NULL UNIQUE,        -- e.g. CS101
    course_name TEXT    NOT NULL,
    credits     INTEGER NOT NULL CHECK(credits BETWEEN 1 AND 6),
    department  TEXT    NOT NULL
);
```

### 3.4 Table: `enrollments`

```sql
CREATE TABLE enrollments (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id   INTEGER NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    enrolled_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(student_id, course_id)
);
```

### 3.5 Table: `marks`

```sql
CREATE TABLE marks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id  INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id   INTEGER NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    marks       REAL    NOT NULL CHECK(marks >= 0 AND marks <= 100),
    recorded_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(student_id, course_id)
);
```

> Grade is **not stored**. It is computed from `marks` on every read.

### 3.6 Table: `attendance`

```sql
CREATE TABLE attendance (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id       INTEGER NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
    total_classes   INTEGER NOT NULL CHECK(total_classes >= 1),
    attended        INTEGER NOT NULL CHECK(attended >= 0),
    recorded_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(student_id, course_id),
    CHECK(attended <= total_classes)
);
```

> Attendance percentage is **not stored**. It is computed from `attended / total_classes` on every read.

---

## 4. API Design

### 4.1 Conventions

- Base path: `/api/v1`
- All request and response bodies are `application/json`
- Timestamps are ISO 8601 UTC strings: `"2026-10-02T14:30:00Z"`
- Dates are `"YYYY-MM-DD"` strings
- All list responses return a JSON array (not wrapped in an object) unless metadata is needed
- Error responses always follow: `{"error": "<human message>", "code": "<SCREAMING_SNAKE_CASE>"}`

### 4.2 Students Endpoints

#### `POST /api/v1/students`
Create a new student.

**Request body:**
```json
{
  "student_id": "STU2024001",
  "full_name": "Aisha Rajan",
  "email": "aisha@example.com",
  "phone": "+91-9876543210",
  "department": "Computer Science",
  "year": 2,
  "section": "A",
  "dob": "2004-06-15"
}
```

**Response 201:**
```json
{
  "id": 1,
  "student_id": "STU2024001",
  "full_name": "Aisha Rajan",
  "email": "aisha@example.com",
  "phone": "+91-9876543210",
  "department": "Computer Science",
  "year": 2,
  "section": "A",
  "dob": "2004-06-15",
  "created_at": "2026-10-02T14:30:00Z"
}
```

**Error responses:** 400 (validation), 409 (duplicate student_id or email)

---

#### `GET /api/v1/students`
Get all students. Supports query parameters:

| Parameter | Type | Description |
|---|---|---|
| `search` | string | Partial match on full_name, email, or exact student_id |
| `department` | string | Filter by department |
| `year` | integer | Filter by year |

**Response 200:** Array of student objects (same shape as POST response).

---

#### `GET /api/v1/students/:id`
Get a single student's full profile.

**Response 200:**
```json
{
  "id": 1,
  "student_id": "STU2024001",
  "full_name": "Aisha Rajan",
  "email": "aisha@example.com",
  "phone": "+91-9876543210",
  "department": "Computer Science",
  "year": 2,
  "section": "A",
  "dob": "2004-06-15",
  "created_at": "2026-10-02T14:30:00Z",
  "courses": [
    {
      "enrollment_id": 3,
      "course_id": 1,
      "course_code": "CS101",
      "course_name": "Introduction to Programming",
      "credits": 4,
      "marks": 82.5,
      "grade": "A",
      "attendance_percentage": 87.50
    }
  ],
  "average_marks": 82.5,
  "overall_attendance": 87.50
}
```

**Error responses:** 404 (student not found)

---

#### `PUT /api/v1/students/:id`
Update a student. Accepts same body shape as POST.

**Response 200:** Updated student object.  
**Error responses:** 400, 404, 409

---

#### `DELETE /api/v1/students/:id`
Delete a student and cascade-delete their marks and attendance.

**Response 204:** No body.  
**Error responses:** 404

---

### 4.3 Courses Endpoints

#### `POST /api/v1/courses`
**Request body:**
```json
{
  "course_code": "CS101",
  "course_name": "Introduction to Programming",
  "credits": 4,
  "department": "Computer Science"
}
```
**Response 201:** Course object with `id`.  
**Error responses:** 400, 409 (duplicate course_code)

---

#### `GET /api/v1/courses`
**Response 200:** Array of course objects.

---

#### `GET /api/v1/courses/:id`
**Response 200:** Single course with enrolled student count.  
**Error responses:** 404

---

#### `PUT /api/v1/courses/:id`
**Response 200:** Updated course object.  
**Error responses:** 400, 404, 409

---

#### `DELETE /api/v1/courses/:id`
**Response 204:** No body.  
**Error responses:** 404

---

### 4.4 Enrollment Endpoints

#### `POST /api/v1/enrollments`
Assign a course to a student.

**Request body:**
```json
{
  "student_id": 1,
  "course_id": 1
}
```
**Response 201:** Enrollment object with `id`, `student_id`, `course_id`, `enrolled_at`.  
**Error responses:** 400, 404 (student or course not found), 409 (already enrolled)

---

#### `GET /api/v1/students/:id/courses`
Get all courses a student is enrolled in.  
**Response 200:** Array of course objects.

---

#### `GET /api/v1/courses/:id/students`
Get all students enrolled in a course.  
**Response 200:** Array of student objects.

---

#### `DELETE /api/v1/enrollments/:id`
Remove an enrollment.  
**Response 204:** No body.  
**Error responses:** 404

---

### 4.5 Marks Endpoints

#### `POST /api/v1/marks`
**Request body:**
```json
{
  "student_id": 1,
  "course_id": 1,
  "marks": 82.5
}
```
**Response 201:**
```json
{
  "id": 1,
  "student_id": 1,
  "course_id": 1,
  "marks": 82.5,
  "grade": "A",
  "recorded_at": "2026-10-02T14:30:00Z"
}
```
**Error responses:** 400 (validation, student not enrolled), 409 (marks already recorded)

---

#### `GET /api/v1/students/:id/marks`
Get all marks for a student.

**Response 200:**
```json
{
  "student_id": 1,
  "student_name": "Aisha Rajan",
  "marks": [
    {
      "marks_id": 1,
      "course_id": 1,
      "course_code": "CS101",
      "course_name": "Introduction to Programming",
      "marks": 82.5,
      "grade": "A"
    }
  ],
  "average_marks": 82.5
}
```

---

#### `PUT /api/v1/marks/:id`
Update a marks record.  
**Request body:** `{ "marks": 88.0 }`  
**Response 200:** Updated marks object with recomputed grade.  
**Error responses:** 400, 404

---

### 4.6 Attendance Endpoints

#### `POST /api/v1/attendance`
**Request body:**
```json
{
  "student_id": 1,
  "course_id": 1,
  "total_classes": 48,
  "attended": 36
}
```
**Response 201:**
```json
{
  "id": 1,
  "student_id": 1,
  "course_id": 1,
  "total_classes": 48,
  "attended": 36,
  "percentage": 75.00,
  "recorded_at": "2026-10-02T14:30:00Z"
}
```
**Error responses:** 400 (validation, student not enrolled), 409 (record already exists)

---

#### `GET /api/v1/students/:id/attendance`
Get all attendance records for a student.

**Response 200:**
```json
{
  "student_id": 1,
  "student_name": "Aisha Rajan",
  "attendance": [
    {
      "attendance_id": 1,
      "course_id": 1,
      "course_code": "CS101",
      "course_name": "Introduction to Programming",
      "total_classes": 48,
      "attended": 36,
      "percentage": 75.00
    }
  ],
  "overall_percentage": 75.00
}
```

---

#### `GET /api/v1/attendance/low`
Get all students with attendance below the configured threshold (75%).

**Response 200:** Array of objects:
```json
[
  {
    "student_id": "STU2024001",
    "full_name": "Aisha Rajan",
    "department": "Computer Science",
    "year": 2,
    "overall_attendance": 68.50
  }
]
```

---

#### `PUT /api/v1/attendance/:id`
Update an attendance record.  
**Request body:** `{ "total_classes": 50, "attended": 38 }`  
**Response 200:** Updated attendance object with recomputed percentage.  
**Error responses:** 400, 404

---

### 4.7 Dashboard Endpoint

#### `GET /api/v1/dashboard`
**Response 200:**
```json
{
  "total_students": 120,
  "total_courses": 18,
  "average_attendance": 78.42,
  "low_attendance_students": [
    {
      "student_id": "STU2024007",
      "full_name": "Rohan Mehta",
      "department": "Mechanical Engineering",
      "overall_attendance": 61.20
    }
  ],
  "recent_students": [
    {
      "id": 120,
      "student_id": "STU2024120",
      "full_name": "Priya Singh",
      "department": "Computer Science",
      "year": 1,
      "created_at": "2026-10-02T14:30:00Z"
    }
  ]
}
```

---

## 5. Backend Design

### 5.1 Go Package Responsibilities

| Package | Responsibility |
|---|---|
| `main` | Wire everything: create DB connection, register routes, apply middleware, start server. |
| `config` | Read configuration from environment variables with defaults. Exports `Config` struct. |
| `db` | Open SQLite connection, run SQL migrations on startup. Exports `*sql.DB`. |
| `models` | Define Go structs matching DB tables. Implement all SQL queries as methods. No HTTP logic here. |
| `handlers` | Decode HTTP request, call validator, call model, encode response. No SQL here. |
| `validators` | Pure functions: accept a struct, return `[]ValidationError`. No HTTP, no DB. |
| `middleware` | HTTP middleware functions (CORS). Accept and return `http.Handler`. |

### 5.2 Go Struct Definitions

```go
// models/student.go
type Student struct {
    ID         int    `json:"id"`
    StudentID  string `json:"student_id"`
    FullName   string `json:"full_name"`
    Email      string `json:"email"`
    Phone      string `json:"phone,omitempty"`
    Department string `json:"department"`
    Year       int    `json:"year"`
    Section    string `json:"section"`
    DOB        string `json:"dob"`
    CreatedAt  string `json:"created_at"`
}

// models/course.go
type Course struct {
    ID         int    `json:"id"`
    CourseCode string `json:"course_code"`
    CourseName string `json:"course_name"`
    Credits    int    `json:"credits"`
    Department string `json:"department"`
}

// models/enrollment.go
type Enrollment struct {
    ID         int    `json:"id"`
    StudentID  int    `json:"student_id"`
    CourseID   int    `json:"course_id"`
    EnrolledAt string `json:"enrolled_at"`
}

// models/marks.go
type Marks struct {
    ID         int     `json:"id"`
    StudentID  int     `json:"student_id"`
    CourseID   int     `json:"course_id"`
    Marks      float64 `json:"marks"`
    Grade      string  `json:"grade"`      // computed, not stored
    RecordedAt string  `json:"recorded_at"`
}

// models/attendance.go
type Attendance struct {
    ID            int     `json:"id"`
    StudentID     int     `json:"student_id"`
    CourseID      int     `json:"course_id"`
    TotalClasses  int     `json:"total_classes"`
    Attended      int     `json:"attended"`
    Percentage    float64 `json:"percentage"`  // computed, not stored
    RecordedAt    string  `json:"recorded_at"`
}
```

### 5.3 Routing Table

The router uses Go's standard `net/http` `ServeMux` (Go 1.22+ with method patterns).

```
Method  Pattern                             Handler
------  ----------------------------------  ----------------------------
GET     /                                   ServeFile(frontend/index.html)
GET     /css/...                            ServeDir(frontend/css)
GET     /js/...                             ServeDir(frontend/js)

POST    /api/v1/students                    students.Create
GET     /api/v1/students                    students.List
GET     /api/v1/students/{id}               students.GetByID
PUT     /api/v1/students/{id}               students.Update
DELETE  /api/v1/students/{id}               students.Delete
GET     /api/v1/students/{id}/courses       enrollments.GetCoursesForStudent
GET     /api/v1/students/{id}/marks         marks.GetForStudent
GET     /api/v1/students/{id}/attendance    attendance.GetForStudent

POST    /api/v1/courses                     courses.Create
GET     /api/v1/courses                     courses.List
GET     /api/v1/courses/{id}                courses.GetByID
PUT     /api/v1/courses/{id}                courses.Update
DELETE  /api/v1/courses/{id}                courses.Delete
GET     /api/v1/courses/{id}/students       enrollments.GetStudentsForCourse

POST    /api/v1/enrollments                 enrollments.Create
DELETE  /api/v1/enrollments/{id}            enrollments.Delete

POST    /api/v1/marks                       marks.Create
PUT     /api/v1/marks/{id}                  marks.Update

POST    /api/v1/attendance                  attendance.Create
GET     /api/v1/attendance/low              attendance.GetLowAttendance
PUT     /api/v1/attendance/{id}             attendance.Update

GET     /api/v1/dashboard                   dashboard.Get
```

### 5.4 Error Response Helper

```go
// handlers/response.go
type ErrorResponse struct {
    Error string `json:"error"`
    Code  string `json:"code"`
}

func writeError(w http.ResponseWriter, statusCode int, msg, code string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(ErrorResponse{Error: msg, Code: code})
}
```

### 5.5 Grade Calculation Function

```go
// models/marks.go
func CalculateGrade(marks float64) string {
    switch {
    case marks >= 90:
        return "A+"
    case marks >= 80:
        return "A"
    case marks >= 70:
        return "B"
    case marks >= 60:
        return "C"
    case marks >= 50:
        return "D"
    default:
        return "F"
    }
}
```

### 5.6 Attendance Percentage Function

```go
// models/attendance.go
func CalculateAttendancePercentage(total, attended int) float64 {
    if total == 0 {
        return 0.0
    }
    pct := (float64(attended) / float64(total)) * 100
    return math.Round(pct*100) / 100  // round to 2 decimal places
}
```

### 5.7 Configuration

```go
// config/config.go
type Config struct {
    Port                string  // default: "8080"
    DBPath              string  // default: "./studenthub.db"
    AttendanceThreshold float64 // default: 75.0
}

func Load() Config {
    // reads PORT, DB_PATH, ATTENDANCE_THRESHOLD from environment
    // falls back to defaults if not set
}
```

---

## 6. Frontend Design

### 6.1 Single-Page Application Pattern

The frontend is a single HTML file (`index.html`) that acts as an application shell. JavaScript handles routing by reading `window.location.hash` and rendering page content into a `<main id="app">` container. No full-page reloads after the initial load.

```
/#dashboard        → dashboard.js renders dashboard view
/#students         → students.js renders student list
/#students/new     → students.js renders add-student form
/#students/:id     → student-detail.js renders student profile
/#students/:id/edit → students.js renders edit form
/#courses          → courses.js renders course list
/#marks            → marks.js renders marks entry/view
/#attendance       → attendance.js renders attendance entry/view
```

### 6.2 Navigation Structure

```
Sidebar Navigation
├── Dashboard
├── Students
│   ├── All Students
│   └── Add Student
├── Courses
│   ├── All Courses
│   └── Add Course
├── Marks
└── Attendance
```

### 6.3 Page Designs

#### Dashboard Page
- 4 stat cards: Total Students, Total Courses, Avg Attendance %, Low Attendance Count
- "Low Attendance Students" table: Name, Department, Year, Attendance %
- "Recent Students" table: Name, Student ID, Department, Year, Added

#### Students List Page
- Search bar (live filter on type with debounce)
- Filter dropdowns: Department, Year
- Table: Student ID | Name | Email | Department | Year | Section | Actions (View, Edit, Delete)
- "Add Student" button → opens modal form

#### Student Detail Page
- Student info card (all fields)
- Enrolled Courses table with marks and attendance per course
- Average marks badge and overall attendance badge

#### Add / Edit Student Form (Modal)
- Fields: Student ID, Full Name, Email, Phone, Department (select), Year (select 1–6), Section, Date of Birth
- Inline validation errors shown below each field
- Submit / Cancel buttons

#### Courses List Page
- Table: Course Code | Course Name | Department | Credits | Enrolled Students | Actions
- "Add Course" button → opens modal form

#### Marks Page
- Select Student dropdown → loads their enrolled courses
- Select Course dropdown → pre-fills existing marks if any
- Marks input (0–100)
- Submit button
- Table of all marks records for selected student

#### Attendance Page
- Select Student dropdown → loads enrolled courses
- Select Course dropdown
- Total Classes input, Attended Classes input
- Computed percentage shown live
- Submit / Update button
- Table of attendance records for selected student

### 6.4 CSS Design System

```css
/* Color palette (CSS custom properties) */
:root {
  --color-primary:     #2563EB;  /* blue-600 */
  --color-primary-dk:  #1D4ED8;  /* blue-700 */
  --color-success:     #16A34A;  /* green-600 */
  --color-warning:     #D97706;  /* amber-600 */
  --color-danger:      #DC2626;  /* red-600 */
  --color-neutral-50:  #F9FAFB;
  --color-neutral-100: #F3F4F6;
  --color-neutral-200: #E5E7EB;
  --color-neutral-700: #374151;
  --color-neutral-900: #111827;
  --color-surface:     #FFFFFF;

  --font-sans: 'Inter', system-ui, sans-serif;
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.07);
}
```

### 6.5 Responsive Breakpoints

| Breakpoint | Width | Layout change |
|---|---|---|
| Mobile | < 640px | Sidebar collapses to top hamburger menu. Tables scroll horizontally. |
| Tablet | 640px – 1024px | Sidebar shown as icon-only strip. |
| Desktop | > 1024px | Full sidebar with labels. Two-column layouts where appropriate. |

### 6.6 API Module (`js/api.js`)

All HTTP calls go through a single `api.js` module that:
- Prepends the base URL (`/api/v1`)
- Sets `Content-Type: application/json`
- Parses responses
- Throws structured errors for non-2xx responses

```javascript
// js/api.js
const BASE = '/api/v1';

async function request(method, path, body = null) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(BASE + path, opts);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error', code: 'UNKNOWN' }));
    throw { status: res.status, ...err };
  }
  if (res.status === 204) return null;
  return res.json();
}

export const api = {
  get:    (path)         => request('GET',    path),
  post:   (path, body)   => request('POST',   path, body),
  put:    (path, body)   => request('PUT',    path, body),
  delete: (path)         => request('DELETE', path),
};
```

### 6.7 Client-Side Validation

Validation runs on form `submit` before any API call. Each validator returns an array of `{ field, message }` objects. If any are present, errors are displayed inline and the request is not sent.

```javascript
// js/components/form.js
export function displayErrors(errors) {
  clearErrors();
  errors.forEach(({ field, message }) => {
    const el = document.querySelector(`[data-error="${field}"]`);
    if (el) { el.textContent = message; el.hidden = false; }
  });
}
```

---

## 7. Database Design Details

### 7.1 Indexes

```sql
-- Improve search and filter performance
CREATE INDEX idx_students_department ON students(department);
CREATE INDEX idx_students_year       ON students(year);
CREATE INDEX idx_students_student_id ON students(student_id);
CREATE INDEX idx_enrollments_student ON enrollments(student_id);
CREATE INDEX idx_enrollments_course  ON enrollments(course_id);
CREATE INDEX idx_marks_student       ON marks(student_id);
CREATE INDEX idx_attendance_student  ON attendance(student_id);
```

### 7.2 Foreign Key Enforcement

SQLite foreign keys are **off by default**. The DB initialization must run:

```sql
PRAGMA foreign_keys = ON;
```

This pragma must be set for every new connection (not persisted in the file).

### 7.3 Migration Strategy

- Migrations are numbered SQL files in `db/migrations/`.
- On startup, the server reads all migration files in order and executes them within a transaction.
- A `schema_migrations` table tracks which migrations have run.
- For v1, a single `001_init.sql` creates all tables, indexes, and the migrations table.

---

## 8. Security Design

| Concern | Mitigation |
|---|---|
| SQL injection | All queries use parameterized statements (`?` placeholders). No string-interpolated SQL. |
| Sensitive config | Port, DB path, and threshold come from environment variables. No hardcoded values. |
| Error leakage | Internal errors are logged server-side only. Clients receive generic messages for 500s. |
| Input sanitization | All string inputs are validated and length-capped before DB write. |
| CORS | Middleware restricts allowed origins. In development, `localhost` is allowed. |
| No auth (v1) | Explicitly deferred. Application is intended for a trusted internal network only. |

---

## 9. UI Component Inventory

| Component | Description |
|---|---|
| `<stat-card>` | Dashboard metric card with icon, label, and value. |
| `<data-table>` | Renders an array of objects as a table with configurable columns and action buttons. |
| `<modal>` | Overlay dialog for add/edit forms. Traps keyboard focus. Closeable by Escape key. |
| `<toast>` | Non-blocking notification (success/error/info) that auto-dismisses after 4 seconds. |
| `<search-bar>` | Input with debounce (300ms) that triggers a callback on change. |
| `<filter-bar>` | Row of select dropdowns for department and year filters. |
| `<grade-badge>` | Colored pill showing a letter grade. Color varies by grade. |
| `<attendance-bar>` | Horizontal progress bar showing attendance %. Red if below threshold. |

---

## 10. Kiro Demonstration Touchpoints

This section maps Kiro features to specific parts of the project, for use when demonstrating StudentHub as a Kiro University example.

| Kiro Feature | Where It's Demonstrated |
|---|---|
| Spec-driven development | This spec itself. The three-document workflow: requirements → design → tasks. |
| Steering documents | `.kiro/steering/go-standards.md` — Go coding conventions for this project. `.kiro/steering/api-conventions.md` — REST API rules applied to all handlers. |
| Kiro Hooks | `PostFileSave` hook: run `go vet ./...` when any `.go` file is saved. `PostTaskExec` hook: run the test suite after each task is completed. |
| Property-based testing | `CalculateGrade` and `CalculateAttendancePercentage` tested with rapid/quick-check style tests that generate random valid inputs. |
| Kiro Powers | AWS Blocks Power for a future v2 cloud deployment of StudentHub on AWS Lambda + DynamoDB. |
| MCP | An MCP server exposing StudentHub's REST API as tools, allowing Kiro to query and mutate student data directly from the chat interface. |
| Custom agents | A "StudentHub Reviewer" custom agent that checks all new handler files against the API conventions steering document before they are committed. |
