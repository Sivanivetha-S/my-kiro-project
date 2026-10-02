# StudentHub — Requirements Specification

**Project:** StudentHub — Student Management System  
**Version:** 1.0  
**Context:** Kiro University — Spec-Driven Development Demonstration  
**Date:** 2026-10-02  
**Status:** Draft

---

## 1. Project Overview

StudentHub is a full-stack web application for managing college student academic information. It provides administrators and faculty with tools to manage students, courses, marks, and attendance through a clean, responsive interface backed by a REST API.

The application is intentionally scoped for simplicity and maintainability, making it suitable as a teaching project while demonstrating real-world software development practices including spec-driven development, modular architecture, and proper validation.

---

## 2. Stakeholders and Users

| Role | Description |
|---|---|
| Administrator | Primary user. Manages all records — students, courses, marks, attendance. |
| Faculty (future) | Scoped out of v1. May view and record marks/attendance in a future version. |

> **v1 scope:** No authentication or role-based access control. The application is treated as a single-user internal tool. Authentication is explicitly deferred to a future version.

---

## 3. Ambiguities Resolved

The following ambiguities were identified during requirements analysis and resolved with explicit decisions documented here.

| Ambiguity | Decision |
|---|---|
| Is Student ID system-generated or user-defined? | User-defined (e.g., `STU2024001`). Must be unique. |
| Is Course ID system-generated or user-defined? | System-generated integer primary key. Course Code (e.g., `CS101`) is the human-readable identifier and must be unique. |
| Are marks per-assessment or a single final mark? | A single final marks record per student per course (marks out of 100). Multiple assessments are out of scope for v1. |
| What is the attendance threshold for "low attendance"? | 75%. Configurable as a server-side constant (`ATTENDANCE_THRESHOLD = 75`). |
| What grading scale is used? | Percentage-based. See Section 5.3. |
| What does "Section" mean for a student? | A free-text field representing the class section (e.g., `A`, `B`, `CS-A`). No referential constraint. |
| Are departments a fixed list or free text? | Fixed list defined at the application level. Prevents inconsistent data entry. See Section 5.1. |
| Can a student be enrolled in multiple courses? | Yes. A student can be assigned to many courses. A course can have many students. |
| Can marks be updated after entry? | Yes. Marks records are updatable via a PUT endpoint. |
| Can attendance be updated after entry? | Yes. Attendance records are updatable via a PUT endpoint. |

---

## 4. User Stories

### 4.1 Student Management

**US-01** — As an administrator, I want to add a new student with all their academic details, so that their record exists in the system.

**US-02** — As an administrator, I want to view a paginated list of all students, so that I can get an overview of the student body.

**US-03** — As an administrator, I want to view the full profile of a single student including their courses, marks, and attendance, so that I can review their complete academic record.

**US-04** — As an administrator, I want to edit an existing student's information, so that I can correct mistakes or update their details.

**US-05** — As an administrator, I want to delete a student, so that I can remove records that are no longer needed.

**US-06** — As an administrator, I want to search students by name, email, or Student ID, so that I can quickly find a specific student.

**US-07** — As an administrator, I want to filter students by department and/or year, so that I can narrow down a list to a specific cohort.

### 4.2 Course Management

**US-08** — As an administrator, I want to add a new course with its details, so that it is available to be assigned to students.

**US-09** — As an administrator, I want to view all courses, so that I can see what courses the institution offers.

**US-10** — As an administrator, I want to edit a course's information, so that I can keep it accurate.

**US-11** — As an administrator, I want to delete a course, so that outdated courses are removed.

**US-12** — As an administrator, I want to assign one or more courses to a student, so that their enrollment is recorded.

**US-13** — As an administrator, I want to remove a course assignment from a student, so that enrollment errors can be corrected.

### 4.3 Marks Management

**US-14** — As an administrator, I want to record a student's final marks for a course, so that their academic performance is captured.

**US-15** — As an administrator, I want to view all marks for a student across their enrolled courses, so that I can review their performance.

**US-16** — As an administrator, I want the system to automatically calculate the total marks, average, and letter grade, so that I do not have to compute these manually.

**US-17** — As an administrator, I want to update a marks record, so that data-entry errors can be corrected.

### 4.4 Attendance Management

**US-18** — As an administrator, I want to record the total classes held and classes attended for a student in a course, so that attendance is tracked.

