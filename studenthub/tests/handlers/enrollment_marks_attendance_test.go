package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"studenthub/handlers"
	"studenthub/models"
)

// buildFullMuxDB builds the full mux with the real config type.
func buildFullMuxDB(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	handlers.RegisterStudentRoutes(mux, db)
	handlers.RegisterCourseRoutes(mux, db)
	handlers.RegisterEnrollmentRoutes(mux, db)
	handlers.RegisterMarksRoutes(mux, db)
	handlers.RegisterAttendanceRoutes(mux, db)
	handlers.RegisterDashboardFromDB(mux, db, 75.0)
	return mux
}

// seedStudentAndCourse creates one student and one course, returns their IDs.
func seedStudentAndCourse(t *testing.T, db *sql.DB) (studentID, courseID int) {
	t.Helper()
	s, err := models.CreateStudent(db, models.Student{
		StudentID: "STU9001", FullName: "Test Student",
		Email: "test9001@x.com", Department: "Computer Science",
		Year: 1, Section: "A", DOB: "2004-01-01",
	})
	if err != nil {
		t.Fatalf("seedStudent: %v", err)
	}
	c, err := models.CreateCourse(db, models.Course{
		CourseCode: "TST101", CourseName: "Test Course", Credits: 3, Department: "Computer Science",
	})
	if err != nil {
		t.Fatalf("seedCourse: %v", err)
	}
	return s.ID, c.ID
}

// seedEnrollment creates an enrollment for the given student+course.
func seedEnrollment(t *testing.T, db *sql.DB, studentID, courseID int) models.Enrollment {
	t.Helper()
	e, err := models.CreateEnrollment(db, studentID, courseID)
	if err != nil {
		t.Fatalf("seedEnrollment: %v", err)
	}
	return e
}

// -------- Enrollment tests --------

