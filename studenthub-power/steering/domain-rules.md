# StudentHub — Domain Business Rules

**Source of truth:** `.kiro/specs/studenthub/requirements.md`
**Load this file when:** implementing grade calculation, attendance percentage, validation logic, or any business rule for students, courses, marks, or attendance.

---

## Departments (Fixed List)

The department field on both students and courses must be one of these exact strings. No free text is accepted.

```
Computer Science
Information Technology
Electronics and Communication
Mechanical Engineering
Civil Engineering
Business Administration
Mathematics
Physics
```

This list is defined as a constant in both the Go backend and the JavaScript frontend. If the list changes, both must be updated together.

---

## Student Fields and Validation Rules

Source: `requirements.md § 7.1`

| Field | Type | Rules |
|---|---|---|
| `student_id` | string | Required. 3–20 chars. Alphanumeric + hyphens only. User-defined (e.g. `STU2024001`). Must be unique. |
| `full_name` | string | Required. 2–100 chars. Letters, spaces, hyphens, apostrophes only. |
| `email` | string | Required. Valid email format. Max 255 chars. Must be unique. |
| `phone` | string | Optional. If present: `+?[0-9\s\-()]{7,20}`. |
| `department` | string | Required. Must be from the fixed department list above. |
| `year` | int | Required. Integer 1–6 inclusive. |
| `section` | string | Required. 1–10 chars. Alphanumeric only. |
| `dob` | string | Required. `YYYY-MM-DD`. Student must be at least 15 years old at time of submission. |

**Uniqueness** (`student_id`, `email`) is enforced by the database — `models` returns `ErrDuplicate`, handlers return HTTP 409. Validators only check format and range.

---

## Course Fields and Validation Rules

Source: `requirements.md § 7.2`

| Field | Type | Rules |
|---|---|---|
| `course_code` | string | Required. 2–20 chars. Uppercase letters, digits, hyphens only. Must be unique. |
| `course_name` | string | Required. 3–150 chars. |
| `credits` | int | Required. Integer 1–6 inclusive. |
| `department` | string | Required. Must be from the fixed department list. |

Course `id` is system-generated (INTEGER PRIMARY KEY AUTOINCREMENT). `course_code` is the human-readable unique identifier (e.g. `CS101`).

---

## Marks Fields and Validation Rules

Source: `requirements.md § 7.3`, `§ 5.3`

| Field | Type | Rules |
|---|---|---|
| `student_id` | int | Required. Must reference an existing student (internal DB id). |
| `course_id` | int | Required. Must reference an existing course. |
| `marks` | float64 | Required. Range: 0.00 – 100.00 inclusive. |

**Only one marks record per (student_id, course_id) pair.** Attempting a second INSERT returns `ErrDuplicate` → HTTP 409.

**A marks record may only be created if the student is enrolled in the course.** The handler must verify enrollment before inserting. If not enrolled → HTTP 400.

### Grading Scale (requirements.md § 5.3)

Grades are **computed on every read** from the raw `marks` value. **Grades are never stored in the database.**

| Marks Range | Grade |
|---|---|
| 90 – 100 | A+ |
| 80 – 89  | A  |
| 70 – 79  | B  |
| 60 – 69  | C  |
| 50 – 59  | D  |
| 0  – 49  | F  |

The Go implementation lives in `models/marks.go`:

```go
func CalculateGrade(marks float64) string {
    switch {
    case marks >= 90: return "A+"
    case marks >= 80: return "A"
    case marks >= 70: return "B"
    case marks >= 60: return "C"
    case marks >= 50: return "D"
    default:          return "F"
    }
}
```

All six boundary values are explicitly tested in `tests/models/marks_test.go`. Property-based tests verify the grade is always a valid member of `{A+, A, B, C, D, F}` for any input in [0, 100].

---

## Attendance Fields and Validation Rules

Source: `requirements.md § 7.4`, `§ 5.7`

| Field | Type | Rules |
|---|---|---|
| `student_id` | int | Required. Must reference an existing student. |
| `course_id` | int | Required. Must reference an existing course. |
| `total_classes` | int | Required. Positive integer ≥ 1. |
| `attended` | int | Required. Non-negative integer ≥ 0. Must be ≤ `total_classes`. |

**Only one attendance record per (student_id, course_id) pair.**

**An attendance record may only be created if the student is enrolled in the course.**

### Attendance Percentage (FR-ATT-05)

Attendance percentage is **computed on every read** from `attended` and `total_classes`. **It is never stored in the database.**

```
percentage = (attended / total_classes) × 100
             rounded to 2 decimal places
```

The Go implementation lives in `models/attendance.go`:

```go
func CalculateAttendancePercentage(total, attended int) float64 {
    if total == 0 { return 0.0 }
    pct := (float64(attended) / float64(total)) * 100
    return math.Round(pct*100) / 100
}
```

Examples: `36/48 → 75.00`, `1/3 → 33.33`, `2/3 → 66.67`, `1/1 → 100.00`.

### Low-Attendance Threshold

The threshold is **75%** and is configured via the `ATTENDANCE_THRESHOLD` environment variable (default `75.0` in `config/config.go`).

A student is considered low-attendance when `overall_percentage < 75.0` (strictly less than). A student with exactly 75.00% is **not** low-attendance.

---

## Enrollment Rules

Source: `requirements.md § 5.5`

- A student can be enrolled in many courses; a course can have many students.
- **A student may not be enrolled in the same course twice.** Duplicate enrollment → `ErrDuplicate` → HTTP 409.
- Marks and attendance records may only be created for an enrolled (student, course) pair.
- Deleting a student cascade-deletes their enrollments, marks, and attendance.
- Deleting a course cascade-deletes its enrollments, marks, and attendance.

---

## Computed vs. Stored Values — Critical Rule

| Value | Stored in DB? | Where computed |
|---|---|---|
| Letter grade | **No** | `models.CalculateGrade()` on every read |
| Attendance percentage | **No** | `models.CalculateAttendancePercentage()` on every read |
| Average marks | **No** | Calculated in the query/model when building the student profile |
| Overall attendance | **No** | Calculated in the query/model when building the student profile |

Never add a `grade` or `percentage` column to the database schema. These values are always derived from the raw stored data. This rule exists to eliminate stale cache bugs and is enforced by `architecture.md`.
