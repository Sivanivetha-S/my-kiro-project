package models

import (
	"database/sql"
	"fmt"
)

// StudentCourseEntry is one enrolled course as returned in the student profile.
type StudentCourseEntry struct {
	EnrollmentID        int      `json:"enrollment_id"`
	CourseID            int      `json:"course_id"`
	CourseCode          string   `json:"course_code"`
	CourseName          string   `json:"course_name"`
	Credits             int      `json:"credits"`
	Marks               *float64 `json:"marks"`                // nil if not recorded
	Grade               *string  `json:"grade"`                // nil if not recorded
	AttendancePercentage *float64 `json:"attendance_percentage"` // nil if not recorded
}

// StudentProfile is the full student response including courses, marks, and attendance.
type StudentProfile struct {
	Student
	Courses           []StudentCourseEntry `json:"courses"`
	AverageMarks      *float64             `json:"average_marks"`      // nil if no marks
	OverallAttendance *float64             `json:"overall_attendance"` // nil if no attendance
}

// GetStudentProfile returns the full profile for a student: their details plus
// all enrolled courses with marks and attendance joined in.
// Returns ErrNotFound if the student does not exist.
func GetStudentProfile(db *sql.DB, id int) (StudentProfile, error) {
	// Fetch the base student record.
	student, err := GetStudentByID(db, id)
	if err != nil {
		return StudentProfile{}, err // propagates ErrNotFound
	}

	// Fetch enrolled courses with marks and attendance via LEFT JOINs.
	const query = `
		SELECT
			e.id          AS enrollment_id,
			c.id          AS course_id,
			c.course_code,
			c.course_name,
			c.credits,
			m.marks       AS marks,
			a.total_classes,
			a.attended
		FROM enrollments e
		INNER JOIN courses c  ON c.id = e.course_id
		LEFT  JOIN marks m    ON m.student_id = e.student_id AND m.course_id = e.course_id
		LEFT  JOIN attendance a ON a.student_id = e.student_id AND a.course_id = e.course_id
		WHERE e.student_id = ?
		ORDER BY c.course_code`

	rows, err := db.Query(query, id)
	if err != nil {
		return StudentProfile{}, fmt.Errorf("GetStudentProfile courses: %w", err)
	}
	defer rows.Close()

	var entries []StudentCourseEntry
	var markTotal float64
	var markCount int
	var totalClasses, totalAttended int

	for rows.Next() {
		var e StudentCourseEntry
		var marksVal sql.NullFloat64
		var tc, att sql.NullInt64

		if err := rows.Scan(
			&e.EnrollmentID, &e.CourseID, &e.CourseCode, &e.CourseName, &e.Credits,
			&marksVal, &tc, &att,
		); err != nil {
			return StudentProfile{}, fmt.Errorf("GetStudentProfile scan: %w", err)
		}

		if marksVal.Valid {
			m := marksVal.Float64
			g := CalculateGrade(m)
			e.Marks = &m
			e.Grade = &g
			markTotal += m
			markCount++
		}
		if tc.Valid && tc.Int64 > 0 {
			pct := CalculateAttendancePercentage(int(tc.Int64), int(att.Int64))
			e.AttendancePercentage = &pct
			totalClasses += int(tc.Int64)
			totalAttended += int(att.Int64)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return StudentProfile{}, fmt.Errorf("GetStudentProfile rows: %w", err)
	}
	if entries == nil {
		entries = []StudentCourseEntry{}
	}

	profile := StudentProfile{
		Student: student,
		Courses: entries,
	}
	if markCount > 0 {
		avg := markTotal / float64(markCount)
		profile.AverageMarks = &avg
	}
	if totalClasses > 0 {
		overall := CalculateAttendancePercentage(totalClasses, totalAttended)
		profile.OverallAttendance = &overall
	}
	return profile, nil
}
