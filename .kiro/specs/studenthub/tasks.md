# StudentHub — Implementation Plan

**Project:** StudentHub — Student Management System  
**Version:** 1.0  
**Context:** Kiro University — Spec-Driven Development Demonstration  
**Date:** 2026-10-02  
**Status:** Draft

---

## How to Read This Document

Tasks are organized into phases that build on each other. Each task has:

- A unique ID (e.g., `T-01`)
- A title and description
- Acceptance criteria that must pass before the task is considered done
- Dependencies on earlier tasks
- References to the relevant spec sections

Complete tasks in order within each phase. Do not start a phase until all tasks in the previous phase are verified.

---

## Phase 0 — Project Scaffold

> Goal: A runnable Go server that serves a placeholder page. No features yet.

---

### T-01 · Initialize Go module and project structure

**Description:**  
Create the Go module and the full directory skeleton as defined in `design.md § 2`.

**Steps:**
1. Create the `studenthub/` root directory inside the workspace.
2. Run `go mod init studenthub` to create `go.mod`.
3. Create all directories: `config/`, `db/db/migrations/`, `models/`, `handlers/`, `validators/`, `middleware/`, `frontend/css/`, `frontend/js/pages/`, `frontend/js/components/`, `tests/models/`, `tests/validators/`, `tests/handlers/`.
4. Create stub `main.go` that prints "StudentHub starting…" and exits cleanly.
5. Add `studenthub.db` to `.gitignore`.
6. Add a root-level `README.md` with project name, description, and "how to run" placeholder.

**Acceptance criteria:**
- [ ] `go build ./...` succeeds with no errors.
- [ ] All directories from `design.md § 2` exist.
- [ ] `.gitignore` contains `studenthub.db` and `db/migrations/seed.sql`.

**Dependencies:** None  
**Spec ref:** design.md § 2

---

### T-02 · Add SQLite dependency

**Description:**  
Add the `mattn/go-sqlite3` driver to the module.

**Steps:**
1. Run `go get github.com/mattn/go-sqlite3`.
2. Verify `go.sum` is updated.
3. Create `db/db.go` with an `Open(path string) (*sql.DB, error)` function that opens a SQLite connection, sets `PRAGMA foreign_keys = ON`, and configures `SetMaxOpenConns(1)` (SQLite is single-writer).

**Acceptance criteria:**
- [ ] `go build ./...` succeeds.
- [ ] `db.Open(":memory:")` returns a non-nil `*sql.DB` in a simple test.

**Dependencies:** T-01  
**Spec ref:** design.md § 7.2

---

### T-03 · Create database schema and migration runner

**Description:**  
Write the SQL migration and the Go runner that applies it on startup.

**Steps:**
1. Create `db/migrations/001_init.sql` with all `CREATE TABLE` and `CREATE INDEX` statements from `design.md § 3` and `§ 7.1`.
2. Include the `schema_migrations` table in the migration file.
3. Implement `db.RunMigrations(db *sql.DB, migrationsDir string) error` in `db/db.go` that:
   - Reads migration files in filename order.
   - Skips any already recorded in `schema_migrations`.
   - Runs each in a transaction; rolls back and returns error on failure.
   - Records applied migrations in `schema_migrations`.
4. Create `db/migrations/seed.sql` with 5–10 sample students, 5 courses, enrollments, marks, and attendance records for development use only.

**Acceptance criteria:**
- [ ] Running the migration twice is idempotent (no error on second run).
- [ ] All 5 tables (`students`, `courses`, `enrollments`, `marks`, `attendance`) and `schema_migrations` exist after migration.
- [ ] All indexes from `design.md § 7.1` exist after migration.
- [ ] Foreign key constraint is verified: inserting an enrollment with a non-existent `student_id` fails.

**Dependencies:** T-02  
**Spec ref:** design.md § 3, § 7

---

### T-04 · Configuration and HTTP server bootstrap

**Description:**  
Implement `config/config.go` and wire the HTTP server in `main.go`.

**Steps:**
1. Implement `config.Load()` reading `PORT` (default `"8080"`), `DB_PATH` (default `"./studenthub.db"`), `ATTENDANCE_THRESHOLD` (default `75.0`) from environment.
2. In `main.go`: call `config.Load()`, call `db.Open()`, call `db.RunMigrations()`, register a placeholder `GET /` handler that returns `200 OK` with `"StudentHub API running"`, start `http.ListenAndServe`.
3. Implement `middleware/cors.go` with a `CORS(next http.Handler) http.Handler` that adds permissive headers for development.
4. Apply CORS middleware to all `/api/` routes.

