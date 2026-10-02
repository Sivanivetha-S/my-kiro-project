package models

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
