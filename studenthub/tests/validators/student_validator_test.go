package validators_test

import (
	"testing"

	"studenthub/models"
	"studenthub/validators"
)

// validStudent returns a fully valid student for use as a baseline.
func validStudent() models.Student {
	return models.Student{
		StudentID:  "STU2024001",
		FullName:   "Aisha Rajan",
		Email:      "aisha@example.com",
		Phone:      "+91-9876543210",
		Department: "Computer Science",
		Year:       2,
		Section:    "A",
		DOB:        "2004-06-15",
	}
}

func hasFieldError(errs []validators.ValidationError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

func TestValidateStudent_ValidInput_NoErrors(t *testing.T) {
	errs := validators.ValidateStudent(validStudent())
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid student, got: %v", errs)
	}
}

func TestValidateStudent_MissingStudentID(t *testing.T) {
	s := validStudent()
	s.StudentID = ""
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "student_id") {
		t.Error("expected student_id error for empty student ID")
	}
}

func TestValidateStudent_StudentIDTooShort(t *testing.T) {
	s := validStudent()
	s.StudentID = "AB" // 2 chars — min is 3
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "student_id") {
		t.Error("expected student_id error for 2-char ID")
	}
}

func TestValidateStudent_StudentIDTooLong(t *testing.T) {
	s := validStudent()
	s.StudentID = "ABCDEFGHIJKLMNOPQRSTU" // 21 chars — max is 20
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "student_id") {
		t.Error("expected student_id error for 21-char ID")
	}
}

func TestValidateStudent_StudentIDWithSpecialChars(t *testing.T) {
	s := validStudent()
	s.StudentID = "STU@2024"
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "student_id") {
		t.Error("expected student_id error for ID with @")
	}
}

func TestValidateStudent_StudentIDWithHyphen_Valid(t *testing.T) {
	s := validStudent()
	s.StudentID = "STU-2024"
	errs := validators.ValidateStudent(s)
	if hasFieldError(errs, "student_id") {
		t.Error("hyphen should be allowed in student_id")
	}
}

func TestValidateStudent_MissingFullName(t *testing.T) {
	s := validStudent()
	s.FullName = ""
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "full_name") {
		t.Error("expected full_name error for empty name")
	}
}

func TestValidateStudent_InvalidEmail(t *testing.T) {
	s := validStudent()
	s.Email = "not-an-email"
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "email") {
		t.Error("expected email error for invalid email")
	}
}

func TestValidateStudent_InvalidPhone(t *testing.T) {
	s := validStudent()
	s.Phone = "123" // too short
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "phone") {
		t.Error("expected phone error for short phone")
	}
}

func TestValidateStudent_EmptyPhone_Valid(t *testing.T) {
	s := validStudent()
	s.Phone = ""
	errs := validators.ValidateStudent(s)
	if hasFieldError(errs, "phone") {
		t.Error("empty phone should be valid (optional field)")
	}
}

func TestValidateStudent_InvalidDepartment(t *testing.T) {
	s := validStudent()
	s.Department = "Astronomy"
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "department") {
		t.Error("expected department error for invalid department")
	}
}

func TestValidateStudent_YearTooLow(t *testing.T) {
	s := validStudent()
	s.Year = 0
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "year") {
		t.Error("expected year error for 0")
	}
}

func TestValidateStudent_YearTooHigh(t *testing.T) {
	s := validStudent()
	s.Year = 7
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "year") {
		t.Error("expected year error for 7")
	}
}

func TestValidateStudent_InvalidSection(t *testing.T) {
	s := validStudent()
	s.Section = "Section-A!" // special chars not allowed
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "section") {
		t.Error("expected section error for invalid chars")
	}
}

func TestValidateStudent_DOBTooYoung(t *testing.T) {
	s := validStudent()
	s.DOB = "2020-01-01" // far too young
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "dob") {
		t.Error("expected dob error for too-young student")
	}
}

func TestValidateStudent_DOBExactly15_Valid(t *testing.T) {
	s := validStudent()
	// Use a date exactly 15 years ago (as the validator uses time.Now()).
	// Since we can't control time.Now() in tests, use a known-valid date.
	s.DOB = "2004-01-01" // at least 15 years before 2026.
	errs := validators.ValidateStudent(s)
	if hasFieldError(errs, "dob") {
		t.Error("a student born in 2004 should be >= 15 years old in 2026")
	}
}

func TestValidateStudent_InvalidDOBFormat(t *testing.T) {
	s := validStudent()
	s.DOB = "15-06-2004" // wrong format
	errs := validators.ValidateStudent(s)
	if !hasFieldError(errs, "dob") {
		t.Error("expected dob error for wrong format")
	}
}

func TestValidateStudent_MultipleErrors_ReturnsAll(t *testing.T) {
	s := models.Student{} // all zero values
	errs := validators.ValidateStudent(s)
	// Expect errors for student_id, full_name, email, department, year, section, dob.
	required := []string{"student_id", "full_name", "email", "department", "year", "section", "dob"}
	for _, field := range required {
		if !hasFieldError(errs, field) {
			t.Errorf("expected error for field %s in all-empty student", field)
		}
	}
}
