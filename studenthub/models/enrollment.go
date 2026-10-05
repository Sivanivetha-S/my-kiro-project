package models

import (
	"database/sql"
	"fmt"
)

// Enrollment represents a student-course assignment.
// All fields map directly to the `enrollments` table.
type Enrollment struct {
	ID         int    `json:"id"`
	StudentID  int    `json:"student_id"`
	CourseID   int    `json:"course_id"`
	EnrolledAt string `json:"enrolled_at"`
}

// CreateEnrollment inserts a new enrollment after verifying that both
// the student and course exist. Returns ErrNotFound if either is missing,
// ErrDuplicate if the (student_id, course_id) pair already exists.
func CreateEnrollment(db *sql.DB, studentID, courseID int) (Enrollment, error) {
	// Verify student exists.
	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM students WHERE id = ?", studentID).Scan(&exists); err != nil {
		return Enrollment{}, fmt.Errorf("CreateEnrollment check student: %w", err)
	}
	if exists == 0 {
		return Enrollment{}, ErrNotFound
	}

	// Verify course exists.
	if err := db.QueryRow("SELECT COUNT(*) FROM courses WHERE id = ?", courseID).Scan(&exists); err != nil {
		return Enrollment{}, fmt.Errorf("CreateEnrollment check course: %w", err)
	}
	if exists == 0 {
		return Enrollment{}, ErrNotFound
	}

	const query = `
		INSERT INTO enrollments (student_id, course_id)
		VALUES (?, ?)
		RETURNING id, student_id, course_id, enrolled_at`

	var e Enrollment
	err := db.QueryRow(query, studentID, courseID).Scan(
		&e.ID, &e.StudentID, &e.CourseID, &e.EnrolledAt,
	)
	if err != nil {
		if isDuplicateError(err) {
			return Enrollment{}, ErrDuplicate
		}
		return Enrollment{}, fmt.Errorf("CreateEnrollment: %w", err)
	}
	return e, nil
}

// GetCoursesForStudent returns all courses a student is enrolled in.
// Returns an empty slice (not nil) when the student has no enrollments.
func GetCoursesForStudent(db *sql.DB, studentID int) ([]Course, error) {
	const query = `
		SELECT c.id, c.course_code, c.course_name, c.credits, c.department
		FROM courses c
		INNER JOIN enrollments e ON e.course_id = c.id
		WHERE e.student_id = ?
		ORDER BY c.course_code`

	rows, err := db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("GetCoursesForStudent: %w", err)
	}
	defer rows.Close()

	courses := []Course{}
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.CourseCode, &c.CourseName, &c.Credits, &c.Department); err != nil {
			return nil, fmt.Errorf("GetCoursesForStudent scan: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

// GetStudentsForCourse returns all students enrolled in a course.
// Returns an empty slice (not nil) when the course has no enrollments.
func GetStudentsForCourse(db *sql.DB, courseID int) ([]Student, error) {
	const query = `
		SELECT s.id, s.student_id, s.full_name, s.email, s.phone,
		       s.department, s.year, s.section, s.dob, s.created_at
		FROM students s
		INNER JOIN enrollments e ON e.student_id = s.id
		WHERE e.course_id = ?
		ORDER BY s.full_name`

	rows, err := db.Query(query, courseID)
	if err != nil {
		return nil, fmt.Errorf("GetStudentsForCourse: %w", err)
	}
	defer rows.Close()

	students := []Student{}
	for rows.Next() {
		var s Student
		var ph sql.NullString
		if err := rows.Scan(
			&s.ID, &s.StudentID, &s.FullName, &s.Email, &ph,
			&s.Department, &s.Year, &s.Section, &s.DOB, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetStudentsForCourse scan: %w", err)
		}
		if ph.Valid {
			s.Phone = ph.String
		}
		students = append(students, s)
	}
	return students, rows.Err()
}

// DeleteEnrollment removes an enrollment by primary key.
// Returns ErrNotFound if no enrollment with that id exists.
func DeleteEnrollment(db *sql.DB, id int) error {
	result, err := db.Exec("DELETE FROM enrollments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnrollment: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// IsEnrolled reports whether the student is enrolled in the given course.
// Used by CreateMarks and CreateAttendance to verify the enrollment constraint.
func IsEnrolled(db *sql.DB, studentID, courseID int) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM enrollments WHERE student_id = ? AND course_id = ?",
		studentID, courseID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("IsEnrolled: %w", err)
	}
	return count > 0, nil
}