**Acceptance criteria:**
- [ ] `go run main.go` starts without error.
- [ ] `GET http://localhost:8080/` returns `200 OK`.
- [ ] Setting `PORT=9090` makes the server listen on 9090.
- [ ] Response to any `/api/` route includes `Access-Control-Allow-Origin` header.

**Dependencies:** T-03  
**Spec ref:** design.md § 5.1, § 5.7, § 8

---

### T-05 · Serve static frontend files

**Description:**  
Create the frontend shell and have Go serve it.

**Steps:**
1. Create `frontend/index.html` as the SPA shell: `<nav>` sidebar, `<main id="app">` content area, script tag loading `js/app.js` as a module.
2. Create placeholder `frontend/css/main.css` with CSS variables from `design.md § 6.4`.
3. Create placeholder `frontend/js/app.js` that renders `"Welcome to StudentHub"` into `#app`.
4. In `main.go`, register `GET /` to serve `frontend/index.html` and `GET /css/...`, `GET /js/...` to serve the respective directories using `http.FileServer`.

**Acceptance criteria:**
- [ ] Opening `http://localhost:8080/` in a browser shows the sidebar and "Welcome to StudentHub".
- [ ] CSS variables are defined in `main.css` and the page has no unstyled/raw HTML appearance.
- [ ] Browser DevTools shows no 404 errors for CSS or JS files.

**Dependencies:** T-04  
**Spec ref:** design.md § 6.1, § 6.4

---

## Phase 1 — Student Management (Backend)

> Goal: Full CRUD REST API for students, with validation and tests.

---

### T-06 · Student model — structs and DB queries

**Description:**  
Implement `models/student.go` with all database query functions.

**Steps:**
1. Define the `Student` struct with correct JSON tags (see `design.md § 5.2`).
2. Implement:
   - `CreateStudent(db, s Student) (Student, error)` — INSERT, return with generated `id` and `created_at`.
   - `GetAllStudents(db, filter StudentFilter) ([]Student, error)` — SELECT with optional WHERE clauses for `department`, `year`, and LIKE search on `full_name`/`email`/`student_id`.
   - `GetStudentByID(db, id int) (Student, error)` — SELECT by PK; return `ErrNotFound` if missing.
   - `UpdateStudent(db, id int, s Student) (Student, error)` — UPDATE by PK.
   - `DeleteStudent(db, id int) error` — DELETE by PK.
3. Define `ErrNotFound` and `ErrDuplicate` sentinel errors in `models/errors.go`.
4. All queries use parameterized statements only.

**Acceptance criteria:**
- [ ] Each function can be exercised against an in-memory SQLite DB in a table-driven test.
- [ ] `GetStudentByID` with a missing ID returns `ErrNotFound`.
- [ ] `CreateStudent` with a duplicate `student_id` returns `ErrDuplicate`.
- [ ] `GetAllStudents` with `department="Computer Science"` returns only CS students.

**Dependencies:** T-03  
**Spec ref:** requirements.md § 5.2, design.md § 3.2, § 5.2

---

### T-07 · Student validator

**Description:**  
Implement `validators/student_validator.go`.

**Steps:**
1. Define `ValidationError struct { Field string; Message string }`.
2. Implement `ValidateStudent(s models.Student) []ValidationError` enforcing all rules from `requirements.md § 7.1`:
   - Student ID: required, 3–20 chars, alphanumeric + hyphens.
   - Full Name: required, 2–100 chars, allowed characters.
   - Email: required, valid format, max 255 chars.
   - Phone: optional; if present, matches phone pattern.
   - Department: required, must be in the defined list.
   - Year: required, integer 1–6.
   - Section: required, 1–10 chars, alphanumeric.
   - DOB: required, valid `YYYY-MM-DD`, student age ≥ 15 years.
3. Return all failing validations at once (not just the first).

**Acceptance criteria:**
- [ ] Valid student input returns an empty slice.
- [ ] Each invalid field returns exactly one `ValidationError` with the correct `Field` and a non-empty `Message`.
- [ ] A student with DOB set to today's date returns a validation error.
- [ ] A student with DOB giving age of exactly 15 years is valid.
- [ ] An invalid department value returns a validation error.

**Dependencies:** T-06  
**Spec ref:** requirements.md § 7.1

---

### T-08 · Student HTTP handlers

**Description:**  
Implement `handlers/students.go` wiring validators and models to HTTP.