**US-19** — As an administrator, I want the system to automatically calculate the attendance percentage, so that I can see it at a glance.

**US-20** — As an administrator, I want to update an attendance record, so that it stays current.

**US-21** — As an administrator, I want to see a list of students whose attendance is below 75%, so that I can identify at-risk students.

### 4.5 Dashboard

**US-22** — As an administrator, I want a dashboard that shows total students, total courses, average attendance, and a list of low-attendance students, so that I have a quick operational overview.

**US-23** — As an administrator, I want the dashboard to show the most recently added student records, so that I can see recent activity.

---

## 5. Functional Requirements

### 5.1 Reference Data — Departments

The following departments are supported in v1. The list is defined as a constant in both the backend and frontend.

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

### 5.2 Student Management

| ID | Requirement |
|---|---|
| FR-STU-01 | The system shall allow creation of a student record with all required fields. |
| FR-STU-02 | Student ID shall be unique across all student records. |
| FR-STU-03 | Email shall be unique across all student records. |
| FR-STU-04 | The system shall return a list of all students, sortable and filterable. |
| FR-STU-05 | The system shall support searching students by name (partial match), email (partial match), or exact Student ID. |
| FR-STU-06 | The system shall support filtering students by department and/or year. |
| FR-STU-07 | The system shall return a single student's full profile including enrolled courses, marks, and attendance summary. |
| FR-STU-08 | The system shall allow updating any field of a student record except the system-internal ID. |
| FR-STU-09 | The system shall allow deletion of a student record. Deleting a student shall cascade-delete their marks and attendance records. |
| FR-STU-10 | Year shall be an integer between 1 and 6 inclusive. |

### 5.3 Grading Scale

Marks are recorded out of 100. Grades are computed server-side as follows:

| Marks Range | Grade | Description |
|---|---|---|
| 90 – 100 | A+ | Outstanding |
| 80 – 89 | A | Excellent |
| 70 – 79 | B | Good |
| 60 – 69 | C | Satisfactory |
| 50 – 59 | D | Pass |
| 0 – 49 | F | Fail |

> Grades are never stored in the database. They are computed on every read from the marks value.

### 5.4 Course Management

| ID | Requirement |
|---|---|
| FR-CRS-01 | The system shall allow creation of a course with all required fields. |
| FR-CRS-02 | Course Code shall be unique across all course records. |
| FR-CRS-03 | The system shall return a list of all courses. |
| FR-CRS-04 | The system shall allow updating any field of a course record. |
| FR-CRS-05 | The system shall allow deletion of a course. Deleting a course shall cascade-delete its enrollment, marks, and attendance records. |
| FR-CRS-06 | Credits shall be an integer between 1 and 6 inclusive. |

### 5.5 Enrollment (Student–Course Assignment)

| ID | Requirement |
|---|---|
| FR-ENR-01 | The system shall allow assigning a course to a student (creating an enrollment). |
| FR-ENR-02 | A student shall not be enrolled in the same course more than once. |
| FR-ENR-03 | The system shall allow removing an enrollment. |
| FR-ENR-04 | The system shall return a list of courses for a given student. |
| FR-ENR-05 | The system shall return a list of students enrolled in a given course. |

### 5.6 Marks Management

| ID | Requirement |
|---|---|
| FR-MRK-01 | The system shall allow recording a marks record for a (student, course) pair. |
| FR-MRK-02 | Only one marks record may exist per (student, course) pair. |
| FR-MRK-03 | Marks shall be a numeric value between 0 and 100 inclusive. |
| FR-MRK-04 | The system shall compute and return the letter grade on every read. |
| FR-MRK-05 | The system shall compute and return the average marks across all courses for a student. |
| FR-MRK-06 | The system shall allow updating a marks record. |
| FR-MRK-07 | A marks record may only be created if the student is enrolled in the course. |

### 5.7 Attendance Management

| ID | Requirement |
|---|---|
| FR-ATT-01 | The system shall allow recording an attendance record for a (student, course) pair. |
| FR-ATT-02 | Only one attendance record may exist per (student, course) pair. |
| FR-ATT-03 | Total classes held shall be a positive integer. |
| FR-ATT-04 | Classes attended shall be a non-negative integer and shall not exceed total classes held. |
| FR-ATT-05 | The system shall compute attendance percentage as `(attended / total) * 100`, rounded to two decimal places. |
| FR-ATT-06 | The system shall return a list of students with attendance below 75% (the configured threshold). |
| FR-ATT-07 | The system shall allow updating an attendance record. |
| FR-ATT-08 | An attendance record may only be created if the student is enrolled in the course. |

