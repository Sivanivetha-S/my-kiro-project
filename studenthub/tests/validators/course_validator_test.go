package validators_test

import (
	"testing"

	"studenthub/models"
	"studenthub/validators"
)

func validCourse() models.Course {
	return models.Course{
		CourseCode: "CS101",
		CourseName: "Introduction to Programming",
		Credits:    4,
		Department: "Computer Science",
	}
}

func TestValidateCourse_ValidInput_NoErrors(t *testing.T) {
	errs := validators.ValidateCourse(validCourse())
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid course, got: %v", errs)
	}
}

func TestValidateCourse_MissingCourseCode(t *testing.T) {
	c := validCourse()
	c.CourseCode = ""
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "course_code") {
		t.Error("expected course_code error for empty code")
	}
}

func TestValidateCourse_LowercaseCourseCode(t *testing.T) {
	c := validCourse()
	c.CourseCode = "cs101"
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "course_code") {
		t.Error("expected course_code error for lowercase code")
	}
}

func TestValidateCourse_CourseCodeTooShort(t *testing.T) {
	c := validCourse()
	c.CourseCode = "A" // 1 char — min is 2
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "course_code") {
		t.Error("expected course_code error for 1-char code")
	}
}

func TestValidateCourse_MissingCourseName(t *testing.T) {
	c := validCourse()
	c.CourseName = ""
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "course_name") {
		t.Error("expected course_name error for empty name")
	}
}

func TestValidateCourse_CourseNameTooShort(t *testing.T) {
	c := validCourse()
	c.CourseName = "AB" // 2 chars — min is 3
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "course_name") {
		t.Error("expected course_name error for 2-char name")
	}
}

func TestValidateCourse_CreditsZero(t *testing.T) {
	c := validCourse()
	c.Credits = 0
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "credits") {
		t.Error("expected credits error for 0")
	}
}

func TestValidateCourse_CreditsSeven(t *testing.T) {
	c := validCourse()
	c.Credits = 7
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "credits") {
		t.Error("expected credits error for 7")
	}
}

func TestValidateCourse_CreditsOne_Valid(t *testing.T) {
	c := validCourse()
	c.Credits = 1
	errs := validators.ValidateCourse(c)
	if hasFieldError(errs, "credits") {
		t.Error("credits=1 should be valid")
	}
}

func TestValidateCourse_CreditsSix_Valid(t *testing.T) {
	c := validCourse()
	c.Credits = 6
	errs := validators.ValidateCourse(c)
	if hasFieldError(errs, "credits") {
		t.Error("credits=6 should be valid")
	}
}

func TestValidateCourse_InvalidDepartment(t *testing.T) {
	c := validCourse()
	c.Department = "Underwater Basket Weaving"
	errs := validators.ValidateCourse(c)
	if !hasFieldError(errs, "department") {
		t.Error("expected department error for invalid department")
	}
}
