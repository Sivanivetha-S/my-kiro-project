package models

import (
	"database/sql"
	"fmt"
	"math"
)

// Attendance represents a student's attendance record for a single course.
// Percentage is computed on every read from TotalClasses and Attended —
// it is never stored in the database.
// See requirements.md § 5.7 and design.md § 5.6.
type Attendance struct {
	ID           int     `json:"id"`
	StudentID    int     `json:"student_id"`
	CourseID     int     `json:"course_id"`
	TotalClasses int     `json:"total_classes"`
	Attended     int     `json:"attended"`
	Percentage   float64 `json:"percentage"` // computed, not stored
	RecordedAt   string  `json:"recorded_at"`
}

// StudentAttendanceResponse is returned by GetAttendanceForStudent.
type StudentAttendanceResponse struct {
	StudentID         int               `json:"student_id"`
	StudentName       string            `json:"student_name"`
	Attendance        []AttendanceEntry `json:"attendance"`
	OverallPercentage *float64          `json:"overall_percentage"` // nil when no records
}

// AttendanceEntry is one row in the attendance list for a student.
type AttendanceEntry struct {
	AttendanceID int     `json:"attendance_id"`
	CourseID     int     `json:"course_id"`
	CourseCode   string  `json:"course_code"`
	CourseName   string  `json:"course_name"`
	TotalClasses int     `json:"total_classes"`
	Attended     int     `json:"attended"`
	Percentage   float64 `json:"percentage"`
}

// LowAttendanceStudent is used in the low-attendance list and dashboard.
type LowAttendanceStudent struct {
	StudentID         string  `json:"student_id"`
	FullName          string  `json:"full_name"`
	Department        string  `json:"department"`
	Year              int     `json:"year"`
	OverallAttendance float64 `json:"overall_attendance"`
}

// CalculateAttendancePercentage returns the attendance percentage rounded to
// two decimal places.
//
// The formula is defined in requirements.md § 5.7 (FR-ATT-05):
//
//	percentage = (attended / total_classes) * 100
//
// Special case: if total is 0, returns 0.0 to avoid a divide-by-zero panic.
// In production, total_classes ≥ 1 is enforced by the validator layer
// (validators/attendance_validator.go), so this branch is a safety net only.
//
// Examples:
//
//	CalculateAttendancePercentage(48, 36) → 75.00
//	CalculateAttendancePercentage(3,  1) → 33.33
//	CalculateAttendancePercentage(1,  1) → 100.00
//	CalculateAttendancePercentage(0,  0) → 0.00   (safety net)
func CalculateAttendancePercentage(total, attended int) float64 {
	if total == 0 {
		return 0.0
	}
	pct := (float64(attended) / float64(total)) * 100
	return math.Round(pct*100) / 100
}

// CreateAttendance inserts a new attendance record.
// The caller must have already verified the student is enrolled in the course.
// Returns ErrDuplicate if a record already exists for (student_id, course_id).
func CreateAttendance(db *sql.DB, a Attendance) (Attendance, error) {
	const query = `
		INSERT INTO attendance (student_id, course_id, total_classes, attended)
		VALUES (?, ?, ?, ?)
		RETURNING id, student_id, course_id, total_classes, attended, recorded_at`

	var created Attendance
	err := db.QueryRow(query, a.StudentID, a.CourseID, a.TotalClasses, a.Attended).Scan(
		&created.ID, &created.StudentID, &created.CourseID,
		&created.TotalClasses, &created.Attended, &created.RecordedAt,
	)
	if err != nil {
		if isDuplicateError(err) {
			return Attendance{}, ErrDuplicate
		}
		return Attendance{}, fmt.Errorf("CreateAttendance: %w", err)
	}
	created.Percentage = CalculateAttendancePercentage(created.TotalClasses, created.Attended)
	return created, nil
}