**Steps:**
1. Implement handler functions for all 5 student routes (POST, GET list, GET by ID, PUT, DELETE) as defined in `design.md § 4.2` and § 5.3`.
2. Each handler must:
   - Decode and validate request body (POST, PUT).
   - Call the appropriate model function.
   - Map `ErrNotFound` → 404, `ErrDuplicate` → 409, validation errors → 400, all others → 500.
   - Return JSON with correct HTTP status codes.
3. Implement the `writeError` and `writeJSON` helpers in `handlers/response.go`.
4. Register all student routes in `main.go`.

**Acceptance criteria:**
- [ ] `POST /api/v1/students` with valid body returns 201 and the created student.
- [ ] `POST /api/v1/students` with missing `full_name` returns 400 with a field-level error message.
- [ ] `POST /api/v1/students` with duplicate `student_id` returns 409.
- [ ] `GET /api/v1/students?department=Computer+Science` returns only CS students.
- [ ] `GET /api/v1/students?search=aisha` returns students whose name contains "aisha" (case-insensitive).
- [ ] `GET /api/v1/students/999` returns 404.
- [ ] `DELETE /api/v1/students/:id` returns 204.

**Dependencies:** T-06, T-07  
**Spec ref:** design.md § 4.2, § 5.3, § 5.4, requirements.md § 8

---

### T-09 · Student handler integration tests

**Description:**  
Write integration tests for student handlers using `net/http/httptest`.

**Steps:**
1. Create `tests/handlers/students_handler_test.go`.
2. Write table-driven tests for each endpoint covering happy path and at least two error paths per handler.
3. Use an in-memory SQLite DB per test (isolated).
4. Test that deleting a student cascades to marks and attendance (verify via a subsequent SELECT).

**Acceptance criteria:**
- [ ] `go test ./tests/handlers/...` passes with no failures.
- [ ] At least 15 test cases covering all 5 handlers.
- [ ] Cascade delete verified: student deletion removes associated marks and attendance rows.

**Dependencies:** T-08  
**Spec ref:** requirements.md § 9 (AC-STU), § 10 (TR-05)

---

## Phase 2 — Course Management (Backend)

> Goal: Full CRUD REST API for courses, with validation and tests.

---

### T-10 · Course model and validator

**Description:**  
Implement `models/course.go` and `validators/course_validator.go`.

**Steps:**
1. Define `Course` struct (see `design.md § 5.2`).
2. Implement: `CreateCourse`, `GetAllCourses`, `GetCourseByID`, `UpdateCourse`, `DeleteCourse` — same pattern as T-06.
3. Implement `ValidateCourse(c models.Course) []ValidationError` enforcing all rules from `requirements.md § 7.2`.

**Acceptance criteria:**
- [ ] `CreateCourse` with duplicate `course_code` returns `ErrDuplicate`.
- [ ] `ValidateCourse` with `credits = 7` returns a validation error.
- [ ] `ValidateCourse` with a lowercase course code (e.g., `"cs101"`) returns a validation error (must be uppercase).

**Dependencies:** T-03  
**Spec ref:** requirements.md § 5.4, § 7.2, design.md § 3.3

---

### T-11 · Course HTTP handlers and tests

**Description:**  
Implement `handlers/courses.go` and `tests/handlers/courses_handler_test.go`.

**Steps:**
1. Implement handlers for all 5 course routes (POST, GET list, GET by ID, PUT, DELETE).
2. Register routes in `main.go`.
3. Write integration tests (same pattern as T-09): happy path + error paths per handler.

**Acceptance criteria:**
- [ ] `POST /api/v1/courses` with valid body returns 201.
- [ ] `POST /api/v1/courses` with duplicate `course_code` returns 409.
- [ ] `DELETE /api/v1/courses/:id` returns 204 and cascades enrollments/marks/attendance.
- [ ] `go test ./tests/handlers/...` passes.

**Dependencies:** T-10, T-08  
**Spec ref:** design.md § 4.3, requirements.md § 9 (AC-CRS)

---

## Phase 3 — Enrollment, Marks, and Attendance (Backend)

> Goal: All remaining backend API endpoints working with tests.

---

### T-12 · Enrollment model and handlers

**Description:**  
Implement `models/enrollment.go` and `handlers/enrollments.go`.

**Steps:**
1. Define `Enrollment` struct.
2. Implement: `CreateEnrollment`, `GetCoursesForStudent`, `GetStudentsForCourse`, `DeleteEnrollment`.
3. `CreateEnrollment` must verify both student and course exist before inserting (return `ErrNotFound` otherwise) and return `ErrDuplicate` on a duplicate pair.
4. Implement handlers for all 4 enrollment routes (see `design.md § 4.4`).
5. Register routes in `main.go`.

**Acceptance criteria:**
- [ ] `POST /api/v1/enrollments` with valid student+course IDs returns 201.
- [ ] Enrolling the same student+course twice returns 409.
- [ ] Enrolling with a non-existent `student_id` returns 404.
- [ ] `GET /api/v1/students/:id/courses` returns enrolled courses.
- [ ] `DELETE /api/v1/enrollments/:id` returns 204.

**Dependencies:** T-08, T-11  
**Spec ref:** design.md § 4.4, requirements.md § 5.5 (FR-ENR)

---

### T-13 · Marks model, validator, and handlers

**Description:**  
Implement `models/marks.go`, `validators/marks_validator.go`, and `handlers/marks.go`.

**Steps:**
1. Define `Marks` struct with computed `Grade` field.
2. Implement `CalculateGrade(marks float64) string` as specified in `design.md § 5.5`.
3. Implement: `CreateMarks`, `GetMarksForStudent`, `UpdateMarks`.
4. `CreateMarks` must verify the student is enrolled in the course (query `enrollments` table); return 400 if not enrolled.
5. `GetMarksForStudent` returns all marks for a student including computed grade and the student's average marks.
6. Implement `ValidateMarks` enforcing `marks` is in [0.0, 100.0].
7. Implement handlers for the 3 marks routes (see `design.md § 4.5`).
8. Register routes in `main.go`.

**Acceptance criteria:**
- [ ] `CalculateGrade(95)` → `"A+"`, `CalculateGrade(82)` → `"A"`, `CalculateGrade(73)` → `"B"`, `CalculateGrade(65)` → `"C"`, `CalculateGrade(55)` → `"D"`, `CalculateGrade(40)` → `"F"`.
- [ ] `CalculateGrade(90)` → `"A+"` (boundary: inclusive lower bound of A+).
- [ ] `CalculateGrade(80)` → `"A"` (boundary).
- [ ] `POST /api/v1/marks` for a non-enrolled student returns 400.
- [ ] `POST /api/v1/marks` with `marks = 101` returns 400.
- [ ] Average marks is correctly computed when a student has marks in 3 courses.

**Dependencies:** T-12  
**Spec ref:** design.md § 4.5, § 5.5, requirements.md § 5.3, § 5.6, § 7.3

---

### T-14 · Attendance model, validator, and handlers

**Description:**  
Implement `models/attendance.go`, `validators/attendance_validator.go`, and `handlers/attendance.go`.

**Steps:**
1. Define `Attendance` struct with computed `Percentage` field.
2. Implement `CalculateAttendancePercentage(total, attended int) float64` as specified in `design.md § 5.6`.
3. Implement: `CreateAttendance`, `GetAttendanceForStudent`, `UpdateAttendance`, `GetLowAttendanceStudents(db, threshold float64) ([]LowAttendanceStudent, error)`.
4. `CreateAttendance` must verify enrollment (same pattern as marks).
5. Implement `ValidateAttendance`: `total_classes ≥ 1`, `attended ≥ 0`, `attended ≤ total_classes`.
6. Implement handlers for the 4 attendance routes (see `design.md § 4.6`).
7. Register routes in `main.go`.

**Acceptance criteria:**
- [ ] `CalculateAttendancePercentage(48, 36)` → `75.00`.
- [ ] `CalculateAttendancePercentage(0, 0)` → `0.0` (no panic, no divide-by-zero).
- [ ] `CalculateAttendancePercentage(3, 1)` → `33.33` (2 decimal precision).
- [ ] `POST /api/v1/attendance` with `attended > total_classes` returns 400.
- [ ] `GET /api/v1/attendance/low` returns only students below the 75% threshold.
- [ ] A student with 75.00% attendance is NOT in the low-attendance list (threshold is exclusive: `< 75`).

**Dependencies:** T-12  
**Spec ref:** design.md § 4.6, § 5.6, requirements.md § 5.7, § 7.4

---

### T-15 · Dashboard endpoint

**Description:**  
Implement `handlers/dashboard.go` and `GET /api/v1/dashboard`.

**Steps:**
1. Implement a single DB query (or multiple coordinated queries) to fetch: total student count, total course count, average attendance across all records, low-attendance students (below threshold), 5 most recent students by `created_at`.
2. Return the combined response as defined in `design.md § 4.7`.
3. Register the route in `main.go`.

**Acceptance criteria:**
- [ ] Response includes all 5 top-level keys: `total_students`, `total_courses`, `average_attendance`, `low_attendance_students`, `recent_students`.
- [ ] `total_students` matches the actual count in the DB.
- [ ] `recent_students` contains at most 5 entries in descending `created_at` order.
- [ ] A student with 60% attendance appears in `low_attendance_students`.

**Dependencies:** T-13, T-14  
**Spec ref:** design.md § 4.7, requirements.md § 5.8

---

### T-16 · Student detail profile endpoint

**Description:**  
Implement `GET /api/v1/students/:id` full profile response that joins courses, marks, and attendance.

**Steps:**
1. Extend the student handler for GET by ID to return the full profile shape from `design.md § 4.2`.
2. Write a SQL query (or coordinated queries) that joins `enrollments`, `courses`, `marks`, and `attendance` for the student.
3. Compute `grade` from marks and `attendance_percentage` from attendance inline.
4. Compute `average_marks` and `overall_attendance` across all enrolled courses.

**Acceptance criteria:**
- [ ] Response for a student enrolled in 2 courses includes both courses in the `courses` array.
- [ ] `average_marks` is the arithmetic mean of marks across all courses (null if no marks recorded).
- [ ] A course with no marks recorded shows `marks: null, grade: null`.
- [ ] A course with no attendance recorded shows `attendance_percentage: null`.

**Dependencies:** T-13, T-14  
**Spec ref:** design.md § 4.2 (GET /students/:id)

---

### T-17 · Unit tests for calculation functions and validators

**Description:**  
Write comprehensive unit and property-based tests for pure functions.

**Steps:**
1. Create `tests/models/marks_test.go`: table-driven tests for `CalculateGrade` covering all grade boundaries (including exact boundary values). Add a property-based test using `testing/quick` that confirms `CalculateGrade` never returns an empty string for any float in [0, 100].
2. Create `tests/models/attendance_test.go`: table-driven tests for `CalculateAttendancePercentage` including edge cases. Add a property-based test confirming the result is always in [0, 100] for valid inputs.
3. Create `tests/validators/student_validator_test.go`: table-driven tests for every validation rule in `ValidateStudent`.
4. Create `tests/validators/marks_validator_test.go` and `tests/validators/course_validator_test.go`.

**Acceptance criteria:**
- [ ] `go test ./tests/...` passes with no failures.
- [ ] Coverage of `CalculateGrade` is 100% (all branches exercised).
- [ ] Coverage of `CalculateAttendancePercentage` is 100%.
- [ ] Property-based tests run at least 100 random cases each (default for `testing/quick`).
- [ ] All grade boundary values are explicitly tested: 90, 80, 70, 60, 50, 49.

**Dependencies:** T-13, T-14, T-07, T-10  
**Spec ref:** requirements.md § 10 (TR-01 through TR-07)

---

## Phase 4 — Frontend Implementation

> Goal: A fully functional, responsive single-page application consuming the backend API.

---

### T-18 · Frontend shell, navigation, and routing

**Description:**  
Build the complete application shell with working client-side routing.

**Steps:**
1. Finalize `frontend/index.html`: semantic structure with `<nav>`, `<aside>` sidebar, `<main id="app">`, skip-to-content link for accessibility.
2. Implement the sidebar navigation with links for Dashboard, Students, Courses, Marks, Attendance — highlight the active link.
3. Implement hash-based router in `frontend/js/app.js` mapping hash patterns to page module imports (see `design.md § 6.1`).
4. Implement `frontend/js/api.js` exactly as specified in `design.md § 6.6`.
5. Implement `frontend/js/components/toast.js` for success/error notifications.
6. Implement `frontend/js/components/modal.js` (opens/closes overlay, traps focus, closes on Escape).
7. Apply full CSS: layout, typography, sidebar, CSS variables, responsive breakpoints from `design.md § 6.4–6.5`.
8. Implement hamburger toggle for mobile sidebar.

**Acceptance criteria:**
- [ ] Navigating to `/#students` loads the students page without a full page reload.
- [ ] The active nav item is highlighted correctly on every route change.
- [ ] Modal opens on button click and closes on Escape key.
- [ ] Toast notification appears and auto-dismisses after 4 seconds.
- [ ] At 375px width, the sidebar collapses and the hamburger menu works.
- [ ] No horizontal scrollbar appears at 375px width.
- [ ] All form labels are associated with inputs via `for`/`id` attributes.