func TestCreateEnrollment_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterStudentRoutes(mux, db)
	handlers.RegisterCourseRoutes(mux, db)
	handlers.RegisterEnrollmentRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)

	body, _ := json.Marshal(map[string]any{"student_id": studentID, "course_id": courseID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var e models.Enrollment
	json.NewDecoder(w.Body).Decode(&e)
	if e.ID == 0 {
		t.Error("expected non-zero enrollment id")
	}
}

func TestCreateEnrollment_Duplicate_Returns409(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterStudentRoutes(mux, db)
	handlers.RegisterCourseRoutes(mux, db)
	handlers.RegisterEnrollmentRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{"student_id": studentID, "course_id": courseID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateEnrollment_NonExistentStudent_Returns404(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterEnrollmentRoutes(mux, db)

	body, _ := json.Marshal(map[string]any{"student_id": 9999, "course_id": 9999})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteEnrollment_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterEnrollmentRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	e := seedEnrollment(t, db, studentID, courseID)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/enrollments/%d", e.ID), nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestGetCoursesForStudent_ReturnsEnrolledCourses(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterStudentRoutes(mux, db)
	handlers.RegisterEnrollmentRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/students/%d/courses", studentID), nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var courses []models.Course
	json.NewDecoder(w.Body).Decode(&courses)
	if len(courses) != 1 {
		t.Errorf("expected 1 course, got %d", len(courses))
	}
}

// -------- Marks tests --------

func TestCreateMarks_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterMarksRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{
		"student_id": studentID,
		"course_id":  courseID,
		"marks":      82.5,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var m models.Marks
	json.NewDecoder(w.Body).Decode(&m)
	if m.Grade != "A" {
		t.Errorf("expected grade A for 82.5, got %s", m.Grade)
	}
}

func TestCreateMarks_NotEnrolled_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterMarksRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	// Do NOT enroll the student.

	body, _ := json.Marshal(map[string]any{
		"student_id": studentID,
		"course_id":  courseID,
		"marks":      75.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateMarks_OutOfRange_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterMarksRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{
		"student_id": studentID,
		"course_id":  courseID,
		"marks":      101.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateMarks_Duplicate_Returns409(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterMarksRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{"student_id": studentID, "course_id": courseID, "marks": 80.0})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/marks", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(httptest.NewRecorder(), req1)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/marks", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req2)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateMarks_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterMarksRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	// Create marks.
	createBody, _ := json.Marshal(map[string]any{"student_id": studentID, "course_id": courseID, "marks": 70.0})
	cr := httptest.NewRequest(http.MethodPost, "/api/v1/marks", bytes.NewReader(createBody))
	cr.Header.Set("Content-Type", "application/json")
	crw := httptest.NewRecorder()
	mux.ServeHTTP(crw, cr)
	var created models.Marks
	json.NewDecoder(crw.Body).Decode(&created)

	// Update marks.
	updateBody, _ := json.Marshal(map[string]any{"marks": 91.0})
	ur := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/marks/%d", created.ID), bytes.NewReader(updateBody))
	ur.Header.Set("Content-Type", "application/json")
	uw := httptest.NewRecorder()
	mux.ServeHTTP(uw, ur)

	if uw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", uw.Code, uw.Body.String())
	}
	var updated models.Marks
	json.NewDecoder(uw.Body).Decode(&updated)
	if updated.Grade != "A+" {
		t.Errorf("expected A+ after updating to 91, got %s", updated.Grade)
	}
}

func TestGetMarksForStudent_ComputesAverage(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterMarksRoutes(mux, db)

	studentID, _ := seedStudentAndCourse(t, db)

	// Add two more courses and enroll + add marks.
	c2, _ := models.CreateCourse(db, models.Course{CourseCode: "TST102", CourseName: "C2", Credits: 2, Department: "Mathematics"})
	c3, _ := models.CreateCourse(db, models.Course{CourseCode: "TST103", CourseName: "C3", Credits: 2, Department: "Physics"})
	c1ID := func() int {
		c, _ := models.GetAllCourses(db)
		for _, c := range c {
			if c.CourseCode == "TST101" {
				return c.ID
			}
		}
		return 0
	}()
	models.CreateEnrollment(db, studentID, c1ID)
	models.CreateEnrollment(db, studentID, c2.ID)
	models.CreateEnrollment(db, studentID, c3.ID)
	models.CreateMarks(db, models.Marks{StudentID: studentID, CourseID: c1ID, Marks: 80})
	models.CreateMarks(db, models.Marks{StudentID: studentID, CourseID: c2.ID, Marks: 90})
	models.CreateMarks(db, models.Marks{StudentID: studentID, CourseID: c3.ID, Marks: 70})

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/students/%d/marks", studentID), nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp models.StudentMarksResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.AverageMarks == nil {
		t.Fatal("expected non-nil average_marks")
	}
	if *resp.AverageMarks != 80.0 {
		t.Errorf("expected average 80.0, got %f", *resp.AverageMarks)
	}
}

// -------- Attendance tests --------

func TestCreateAttendance_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{
		"student_id":    studentID,
		"course_id":     courseID,
		"total_classes": 48,
		"attended":      36,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attendance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var a models.Attendance
	json.NewDecoder(w.Body).Decode(&a)
	if a.Percentage != 75.00 {
		t.Errorf("expected percentage 75.00, got %f", a.Percentage)
	}
}

func TestCreateAttendance_AttendedExceedsTotal_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{
		"student_id":    studentID,
		"course_id":     courseID,
		"total_classes": 10,
		"attended":      15, // exceeds total
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attendance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAttendance_NotEnrolled_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)

	body, _ := json.Marshal(map[string]any{
		"student_id":    studentID,
		"course_id":     courseID,
		"total_classes": 10,
		"attended":      8,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attendance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAttendance_Duplicate_Returns409(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	body, _ := json.Marshal(map[string]any{"student_id": studentID, "course_id": courseID, "total_classes": 10, "attended": 8})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/attendance", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(httptest.NewRecorder(), req1)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/attendance", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req2)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetLowAttendance_BelowThreshold(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.SetAttendanceThreshold(75.0)
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)
	// 60% attendance — below threshold.
	models.CreateAttendance(db, models.Attendance{StudentID: studentID, CourseID: courseID, TotalClasses: 10, Attended: 6})

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/attendance/low", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []models.LowAttendanceStudent
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 {
		t.Errorf("expected 1 low-attendance student, got %d", len(list))
	}
}

func TestGetLowAttendance_ExactlyAtThreshold_NotIncluded(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.SetAttendanceThreshold(75.0)
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)
	// Exactly 75.00% — must NOT be in low-attendance list.
	models.CreateAttendance(db, models.Attendance{StudentID: studentID, CourseID: courseID, TotalClasses: 4, Attended: 3})

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/attendance/low", nil))

	var list []models.LowAttendanceStudent
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 0 {
		t.Errorf("expected 0 low-attendance students at exactly 75%%, got %d", len(list))
	}
}

func TestUpdateAttendance_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)

	// Create attendance.
	createBody, _ := json.Marshal(map[string]any{"student_id": studentID, "course_id": courseID, "total_classes": 10, "attended": 7})
	cr := httptest.NewRequest(http.MethodPost, "/api/v1/attendance", bytes.NewReader(createBody))
	cr.Header.Set("Content-Type", "application/json")
	crw := httptest.NewRecorder()
	mux.ServeHTTP(crw, cr)
	var created models.Attendance
	json.NewDecoder(crw.Body).Decode(&created)

	// Update.
	updateBody, _ := json.Marshal(map[string]any{"total_classes": 48, "attended": 36})
	ur := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/attendance/%d", created.ID), bytes.NewReader(updateBody))
	ur.Header.Set("Content-Type", "application/json")
	uw := httptest.NewRecorder()
	mux.ServeHTTP(uw, ur)

	if uw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", uw.Code, uw.Body.String())
	}
	var updated models.Attendance
	json.NewDecoder(uw.Body).Decode(&updated)
	if updated.Percentage != 75.00 {
		t.Errorf("expected 75.00%%, got %f", updated.Percentage)
	}
}