// GetAttendanceForStudent returns all attendance records for a student
// along with the overall percentage across all courses.
func GetAttendanceForStudent(db *sql.DB, studentID int) (StudentAttendanceResponse, error) {
	var studentName string
	err := db.QueryRow("SELECT full_name FROM students WHERE id = ?", studentID).Scan(&studentName)
	if err == sql.ErrNoRows {
		return StudentAttendanceResponse{}, ErrNotFound
	}
	if err != nil {
		return StudentAttendanceResponse{}, fmt.Errorf("GetAttendanceForStudent: %w", err)
	}

	const query = `
		SELECT a.id, a.course_id, c.course_code, c.course_name, a.total_classes, a.attended
		FROM attendance a
		INNER JOIN courses c ON c.id = a.course_id
		WHERE a.student_id = ?
		ORDER BY c.course_code`

	rows, err := db.Query(query, studentID)
	if err != nil {
		return StudentAttendanceResponse{}, fmt.Errorf("GetAttendanceForStudent query: %w", err)
	}
	defer rows.Close()

	var entries []AttendanceEntry
	var totalAttended, totalClasses int
	for rows.Next() {
		var e AttendanceEntry
		if err := rows.Scan(
			&e.AttendanceID, &e.CourseID, &e.CourseCode, &e.CourseName,
			&e.TotalClasses, &e.Attended,
		); err != nil {
			return StudentAttendanceResponse{}, fmt.Errorf("GetAttendanceForStudent scan: %w", err)
		}
		e.Percentage = CalculateAttendancePercentage(e.TotalClasses, e.Attended)
		entries = append(entries, e)
		totalAttended += e.Attended
		totalClasses += e.TotalClasses
	}
	if err := rows.Err(); err != nil {
		return StudentAttendanceResponse{}, fmt.Errorf("GetAttendanceForStudent rows: %w", err)
	}

	resp := StudentAttendanceResponse{
		StudentID:   studentID,
		StudentName: studentName,
		Attendance:  entries,
	}
	if totalClasses > 0 {
		overall := CalculateAttendancePercentage(totalClasses, totalAttended)
		resp.OverallPercentage = &overall
	}
	return resp, nil
}

// UpdateAttendance replaces total_classes and attended for an existing record
// and returns the updated record with a freshly computed percentage.
// Returns ErrNotFound if no attendance record with that id exists.
func UpdateAttendance(db *sql.DB, id, totalClasses, attended int) (Attendance, error) {
	result, err := db.Exec(
		"UPDATE attendance SET total_classes = ?, attended = ? WHERE id = ?",
		totalClasses, attended, id,
	)
	if err != nil {
		return Attendance{}, fmt.Errorf("UpdateAttendance: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Attendance{}, ErrNotFound
	}

	const query = `SELECT id, student_id, course_id, total_classes, attended, recorded_at FROM attendance WHERE id = ?`
	var a Attendance
	if err := db.QueryRow(query, id).Scan(
		&a.ID, &a.StudentID, &a.CourseID, &a.TotalClasses, &a.Attended, &a.RecordedAt,
	); err != nil {
		return Attendance{}, fmt.Errorf("UpdateAttendance fetch: %w", err)
	}
	a.Percentage = CalculateAttendancePercentage(a.TotalClasses, a.Attended)
	return a, nil
}

// GetLowAttendanceStudents returns all students whose overall attendance
// percentage is strictly below the given threshold.
// A student with exactly the threshold percentage is NOT included (strict less-than).
func GetLowAttendanceStudents(db *sql.DB, threshold float64) ([]LowAttendanceStudent, error) {
	// Aggregate total and attended per student, then compute percentage in Go
	// to use the same rounding logic as CalculateAttendancePercentage.
	const query = `
		SELECT s.student_id, s.full_name, s.department, s.year,
		       SUM(a.total_classes) AS total, SUM(a.attended) AS attended
		FROM attendance a
		INNER JOIN students s ON s.id = a.student_id
		GROUP BY a.student_id
		ORDER BY s.full_name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("GetLowAttendanceStudents: %w", err)
	}
	defer rows.Close()

	var result []LowAttendanceStudent
	for rows.Next() {
		var sid, name, dept string
		var year, total, attended int
		if err := rows.Scan(&sid, &name, &dept, &year, &total, &attended); err != nil {
			return nil, fmt.Errorf("GetLowAttendanceStudents scan: %w", err)
		}
		pct := CalculateAttendancePercentage(total, attended)
		if pct < threshold {
			result = append(result, LowAttendanceStudent{
				StudentID:         sid,
				FullName:          name,
				Department:        dept,
				Year:              year,
				OverallAttendance: pct,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetLowAttendanceStudents rows: %w", err)
	}
	if result == nil {
		result = []LowAttendanceStudent{}
	}
	return result, nil
}

// GetAttendanceByStudentAndCourse returns a single attendance record for a
// (student_id, course_id) pair. Returns ErrNotFound if none exists.
func GetAttendanceByStudentAndCourse(db *sql.DB, studentID, courseID int) (Attendance, error) {
	const query = `
		SELECT id, student_id, course_id, total_classes, attended, recorded_at
		FROM attendance WHERE student_id = ? AND course_id = ?`
	var a Attendance
	err := db.QueryRow(query, studentID, courseID).Scan(
		&a.ID, &a.StudentID, &a.CourseID, &a.TotalClasses, &a.Attended, &a.RecordedAt,
	)
	if err == sql.ErrNoRows {
		return Attendance{}, ErrNotFound
	}
	if err != nil {
		return Attendance{}, fmt.Errorf("GetAttendanceByStudentAndCourse: %w", err)
	}
	a.Percentage = CalculateAttendancePercentage(a.TotalClasses, a.Attended)
	return a, nil
}
