package validators

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"studenthub/models"
)

// studentIDRe allows 3–20 alphanumeric characters plus hyphens.
var studentIDRe = regexp.MustCompile(`^[A-Za-z0-9\-]{3,20}$`)

// phoneRe: optional leading '+', then digits/spaces/hyphens/parens, total 7–20 chars.
var phoneRe = regexp.MustCompile(`^\+?[0-9\s\-()+]{7,20}$`)

// emailRe: simplified RFC 5322 — local@domain.tld.
var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// sectionRe: 1–10 alphanumeric characters.
var sectionRe = regexp.MustCompile(`^[A-Za-z0-9]{1,10}$`)

// ValidateStudent validates all fields of a student create/update request.
// Returns ALL failing fields at once; an empty slice means the input is valid.
func ValidateStudent(s models.Student) []ValidationError {
	var errs []ValidationError

	// student_id
	if strings.TrimSpace(s.StudentID) == "" {
		errs = append(errs, ValidationError{"student_id", "Student ID is required."})
	} else if !studentIDRe.MatchString(s.StudentID) {
		errs = append(errs, ValidationError{"student_id", "Student ID must be 3–20 alphanumeric characters or hyphens."})
	}

	// full_name
	if strings.TrimSpace(s.FullName) == "" {
		errs = append(errs, ValidationError{"full_name", "Full name is required."})
	} else {
		n := []rune(strings.TrimSpace(s.FullName))
		if len(n) < 2 || len(n) > 100 {
			errs = append(errs, ValidationError{"full_name", "Full name must be 2–100 characters."})
		} else if !isValidName(s.FullName) {
			errs = append(errs, ValidationError{"full_name", "Full name may only contain letters, spaces, hyphens, and apostrophes."})
		}
	}

	// email
	if strings.TrimSpace(s.Email) == "" {
		errs = append(errs, ValidationError{"email", "Email is required."})
	} else if len(s.Email) > 255 {
		errs = append(errs, ValidationError{"email", "Email must be at most 255 characters."})
	} else if !emailRe.MatchString(s.Email) {
		errs = append(errs, ValidationError{"email", "Email must be a valid email address."})
	}

	// phone (optional)
	if s.Phone != "" && !phoneRe.MatchString(s.Phone) {
		errs = append(errs, ValidationError{"phone", "Phone number must be 7–20 characters: digits, spaces, hyphens, parentheses, or leading '+' only."})
	}

	// department
	if strings.TrimSpace(s.Department) == "" {
		errs = append(errs, ValidationError{"department", "Department is required."})
	} else if !isValidDepartment(s.Department) {
		errs = append(errs, ValidationError{"department", "Department must be one of the defined department list."})
	}

	// year
	if s.Year < 1 || s.Year > 6 {
		errs = append(errs, ValidationError{"year", "Year must be an integer between 1 and 6."})
	}

	// section
	if strings.TrimSpace(s.Section) == "" {
		errs = append(errs, ValidationError{"section", "Section is required."})
	} else if !sectionRe.MatchString(s.Section) {
		errs = append(errs, ValidationError{"section", "Section must be 1–10 alphanumeric characters."})
	}

	// dob
	if strings.TrimSpace(s.DOB) == "" {
		errs = append(errs, ValidationError{"dob", "Date of birth is required."})
	} else {
		dob, err := time.Parse("2006-01-02", s.DOB)
		if err != nil {
			errs = append(errs, ValidationError{"dob", "Date of birth must be in YYYY-MM-DD format."})
		} else {
			// Student must be at least 15 years old.
			minDOB := time.Now().AddDate(-15, 0, 0)
			if dob.After(minDOB) {
				errs = append(errs, ValidationError{"dob", "Student must be at least 15 years old."})
			}
		}
	}

	return errs
}

// isValidName checks that a full_name contains only letters (any script),
// spaces, hyphens, and apostrophes.
func isValidName(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' {
			return false
		}
	}
	return true
}
