package validators_test

import (
	"testing"

	"studenthub/models"
	"studenthub/validators"
)

func TestValidateMarks_ValidInput_NoErrors(t *testing.T) {
	m := models.Marks{StudentID: 1, CourseID: 1, Marks: 75.0}
	errs := validators.ValidateMarks(m)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid marks, got: %v", errs)
	}
}

func TestValidateMarks_ZeroMarks_Valid(t *testing.T) {
	m := models.Marks{StudentID: 1, CourseID: 1, Marks: 0.0}
	errs := validators.ValidateMarks(m)
	if hasFieldError(errs, "marks") {
		t.Error("0.0 marks should be valid")
	}
}

func TestValidateMarks_HundredMarks_Valid(t *testing.T) {
	m := models.Marks{StudentID: 1, CourseID: 1, Marks: 100.0}
	errs := validators.ValidateMarks(m)
	if hasFieldError(errs, "marks") {
		t.Error("100.0 marks should be valid")
	}
}

func TestValidateMarks_NegativeMarks(t *testing.T) {
	m := models.Marks{StudentID: 1, CourseID: 1, Marks: -1.0}
	errs := validators.ValidateMarks(m)
	if !hasFieldError(errs, "marks") {
		t.Error("expected marks error for -1.0")
	}
}

func TestValidateMarks_Over100(t *testing.T) {
	m := models.Marks{StudentID: 1, CourseID: 1, Marks: 100.01}
	errs := validators.ValidateMarks(m)
	if !hasFieldError(errs, "marks") {
		t.Error("expected marks error for 100.01")
	}
}

func TestValidateMarks_MissingStudentID(t *testing.T) {
	m := models.Marks{StudentID: 0, CourseID: 1, Marks: 75.0}
	errs := validators.ValidateMarks(m)
	if !hasFieldError(errs, "student_id") {
		t.Error("expected student_id error for 0")
	}
}

func TestValidateMarks_MissingCourseID(t *testing.T) {
	m := models.Marks{StudentID: 1, CourseID: 0, Marks: 75.0}
	errs := validators.ValidateMarks(m)
	if !hasFieldError(errs, "course_id") {
		t.Error("expected course_id error for 0")
	}
}

func TestValidateMarksUpdate_ValidMarks(t *testing.T) {
	errs := validators.ValidateMarksUpdate(50.0)
	if len(errs) != 0 {
		t.Errorf("expected no errors for 50.0, got: %v", errs)
	}
}

func TestValidateMarksUpdate_OutOfRange(t *testing.T) {
	errs := validators.ValidateMarksUpdate(101.0)
	if !hasFieldError(errs, "marks") {
		t.Error("expected marks error for 101.0 update")
	}
}