**Dependencies:** T-05, T-08 (backend must be running)  
**Spec ref:** design.md § 6.1–6.5

---

### T-19 · Dashboard page

**Description:**  
Implement `frontend/js/pages/dashboard.js`.

**Steps:**
1. On load, call `GET /api/v1/dashboard`.
2. Render 4 stat cards: Total Students, Total Courses, Avg Attendance %, Low Attendance Count.
3. Render "Low Attendance Students" table: Name, Department, Year, Attendance % (red-highlighted if below 75%).
4. Render "Recent Students" table: Name, Student ID, Department, Year, Added date.
5. Show a loading skeleton while data is fetching.
6. Show an error state if the API call fails.

**Acceptance criteria:**
- [ ] Dashboard loads and displays real data from the API.
- [ ] Stat card numbers update when the database changes and the page is refreshed.
- [ ] A student with 60% attendance appears in the low-attendance table with red highlighting.
- [ ] The "Recent Students" table shows at most 5 entries.

**Dependencies:** T-18, T-15  
**Spec ref:** design.md § 6.3 (Dashboard Page), requirements.md § 5.8

---

### T-20 · Students list page

**Description:**  
Implement `frontend/js/pages/students.js` (list view).

**Steps:**
1. On load, call `GET /api/v1/students` and render results in a table.
2. Columns: Student ID | Full Name | Email | Department | Year | Section | Actions (View, Edit, Delete).
3. Implement search bar with 300ms debounce calling `GET /api/v1/students?search=...`.
4. Implement Department and Year filter dropdowns calling `GET /api/v1/students?department=...&year=...`.
5. "Add Student" button opens a modal with the add-student form.
6. Delete button shows a confirmation dialog before calling `DELETE /api/v1/students/:id`.
7. On successful add/delete, refresh the list and show a success toast.
8. Show empty state message when no students match the current filters.

