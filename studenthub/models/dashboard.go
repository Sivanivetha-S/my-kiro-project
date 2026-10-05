package models

import (
	"database/sql"
	"fmt"
)

// DashboardResponse is the full payload returned by GET /api/v1/dashboard.
type DashboardResponse struct {
	TotalStudents          int                    `json:"total_students"`
	TotalCourses           int                    `json:"total_courses"`
	AverageAttendance      float64                `json:"average_attendance"`
	LowAttendanceStudents  []LowAttendanceStudent `json:"low_attendance_students"`
	RecentStudents         []RecentStudent        `json:"recent_students"`
}

// RecentStudent is a compact student record used in the dashboard.
type RecentStudent struct {
	ID         int    `json:"id"`
	StudentID  string `json:"student_id"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
	Year       int    `json:"year"`
	CreatedAt  string `json:"created_at"`
}

// GetDashboard fetches all data needed for the dashboard endpoint in a single
// coordinated set of queries.
func GetDashboard(db *sql.DB, threshold float64) (DashboardResponse, error) {
	var resp DashboardResponse

	// 1. Total students.
	if err := db.QueryRow("SELECT COUNT(*) FROM students").Scan(&resp.TotalStudents); err != nil {
		return DashboardResponse{}, fmt.Errorf("GetDashboard total students: %w", err)
	}

	// 2. Total courses.
	if err := db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&resp.TotalCourses); err != nil {
		return DashboardResponse{}, fmt.Errorf("GetDashboard total courses: %w", err)
	}

	// 3. Average attendance across all records.
	// Compute by aggregating total/attended so we use the same rounding logic.
	var totalSum, attendedSum sql.NullInt64
	if err := db.QueryRow(
		"SELECT SUM(total_classes), SUM(attended) FROM attendance",
	).Scan(&totalSum, &attendedSum); err != nil {
		return DashboardResponse{}, fmt.Errorf("GetDashboard attendance sums: %w", err)
	}
	if totalSum.Valid && totalSum.Int64 > 0 {
		resp.AverageAttendance = CalculateAttendancePercentage(int(totalSum.Int64), int(attendedSum.Int64))
	}

	// 4. Low attendance students.
	low, err := GetLowAttendanceStudents(db, threshold)
	if err != nil {
		return DashboardResponse{}, fmt.Errorf("GetDashboard low attendance: %w", err)
	}
	resp.LowAttendanceStudents = low

	// 5. 5 most recently added students.
	const recentQuery = `
		SELECT id, student_id, full_name, department, year, created_at
		FROM students
		ORDER BY created_at DESC
		LIMIT 5`

	rows, err := db.Query(recentQuery)
	if err != nil {
		return DashboardResponse{}, fmt.Errorf("GetDashboard recent students: %w", err)
	}
	defer rows.Close()

	var recent []RecentStudent
	for rows.Next() {
		var s RecentStudent
		if err := rows.Scan(&s.ID, &s.StudentID, &s.FullName, &s.Department, &s.Year, &s.CreatedAt); err != nil {
			return DashboardResponse{}, fmt.Errorf("GetDashboard recent scan: %w", err)
		}
		recent = append(recent, s)
	}
	if err := rows.Err(); err != nil {
		return DashboardResponse{}, fmt.Errorf("GetDashboard recent rows: %w", err)
	}
	if recent == nil {
		recent = []RecentStudent{}
	}
	resp.RecentStudents = recent

	return resp, nil
}
