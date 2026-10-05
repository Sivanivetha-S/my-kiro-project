package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"studenthub/handlers"
	"studenthub/models"
)

// buildCourseMux wires course routes onto a fresh ServeMux backed by db.
func buildCourseMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	handlers.RegisterCourseRoutes(mux, db)
	return mux
}

func validCourseBody() []byte {
	b, _ := json.Marshal(map[string]any{
		"course_code": "CS101",
		"course_name": "Introduction to Programming",
		"credits":     4,
		"department":  "Computer Science",
	})
	return b
}

func TestCreateCourse_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp models.Course
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.CourseCode != "CS101" {
		t.Errorf("expected CS101, got %s", resp.CourseCode)
	}
	if resp.ID == 0 {
		t.Error("expected non-zero id")
	}
}

func TestCreateCourse_DuplicateCourseCode_Returns409(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	mux.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody())))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody()))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateCourse_InvalidCredits_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	body, _ := json.Marshal(map[string]any{
		"course_code": "CS102",
		"course_name": "Advanced Programming",
		"credits":     7, // invalid: max is 6
		"department":  "Computer Science",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateCourse_LowercaseCourseCode_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	body, _ := json.Marshal(map[string]any{
		"course_code": "cs101",
		"course_name": "Some Course",
		"credits":     3,
		"department":  "Computer Science",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListCourses_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody()))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(httptest.NewRecorder(), req)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []models.Course
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 {
		t.Errorf("expected 1 course, got %d", len(list))
	}
}

func TestListCourses_EmptyReturnsArray(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/courses", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []models.Course
	json.NewDecoder(w.Body).Decode(&list)
	if list == nil {
		t.Error("expected empty array, not null")
	}
}

func TestGetCourse_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	cr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody()))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(cr, req)
	var created models.Course
	json.NewDecoder(cr.Body).Decode(&created)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+jsonInt(created.ID), nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCourse_NotFound_Returns404(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/courses/999", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateCourse_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	cr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody()))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(cr, req)
	var created models.Course
	json.NewDecoder(cr.Body).Decode(&created)

	updateBody, _ := json.Marshal(map[string]any{
		"course_code": "CS101",
		"course_name": "Updated Course Name",
		"credits":     3,
		"department":  "Computer Science",
	})
	w := httptest.NewRecorder()
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/courses/"+jsonInt(created.ID), bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, updateReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updated models.Course
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.CourseName != "Updated Course Name" {
		t.Errorf("expected Updated Course Name, got %s", updated.CourseName)
	}
}

func TestDeleteCourse_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	cr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", bytes.NewReader(validCourseBody()))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(cr, req)
	var created models.Course
	json.NewDecoder(cr.Body).Decode(&created)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/courses/"+jsonInt(created.ID), nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify gone.
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/v1/courses/"+jsonInt(created.ID), nil))
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestDeleteCourse_NotFound_Returns404(t *testing.T) {
	db := newTestDB(t)
	mux := buildCourseMux(db)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/courses/999", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
