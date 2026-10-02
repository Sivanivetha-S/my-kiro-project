package models

import "math"

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
	Percentage   float64 `json:"percentage"`  // computed, not stored
	RecordedAt   string  `json:"recorded_at"`
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