// -------- Dashboard tests --------

func TestGetDashboard_Structure(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterDashboardFromDB(mux, db, 75.0)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.DashboardResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	// Empty DB: counts should be 0.
	if resp.TotalStudents != 0 {
		t.Errorf("expected 0 students, got %d", resp.TotalStudents)
	}
	if resp.LowAttendanceStudents == nil {
		t.Error("expected non-nil low_attendance_students array")
	}
	if resp.RecentStudents == nil {
		t.Error("expected non-nil recent_students array")
	}
}

func TestGetDashboard_LowAttendanceStudent(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterDashboardFromDB(mux, db, 75.0)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)
	models.CreateAttendance(db, models.Attendance{StudentID: studentID, CourseID: courseID, TotalClasses: 10, Attended: 6}) // 60%

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil))

	var resp models.DashboardResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.LowAttendanceStudents) != 1 {
		t.Errorf("expected 1 low-attendance student, got %d", len(resp.LowAttendanceStudents))
	}
}

// -------- Cascade delete test --------

func TestDeleteStudent_CascadesMarksAndAttendance(t *testing.T) {
	db := newTestDB(t)
	mux := http.NewServeMux()
	handlers.RegisterStudentRoutes(mux, db)
	handlers.RegisterMarksRoutes(mux, db)
	handlers.RegisterAttendanceRoutes(mux, db)

	studentID, courseID := seedStudentAndCourse(t, db)
	seedEnrollment(t, db, studentID, courseID)
	models.CreateMarks(db, models.Marks{StudentID: studentID, CourseID: courseID, Marks: 80})
	models.CreateAttendance(db, models.Attendance{StudentID: studentID, CourseID: courseID, TotalClasses: 10, Attended: 8})

	// Delete student.
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/students/%d", studentID), nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify marks are gone.
	var markCount int
	db.QueryRow("SELECT COUNT(*) FROM marks WHERE student_id = ?", studentID).Scan(&markCount)
	if markCount != 0 {
		t.Errorf("expected 0 marks after student delete, got %d", markCount)
	}

	// Verify attendance is gone.
	var attCount int
	db.QueryRow("SELECT COUNT(*) FROM attendance WHERE student_id = ?", studentID).Scan(&attCount)
	if attCount != 0 {
		t.Errorf("expected 0 attendance records after student delete, got %d", attCount)
	}
}