### 5.8 Dashboard

| ID | Requirement |
|---|---|
| FR-DSH-01 | The dashboard endpoint shall return total student count. |
| FR-DSH-02 | The dashboard endpoint shall return total course count. |
| FR-DSH-03 | The dashboard endpoint shall return the overall average attendance percentage across all students and courses. |
| FR-DSH-04 | The dashboard endpoint shall return a list of students with attendance below the threshold. |
| FR-DSH-05 | The dashboard endpoint shall return the 5 most recently added student records (by creation timestamp). |

---

## 6. Non-Functional Requirements

| ID | Category | Requirement |
|---|---|---|
| NFR-01 | Responsiveness | The UI shall be usable on screens from 320px (mobile) to 1920px (desktop) without horizontal scrolling. |
| NFR-02 | Accessibility | All form inputs shall have associated `<label>` elements. All interactive controls shall be keyboard-navigable. Color contrast shall meet WCAG 2.1 AA minimum ratios. |
| NFR-03 | Validation | All user inputs shall be validated on both the client side (before submission) and the server side (before database write). |
| NFR-04 | Error Handling | The API shall return structured JSON error responses. The UI shall display user-friendly error messages for all known error states. |
| NFR-05 | Security | No credentials, API keys, or secrets shall be hardcoded in source code. Environment variables or a config file (excluded from version control) shall be used for any runtime configuration. |
| NFR-06 | Performance | API responses for list endpoints shall return within 500ms for datasets up to 10,000 records on a standard development machine. |
| NFR-07 | Maintainability | Backend code shall be organized into packages: `main`, `handlers`, `models`, `db`, `validators`. Frontend code shall be organized into modules by feature. |
| NFR-08 | Portability | The application shall run on any platform where Go and SQLite are available without OS-specific dependencies. |
| NFR-09 | No Hardcoded Data | No student, course, marks, or attendance data shall be hardcoded in the production application. Seed data for development is acceptable in a clearly named `seed.sql` file. |
| NFR-10 | Testing | The application shall include unit tests for all validation functions and grade/attendance calculation logic. |

---

## 7. Validation Rules

### 7.1 Student Fields

| Field | Rules |
|---|---|
| Student ID | Required. 3–20 characters. Alphanumeric and hyphens only. Must be unique. |
| Full Name | Required. 2–100 characters. Letters, spaces, hyphens, and apostrophes only. |
| Email | Required. Valid email format (RFC 5322 simplified). Must be unique. Max 255 characters. |
| Phone | Optional. If provided, must match pattern `+?[0-9\s\-()]{7,20}`. |
| Department | Required. Must be one of the defined department list. |
| Year | Required. Integer 1–6. |
| Section | Required. 1–10 characters. Alphanumeric only. |
| Date of Birth | Required. Valid date in `YYYY-MM-DD` format. Student must be at least 15 years old. |

### 7.2 Course Fields

| Field | Rules |
|---|---|
| Course Code | Required. 2–20 characters. Uppercase letters, digits, and hyphens only. Must be unique. |
| Course Name | Required. 3–150 characters. |
| Credits | Required. Integer 1–6. |
| Department | Required. Must be one of the defined department list. |

### 7.3 Marks Fields

| Field | Rules |
|---|---|
| Student ID | Required. Must reference an existing student. |
| Course ID | Required. Must reference an existing course. |
| Marks | Required. Numeric (decimal allowed). Range: 0.00 – 100.00. |

### 7.4 Attendance Fields

| Field | Rules |
|---|---|
| Student ID | Required. Must reference an existing student. |
| Course ID | Required. Must reference an existing course. |
| Total Classes | Required. Positive integer (≥ 1). |
| Attended Classes | Required. Non-negative integer (≥ 0). Must be ≤ Total Classes. |

---

## 8. Error Handling Requirements

