package validators

import "studenthub/models"

// ValidateMarks validates a marks create/update request.
// Returns ALL failing fields at once.
func ValidateMarks(m models.Marks) []ValidationError {
	var errs []ValidationError

	if m.StudentID <= 0 {
		errs = append(errs, ValidationError{"student_id", "Student ID is required."})
	}
	if m.CourseID <= 0 {
		errs = append(errs, ValidationError{"course_id", "Course ID is required."})
	}
	if m.Marks < 0 || m.Marks > 100 {
		errs = append(errs, ValidationError{"marks", "Marks must be between 0 and 100."})
	}

	return errs
}

// ValidateMarksUpdate validates a marks update request (marks value only).
func ValidateMarksUpdate(marks float64) []ValidationError {
	if marks < 0 || marks > 100 {
		return []ValidationError{{"marks", "Marks must be between 0 and 100."}}
	}
	return nil
}
