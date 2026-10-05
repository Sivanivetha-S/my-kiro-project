package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"studenthub/models"
	"studenthub/validators"
)

// RegisterStudentRoutes registers all /api/v1/students routes on mux.
// db is closed to the handlers via closure.
func RegisterStudentRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("POST /api/v1/students", createStudent(db))
	mux.HandleFunc("GET /api/v1/students", listStudents(db))
	mux.HandleFunc("GET /api/v1/students/{id}", getStudent(db))
	mux.HandleFunc("PUT /api/v1/students/{id}", updateStudent(db))
	mux.HandleFunc("DELETE /api/v1/students/{id}", deleteStudent(db))
}

func createStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.Student
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}

		if errs := validators.ValidateStudent(s); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}

		created, err := models.CreateStudent(db, s)
		if err != nil {
			if errors.Is(err, models.ErrDuplicate) {
				writeError(w, http.StatusConflict, "Student ID or email already exists", "DUPLICATE")
				return
			}
			log.Printf("createStudent: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func listStudents(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		year := 0
		if y := q.Get("year"); y != "" {
			if v, err := strconv.Atoi(y); err == nil {
				year = v
			}
		}

		filter := models.StudentFilter{
			Search:     q.Get("search"),
			Department: q.Get("department"),
			Year:       year,
		}

		students, err := models.GetAllStudents(db, filter)
		if err != nil {
			log.Printf("listStudents: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}

		// Return empty array rather than null when there are no results.
		if students == nil {
			students = []models.Student{}
		}
		writeJSON(w, http.StatusOK, students)
	}
}

func getStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid student ID", "BAD_REQUEST")
			return
		}

		student, err := models.GetStudentByID(db, id)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Student not found", "NOT_FOUND")
				return
			}
			log.Printf("getStudent: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, student)
	}
}

func updateStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid student ID", "BAD_REQUEST")
			return
		}

		var s models.Student
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}

		if errs := validators.ValidateStudent(s); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}

		updated, err := models.UpdateStudent(db, id, s)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrNotFound):
				writeError(w, http.StatusNotFound, "Student not found", "NOT_FOUND")
			case errors.Is(err, models.ErrDuplicate):
				writeError(w, http.StatusConflict, "Student ID or email already exists", "DUPLICATE")
			default:
				log.Printf("updateStudent: %v", err)
				writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			}
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func deleteStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid student ID", "BAD_REQUEST")
			return
		}

		if err := models.DeleteStudent(db, id); err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Student not found", "NOT_FOUND")
				return
			}
			log.Printf("deleteStudent: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// parseID extracts a named path parameter as an integer.
func parseID(r *http.Request, name string) (int, error) {
	return strconv.Atoi(r.PathValue(name))
}