**Acceptance criteria:**
- [ ] Student list renders all students from the API.
- [ ] Typing in the search bar filters results within 500ms.
- [ ] Selecting a department filter returns only students in that department.
- [ ] Clicking Delete and confirming removes the student from the list.
- [ ] Empty state message is shown when no results match.

**Dependencies:** T-18, T-08  
**Spec ref:** design.md § 6.3, requirements.md US-02, US-05, US-06, US-07

---

### T-21 · Add and edit student forms

**Description:**  
Implement the add-student modal form and the edit-student page within `students.js`.

**Steps:**
1. Build the form with all 8 fields (see `requirements.md § 7.1`) using appropriate input types.
2. Department → `<select>` with the defined department list.
3. Year → `<select>` with options 1–6.
4. Implement client-side validation on submit — call `validateStudentForm(data)` before any API call.
5. Display inline field-level errors (below each input) for validation failures.
6. On server-returned 400/409, parse the error and display the appropriate message.
7. Edit form pre-fills current values; submits `PUT /api/v1/students/:id`.
8. Success closes the modal and shows a toast.

**Acceptance criteria:**
- [ ] Submitting an empty form shows required-field errors on all required fields.
- [ ] An invalid email shows an inline error on the email field specifically.
- [ ] A valid form submission creates the student and shows it in the list.
- [ ] Edit form is pre-filled with the student's current values.
- [ ] A 409 from the server (duplicate student ID) shows the error message near the Student ID field.