| ID | Requirement |
|---|---|
| EH-01 | All API errors shall return a JSON body in the format `{"error": "<message>", "code": "<error_code>"}`. |
| EH-02 | Validation failures shall return HTTP 400 with a list of all failing field messages. |
| EH-03 | Resource not found shall return HTTP 404 with a descriptive message. |
| EH-04 | Duplicate unique-field violations shall return HTTP 409 with the conflicting field identified. |
| EH-05 | Internal server errors shall return HTTP 500 with a generic message. The full error shall be logged server-side only (never exposed to the client). |
| EH-06 | The frontend shall display inline field-level error messages for validation failures. |
| EH-07 | The frontend shall display a dismissible global notification for server errors (network failure, 500). |
| EH-08 | The frontend shall never show raw stack traces or Go error strings to the user. |
| EH-09 | Database constraint violations (duplicate key, foreign key) shall be caught and translated to user-facing messages before returning a response. |

---

## 9. Acceptance Criteria

### AC-STU (Student Management)

- [ ] A student can be created with all valid fields and appears in the student list.
- [ ] Creating a student with a duplicate Student ID returns a 409 error with a clear message.
- [ ] Creating a student with a duplicate email returns a 409 error.
- [ ] Creating a student with an invalid date of birth (future date, or age < 15) returns a 400 error.
- [ ] The student list can be filtered by department and year simultaneously.
- [ ] Searching by partial name returns all matching students.
- [ ] Deleting a student removes their marks and attendance records.

### AC-CRS (Course Management)

- [ ] A course can be created and appears in the course list.
- [ ] Creating a course with a duplicate Course Code returns a 409 error.
- [ ] Deleting a course removes all associated enrollment, marks, and attendance records.

### AC-ENR (Enrollment)

- [ ] A student can be assigned to a course and the enrollment appears in the student's profile.
- [ ] Assigning the same student to the same course twice returns a 409 error.

### AC-MRK (Marks)

- [ ] Marks of 95 → Grade A+, 82 → A, 73 → B, 65 → C, 55 → D, 40 → F.
- [ ] Marks outside 0–100 are rejected with a 400 error.
- [ ] Recording marks for a non-enrolled (student, course) pair is rejected with a 400 error.
- [ ] The average marks for a student is correctly computed across all their courses.

### AC-ATT (Attendance)

- [ ] Attendance percentage is correctly computed: 36/48 → 75.00%.
- [ ] Setting attended > total is rejected with a 400 error.
- [ ] Students below 75% attendance appear in the low-attendance list.

### AC-DSH (Dashboard)

- [ ] Dashboard counts reflect the actual number of students and courses in the database.
- [ ] The 5 most recent students are shown in insertion order.
- [ ] A student with 60% attendance appears in the low-attendance widget.

### AC-UI (User Interface)

- [ ] All forms display inline validation errors before submission.
- [ ] The UI is usable on a 375px-wide mobile screen without horizontal overflow.
- [ ] All form labels are associated with their inputs (verified by checking `for`/`id` attributes).
- [ ] The application works with JavaScript-disabled state gracefully degraded (forms submit to API directly).

---

## 10. Testing Requirements

| ID | Type | Description |
|---|---|---|
| TR-01 | Unit — Backend | Test the `CalculateGrade(marks float64) string` function for all grade boundary values. |
| TR-02 | Unit — Backend | Test the `CalculateAttendancePercentage(total, attended int) float64` function including edge cases (total = 0). |
| TR-03 | Unit — Backend | Test all validation functions in the `validators` package for valid and invalid inputs. |
| TR-04 | Unit — Backend | Test request-parsing helpers to ensure malformed JSON returns appropriate errors. |
| TR-05 | Integration — Backend | Test each REST endpoint using Go's `net/http/httptest` package for happy-path and error-path scenarios. |
| TR-06 | Unit — Frontend | Test client-side validation functions (e.g., `validateEmail`, `validateStudentId`) using a lightweight test runner. |
| TR-07 | Property-Based — Backend | Use property-based testing to verify that `CalculateGrade` always returns a valid grade for any float in [0, 100] and never panics. (Demonstrates Kiro property-based testing capability.) |
| TR-08 | Manual | Verify responsive layout on Chrome DevTools at 375px, 768px, and 1280px breakpoints. |
| TR-09 | Manual | Verify keyboard navigation through all forms using Tab and Enter keys. |

---

## 11. Out of Scope (v1)

The following features are explicitly out of scope for v1 and should not be implemented:

- User authentication and authorization
- Role-based access control (admin vs. faculty)
- Multiple assessments per course (only final marks)
- File upload (e.g., student photos)
- Email notifications
- Report generation / PDF export
- Audit logging
- Pagination on the backend (frontend handles display limiting for v1)
- Soft delete (deletes are hard deletes in v1)
