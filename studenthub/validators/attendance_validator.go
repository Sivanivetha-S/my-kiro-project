package validators

import "studenthub/models"

// ValidateAttendance validates an attendance create request.
// Returns ALL failing fields at once.
func ValidateAttendance(a models.Attendance) []ValidationError {
	var errs []ValidationError

	if a.StudentID <= 0 {
		errs = append(errs, ValidationError{"student_id", "Student ID is required."})
	}
	if a.CourseID <= 0 {
		errs = append(errs, ValidationError{"course_id", "Course ID is required."})
	}
	if a.TotalClasses < 1 {
		errs = append(errs, ValidationError{"total_classes", "Total classes must be at least 1."})
	}
	if a.Attended < 0 {
		errs = append(errs, ValidationError{"attended", "Attended classes cannot be negative."})
	}
	// Only check the attended <= total constraint when total is valid.
	if a.TotalClasses >= 1 && a.Attended > a.TotalClasses {
		errs = append(errs, ValidationError{"attended", "Attended classes cannot exceed total classes."})
	}

	return errs
}

// ValidateAttendanceUpdate validates an attendance update request.
func ValidateAttendanceUpdate(totalClasses, attended int) []ValidationError {
	var errs []ValidationError
	if totalClasses < 1 {
		errs = append(errs, ValidationError{"total_classes", "Total classes must be at least 1."})
	}
	if attended < 0 {
		errs = append(errs, ValidationError{"attended", "Attended classes cannot be negative."})
	}
	if totalClasses >= 1 && attended > totalClasses {
		errs = append(errs, ValidationError{"attended", "Attended classes cannot exceed total classes."})
	}
	return errs
}
