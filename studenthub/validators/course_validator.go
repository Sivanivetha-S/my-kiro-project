package validators

import (
	"regexp"
	"strings"

	"studenthub/models"
)

// courseCodeRe: 2–20 uppercase letters, digits, or hyphens.
var courseCodeRe = regexp.MustCompile(`^[A-Z0-9\-]{2,20}$`)

// ValidateCourse validates all fields of a course create/update request.
// Returns ALL failing fields at once.
func ValidateCourse(c models.Course) []ValidationError {
	var errs []ValidationError

	// course_code
	if strings.TrimSpace(c.CourseCode) == "" {
		errs = append(errs, ValidationError{"course_code", "Course code is required."})
	} else if !courseCodeRe.MatchString(c.CourseCode) {
		errs = append(errs, ValidationError{"course_code", "Course code must be 2–20 uppercase letters, digits, or hyphens."})
	}

	// course_name
	name := strings.TrimSpace(c.CourseName)
	if name == "" {
		errs = append(errs, ValidationError{"course_name", "Course name is required."})
	} else if len([]rune(name)) < 3 || len([]rune(name)) > 150 {
		errs = append(errs, ValidationError{"course_name", "Course name must be 3–150 characters."})
	}

	// credits
	if c.Credits < 1 || c.Credits > 6 {
		errs = append(errs, ValidationError{"credits", "Credits must be an integer between 1 and 6."})
	}

	// department
	if strings.TrimSpace(c.Department) == "" {
		errs = append(errs, ValidationError{"department", "Department is required."})
	} else if !isValidDepartment(c.Department) {
		errs = append(errs, ValidationError{"department", "Department must be one of the defined department list."})
	}

	return errs
}
