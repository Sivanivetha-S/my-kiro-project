package models

import (
	"database/sql"
	"fmt"
)

// Marks represents a student's final marks record for a single course.
// Grade is computed on every read from Marks — it is never stored in the database.
// See requirements.md § 5.3 and design.md § 5.5 for the grading scale.
type Marks struct {
	ID         int     `json:"id"`
	StudentID  int     `json:"student_id"`
	CourseID   int     `json:"course_id"`
	Marks      float64 `json:"marks"`
	Grade      string  `json:"grade"`       // computed, not stored
	RecordedAt string  `json:"recorded_at"`
}

// StudentMarksResponse is returned by GetMarksForStudent.
type StudentMarksResponse struct {
	StudentID   int          `json:"student_id"`
	StudentName string       `json:"student_name"`
	Marks       []MarksEntry `json:"marks"`
	AverageMarks *float64    `json:"average_marks"` // nil when no marks recorded
}

// MarksEntry is one row in the marks list for a student.
type MarksEntry struct {
	MarksID    int     `json:"marks_id"`
	CourseID   int     `json:"course_id"`
	CourseCode string  `json:"course_code"`
	CourseName string  `json:"course_name"`
	Marks      float64 `json:"marks"`
	Grade      string  `json:"grade"`
}

// CalculateGrade returns the letter grade for a given marks value.
//
// The grading scale is defined in requirements.md § 5.3:
//
//	90 – 100  →  A+  (Outstanding)
//	80 – 89   →  A   (Excellent)
//	70 – 79   →  B   (Good)
//	60 – 69   →  C   (Satisfactory)
//	50 – 59   →  D   (Pass)
//	 0 – 49   →  F   (Fail)
//
// The caller is responsible for ensuring marks is in [0, 100].
// Inputs outside that range are accepted by this function but produce
// grades consistent with the boundary rules (e.g., −1 → "F", 101 → "A+").
// The validator layer (validators/marks_validator.go) enforces the range
// before any value reaches this function in production.
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

// CreateMarks inserts a new marks record. The caller must have already verified
// the student is enrolled in the course (EnrollmentRequired constraint).
// Returns ErrDuplicate if marks already exist for this (student, course) pair.
func CreateMarks(db *sql.DB, m Marks) (Marks, error) {
	const query = `
		INSERT INTO marks (student_id, course_id, marks)
		VALUES (?, ?, ?)
		RETURNING id, student_id, course_id, marks, recorded_at`

	var created Marks
	err := db.QueryRow(query, m.StudentID, m.CourseID, m.Marks).Scan(
		&created.ID, &created.StudentID, &created.CourseID,
		&created.Marks, &created.RecordedAt,
	)
	if err != nil {
		if isDuplicateError(err) {
			return Marks{}, ErrDuplicate
		}
		return Marks{}, fmt.Errorf("CreateMarks: %w", err)
	}
	created.Grade = CalculateGrade(created.Marks)
	return created, nil
}

// GetMarksForStudent returns all marks for a student plus their average.
func GetMarksForStudent(db *sql.DB, studentID int) (StudentMarksResponse, error) {
	// Get student name.
	var studentName string
	err := db.QueryRow("SELECT full_name FROM students WHERE id = ?", studentID).Scan(&studentName)
	if err == sql.ErrNoRows {
		return StudentMarksResponse{}, ErrNotFound
	}
	if err != nil {
		return StudentMarksResponse{}, fmt.Errorf("GetMarksForStudent: %w", err)
	}

	const query = `
		SELECT m.id, m.course_id, c.course_code, c.course_name, m.marks
		FROM marks m
		INNER JOIN courses c ON c.id = m.course_id
		WHERE m.student_id = ?
		ORDER BY c.course_code`

	rows, err := db.Query(query, studentID)
	if err != nil {
		return StudentMarksResponse{}, fmt.Errorf("GetMarksForStudent query: %w", err)
	}
	defer rows.Close()

	var entries []MarksEntry
	var total float64
	for rows.Next() {
		var e MarksEntry
		if err := rows.Scan(&e.MarksID, &e.CourseID, &e.CourseCode, &e.CourseName, &e.Marks); err != nil {
			return StudentMarksResponse{}, fmt.Errorf("GetMarksForStudent scan: %w", err)
		}
		e.Grade = CalculateGrade(e.Marks)
		entries = append(entries, e)
		total += e.Marks
	}
	if err := rows.Err(); err != nil {
		return StudentMarksResponse{}, fmt.Errorf("GetMarksForStudent rows: %w", err)
	}

	resp := StudentMarksResponse{
		StudentID:   studentID,
		StudentName: studentName,
		Marks:       entries,
	}
	if len(entries) > 0 {
		avg := total / float64(len(entries))
		resp.AverageMarks = &avg
	}
	return resp, nil
}

// UpdateMarks replaces the marks value for an existing record and returns
// the updated record with a freshly computed grade.
// Returns ErrNotFound if no marks record with that id exists.
func UpdateMarks(db *sql.DB, id int, newMarks float64) (Marks, error) {
	result, err := db.Exec("UPDATE marks SET marks = ? WHERE id = ?", newMarks, id)
	if err != nil {
		return Marks{}, fmt.Errorf("UpdateMarks: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Marks{}, ErrNotFound
	}

	const query = `SELECT id, student_id, course_id, marks, recorded_at FROM marks WHERE id = ?`
	var m Marks
	if err := db.QueryRow(query, id).Scan(
		&m.ID, &m.StudentID, &m.CourseID, &m.Marks, &m.RecordedAt,
	); err != nil {
		return Marks{}, fmt.Errorf("UpdateMarks fetch: %w", err)
	}
	m.Grade = CalculateGrade(m.Marks)
	return m, nil
}

// GetMarksByStudentAndCourse returns a marks record for a specific (student, course) pair.
// Returns ErrNotFound if no record exists.
func GetMarksByStudentAndCourse(db *sql.DB, studentID, courseID int) (Marks, error) {
	const query = `SELECT id, student_id, course_id, marks, recorded_at FROM marks WHERE student_id = ? AND course_id = ?`
	var m Marks
	err := db.QueryRow(query, studentID, courseID).Scan(
		&m.ID, &m.StudentID, &m.CourseID, &m.Marks, &m.RecordedAt,
	)
	if err == sql.ErrNoRows {
		return Marks{}, ErrNotFound
	}
	if err != nil {
		return Marks{}, fmt.Errorf("GetMarksByStudentAndCourse: %w", err)
	}
	m.Grade = CalculateGrade(m.Marks)
	return m, nil
}