**Dependencies:** T-20  
**Spec ref:** design.md § 6.3, § 6.7, requirements.md § 7.1, NFR-03

---

### T-22 · Student detail page

**Description:**  
Implement `frontend/js/pages/student-detail.js`.

**Steps:**
1. On load, call `GET /api/v1/students/:id` full profile endpoint.
2. Render a student info card with all fields.
3. Render an enrolled courses table: Course Code | Course Name | Credits | Marks | Grade (badge) | Attendance % (bar) | Actions (Remove enrollment).
4. Display average marks and overall attendance as summary badges.
5. "Assign Course" button opens a dropdown of available courses (those the student is NOT enrolled in) and submits `POST /api/v1/enrollments`.
6. "Remove" enrollment button calls `DELETE /api/v1/enrollments/:id` with confirmation.

**Acceptance criteria:**
- [ ] Student detail page shows all profile fields correctly.
- [ ] Grade badge is colored: A+/A = green, B = blue, C = amber, D = orange, F = red.
- [ ] Attendance bar is red when below 75%, green otherwise.
- [ ] Assigning a course adds it to the enrolled courses table without a full reload.
- [ ] Removing an enrollment removes it from the table.

**Dependencies:** T-18, T-16, T-12  
**Spec ref:** design.md § 6.3 (Student Detail Page), requirements.md US-03, US-12, US-13

---

### T-23 · Courses page

**Description:**  
Implement `frontend/js/pages/courses.js`.

**Steps:**
1. On load, call `GET /api/v1/courses` and render in a table.
2. Columns: Course Code | Course Name | Department | Credits | Actions (Edit, Delete).
3. "Add Course" opens a modal form with client-side validation.
4. Edit and delete follow the same pattern as students.

**Acceptance criteria:**
- [ ] Course list renders correctly.
- [ ] Add course form validates required fields and unique course code (client-side format check; server validates uniqueness).
- [ ] Deleting a course shows a confirmation and removes it from the list.

