package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"studenthub/handlers"
	"studenthub/models"
)

// buildStudentMux wires student routes onto a fresh ServeMux backed by db.
func buildStudentMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	handlers.RegisterStudentRoutes(mux, db)
	return mux
}

// validStudentBody returns a JSON body for a valid student create request.
func validStudentBody() []byte {
	b, _ := json.Marshal(map[string]any{
		"student_id": "STU2024001",
		"full_name":  "Aisha Rajan",
		"email":      "aisha@example.com",
		"phone":      "+91-9876543210",
		"department": "Computer Science",
		"year":       2,
		"section":    "A",
		"dob":        "2004-06-15",
	})
	return b
}

func TestCreateStudent_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(validStudentBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.Student
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.StudentID != "STU2024001" {
		t.Errorf("expected student_id STU2024001, got %s", resp.StudentID)
	}
	if resp.ID == 0 {
		t.Error("expected non-zero id")
	}
	if resp.CreatedAt == "" {
		t.Error("expected non-empty created_at")
	}
}

func TestCreateStudent_MissingFullName_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	body, _ := json.Marshal(map[string]any{
		"student_id": "STU2024002",
		"email":      "b@example.com",
		"department": "Computer Science",
		"year":       1,
		"section":    "A",
		"dob":        "2005-01-01",
		// full_name intentionally omitted
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateStudent_DuplicateStudentID_Returns409(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	// Create once.
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(validStudentBody()))
	req1.Header.Set("Content-Type", "application/json")
	httptest.NewRecorder() // discard
	mux.ServeHTTP(httptest.NewRecorder(), req1)

	// Create duplicate.
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(validStudentBody()))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestCreateStudent_InvalidDOB_Returns400(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	body, _ := json.Marshal(map[string]any{
		"student_id": "STU2024003",
		"full_name":  "Test User",
		"email":      "test@example.com",
		"department": "Computer Science",
		"year":       1,
		"section":    "A",
		"dob":        "2020-01-01", // too young
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListStudents_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	// Create a student first.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(validStudentBody()))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/students", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req2)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var list []models.Student
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 student, got %d", len(list))
	}
}

func TestListStudents_FilterByDepartment(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	// Create CS student.
	mux.ServeHTTP(httptest.NewRecorder(),
		newJSONRequest(http.MethodPost, "/api/v1/students", map[string]any{
			"student_id": "STU0001", "full_name": "Alice Smith",
			"email": "alice@x.com", "department": "Computer Science",
			"year": 1, "section": "A", "dob": "2004-01-01",
		}))

	// Create Math student.
	mux.ServeHTTP(httptest.NewRecorder(),
		newJSONRequest(http.MethodPost, "/api/v1/students", map[string]any{
			"student_id": "STU0002", "full_name": "Bob Jones",
			"email": "bob@x.com", "department": "Mathematics",
			"year": 2, "section": "B", "dob": "2003-01-01",
		}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/students?department=Computer+Science", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var list []models.Student
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 || list[0].Department != "Computer Science" {
		t.Errorf("expected 1 CS student, got %v", list)
	}
}

func TestGetStudent_NotFound_Returns404(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/students/999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateStudent_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	// Create.
	cr := httptest.NewRecorder()
	mux.ServeHTTP(cr, newJSONRequest(http.MethodPost, "/api/v1/students", map[string]any{
		"student_id": "STU0010", "full_name": "Original Name",
		"email": "orig@x.com", "department": "Mathematics",
		"year": 1, "section": "A", "dob": "2004-01-01",
	}))
	var created models.Student
	json.NewDecoder(cr.Body).Decode(&created)

	// Update.
	updURL := "/api/v1/students/" + jsonInt(created.ID)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, newJSONRequest(http.MethodPut, updURL, map[string]any{
		"student_id": "STU0010", "full_name": "Updated Name",
		"email": "orig@x.com", "department": "Mathematics",
		"year": 2, "section": "B", "dob": "2004-01-01",
	}))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updated models.Student
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.FullName != "Updated Name" {
		t.Errorf("expected Updated Name, got %s", updated.FullName)
	}
}

func TestDeleteStudent_HappyPath(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	// Create.
	cr := httptest.NewRecorder()
	mux.ServeHTTP(cr, newJSONRequest(http.MethodPost, "/api/v1/students", map[string]any{
		"student_id": "STU0020", "full_name": "Del Me",
		"email": "del@x.com", "department": "Physics",
		"year": 1, "section": "A", "dob": "2004-01-01",
	}))
	var created models.Student
	json.NewDecoder(cr.Body).Decode(&created)

	// Delete.
	delURL := "/api/v1/students/" + jsonInt(created.ID)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, delURL, nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify gone.
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, delURL, nil))
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestDeleteStudent_NotFound_Returns404(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/students/999", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestSearchStudents_PartialName(t *testing.T) {
	db := newTestDB(t)
	mux := buildStudentMux(db)

	mux.ServeHTTP(httptest.NewRecorder(), newJSONRequest(http.MethodPost, "/api/v1/students", map[string]any{
		"student_id": "STU0030", "full_name": "Aisha Rajan",
		"email": "aisha2@x.com", "department": "Computer Science",
		"year": 1, "section": "A", "dob": "2004-01-01",
	}))
	mux.ServeHTTP(httptest.NewRecorder(), newJSONRequest(http.MethodPost, "/api/v1/students", map[string]any{
		"student_id": "STU0031", "full_name": "Rohan Mehta",
		"email": "rohan@x.com", "department": "Computer Science",
		"year": 1, "section": "A", "dob": "2004-01-01",
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/students?search=aisha", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var list []models.Student
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 || list[0].FullName != "Aisha Rajan" {
		t.Errorf("search for 'aisha' returned unexpected results: %v", list)
	}
}

// --- helpers ---

func newJSONRequest(method, path string, body map[string]any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func jsonInt(n int) string {
	return strconv.Itoa(n)
}