**Dependencies:** T-18, T-11  
**Spec ref:** design.md § 6.3 (Courses Page), requirements.md US-08 through US-11

---

### T-24 · Marks page

**Description:**  
Implement `frontend/js/pages/marks.js`.

**Steps:**
1. Render a student selector dropdown (populated from `GET /api/v1/students`).
2. On student selection, populate a course dropdown from `GET /api/v1/students/:id/courses`.
3. On course selection, pre-fill existing marks if any (from the student's marks data).
4. Marks input (number, 0–100) with live grade preview label.
5. Submit calls `POST /api/v1/marks` (create) or `PUT /api/v1/marks/:id` (update if exists).
6. Below the form, render a table of all marks for the selected student.

**Acceptance criteria:**
- [ ] Selecting a student and course shows existing marks if recorded.
- [ ] The grade preview label updates live as the user types a marks value.
- [ ] Submitting marks outside 0–100 shows an inline validation error.
- [ ] The marks table updates after a successful submission.

**Dependencies:** T-18, T-13  
**Spec ref:** design.md § 6.3 (Marks Page), requirements.md US-14 through US-17

---

### T-25 · Attendance page

**Description:**  
Implement `frontend/js/pages/attendance.js`.

**Steps:**
1. Same student+course selector pattern as marks page.
2. Total Classes and Attended Classes inputs.
3. Live computed attendance percentage displayed below inputs.
4. Submit calls `POST /api/v1/attendance` or `PUT /api/v1/attendance/:id`.
5. Below the form, render an attendance table for the selected student.
6. Highlight rows where percentage < 75% in amber/red.

**Acceptance criteria:**
- [ ] Attendance percentage updates live as inputs change.
- [ ] Entering `attended > total_classes` shows a client-side validation error.
- [ ] Below-threshold rows are visually highlighted in the table.
- [ ] Existing attendance record is pre-filled when a student+course with a record is selected.

**Dependencies:** T-18, T-14  
**Spec ref:** design.md § 6.3 (Attendance Page), requirements.md US-18 through US-21

---

## Phase 5 — Polish, Validation, and Kiro Demonstration Artifacts

> Goal: Final quality pass and creation of Kiro-specific project artifacts.

---

### T-26 · Accessibility and responsive layout audit

**Description:**  
Verify and fix all accessibility and responsive layout issues.

**Steps:**
1. Audit all forms: verify every `<input>` has a matching `<label for="">`.
2. Verify all interactive elements are keyboard-reachable and have visible focus rings.
3. Verify modal traps focus (Tab does not leave the modal while it is open).
4. Check color contrast for all text using browser DevTools or a contrast checker; ensure WCAG 2.1 AA.
5. Test at 375px (mobile), 768px (tablet), and 1280px (desktop) in browser DevTools.
6. Fix any horizontal overflow at 375px.
7. Add `aria-label` or `aria-describedby` to icon-only buttons.

**Acceptance criteria:**
- [ ] Zero inputs without associated labels (verified by DOM inspection).
- [ ] All buttons and links have visible focus outlines.
- [ ] Modal focus trap works: Tab from last element returns to first element inside modal.
- [ ] No horizontal scrollbar at 375px.
- [ ] All body text passes WCAG AA contrast ratio (4.5:1 minimum).

**Dependencies:** T-18 through T-25  
**Spec ref:** requirements.md NFR-01, NFR-02

---

### T-27 · End-to-end validation and error handling review

**Description:**  
Verify that all validation and error handling requirements are met end-to-end.

**Steps:**
1. Manually test each form for every validation rule in `requirements.md § 7`.
2. Verify that server-side 400 errors with field messages are displayed inline on the correct fields.
3. Verify that 409 (duplicate) errors are displayed near the relevant field.
4. Simulate a network failure (disable Go server) and verify the frontend shows a toast error, not a raw error.
5. Verify that no Go error strings or stack traces are visible in the browser.
6. Test that cascade deletes work: delete a student, verify their marks and attendance are gone via direct DB inspection.

**Acceptance criteria:**
- [ ] All acceptance criteria in `requirements.md § 9` are verified as passing.
- [ ] Network failure shows a user-friendly error toast, not `undefined` or a raw stack trace.
- [ ] Cascade delete confirmed for both student and course deletion.

**Dependencies:** T-25  
**Spec ref:** requirements.md § 7, § 8, § 9

---

### T-28 · Create Kiro steering documents

**Description:**  
Create steering documents for the project to guide ongoing development.

**Steps:**
1. Create `.kiro/steering/go-standards.md` covering:
   - Package organization rules for this project.
   - Naming conventions (handlers, models, validators).
   - Error handling pattern (use sentinel errors, not fmt.Errorf wrapping for known errors).
   - No direct SQL in handlers — only models.
   - All queries must use parameterized statements.
2. Create `.kiro/steering/api-conventions.md` covering:
   - REST URL patterns used in this project.
   - JSON field naming (snake_case).
   - Standard error response shape.
   - HTTP status code usage.
   - No business logic in handlers.
3. Create `.kiro/steering/frontend-conventions.md` covering:
   - All API calls go through `api.js`.
   - No inline styles — CSS classes only.
   - Client-side validation must mirror server-side rules.
   - Accessibility requirements for new components.

**Acceptance criteria:**
- [ ] Three steering files exist in `.kiro/steering/`.
- [ ] Each file is specific enough to guide an AI agent working on this codebase.
- [ ] The Go standards doc references actual package names in this project.

**Dependencies:** T-27  
**Spec ref:** design.md § 10

---

### T-29 · Create Kiro hooks

**Description:**  
Create Kiro hooks to automate quality checks during development.

**Steps:**
1. Create a `PostFileSave` hook (`.kiro/hooks/go-vet-on-save.json`) that runs `go vet ./...` whenever a `.go` file is saved. Display the output in the terminal.
2. Create a `PostTaskExec` hook (`.kiro/hooks/test-after-task.json`) that runs `go test ./tests/...` after each spec task is marked complete.
3. Create a `PostFileSave` hook (`.kiro/hooks/lint-frontend-on-save.json`) that runs a simple syntax check on `.js` files (using `node --check`) when a JS file is saved.

**Acceptance criteria:**
- [ ] Saving a `.go` file with a syntax error surfaces the `go vet` output.
- [ ] All three hook files exist and are valid JSON conforming to the hook schema.

**Dependencies:** T-27  
**Spec ref:** design.md § 10

---

### T-30 · Final integration verification

**Description:**  
Run all tests, verify the full application end-to-end, and confirm the spec is complete.

**Steps:**
1. Run `go test ./...` — all tests must pass.
2. Start the server with `go run main.go`.
3. Walk through every user story in `requirements.md § 4` and verify each works in the browser.
4. Verify all acceptance criteria in `requirements.md § 9` are met.
5. Confirm no hardcoded student/course data exists in production code (`grep` for known test names).
6. Confirm `studenthub.db` is in `.gitignore`.
7. Confirm no credentials or secrets are present in any source file.

**Acceptance criteria:**
- [ ] `go test ./...` passes with zero failures.
- [ ] All 23 user stories (US-01 through US-23) are demonstrably working.
- [ ] All acceptance criteria sections (AC-STU, AC-CRS, AC-ENR, AC-MRK, AC-ATT, AC-DSH, AC-UI) pass.
- [ ] `grep -r "hardcoded\|password\|secret\|api_key" --include="*.go" .` returns no matches in production code.
- [ ] The `seed.sql` file is not served or executed in production (protected by build tag or env check).

**Dependencies:** T-29  
**Spec ref:** requirements.md § 9, § 11

---

## Summary: Task Dependency Graph

```
T-01 → T-02 → T-03 → T-04 → T-05
                  │
                  ├─→ T-06 → T-07 → T-08 → T-09
                  │                    │
                  └─→ T-10 → T-11 ────┤
                                       │
                                       ├─→ T-12 → T-13 → T-17
                                       │       └─→ T-14 ──┘
                                       │             │
                                       │             └─→ T-15 → T-16
                                       │
                                       └──────────────────────────────┐
                                                                       │
T-05 → T-18 → T-19                                                    │
          └─→ T-20 → T-21 → T-22                                      │
          └─→ T-23                                         (requires backend)
          └─→ T-24
          └─→ T-25
               └─→ T-26 → T-27 → T-28 → T-29 → T-30
```

---

## Phase Summary

| Phase | Tasks | Focus |
|---|---|---|
| 0 | T-01 – T-05 | Project scaffold, DB, config, server bootstrap |
| 1 | T-06 – T-09 | Student CRUD backend + tests |
| 2 | T-10 – T-11 | Course CRUD backend + tests |
| 3 | T-12 – T-17 | Enrollment, Marks, Attendance backend + all unit tests |
| 4 | T-18 – T-25 | Full frontend: shell, all 6 pages, forms, validation |
| 5 | T-26 – T-30 | Accessibility audit, steering docs, hooks, final verification |

**Total tasks:** 30  
**Estimated implementation sessions:** 8–12 focused sessions
