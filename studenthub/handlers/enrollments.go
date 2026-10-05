package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"studenthub/models"
)

// RegisterEnrollmentRoutes registers all enrollment-related routes on mux.
func RegisterEnrollmentRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("POST /api/v1/enrollments", createEnrollment(db))
	mux.HandleFunc("DELETE /api/v1/enrollments/{id}", deleteEnrollment(db))
	mux.HandleFunc("GET /api/v1/students/{id}/courses", getCoursesForStudent(db))
	mux.HandleFunc("GET /api/v1/courses/{id}/students", getStudentsForCourse(db))
}

func createEnrollment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			StudentID int `json:"student_id"`
			CourseID  int `json:"course_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}
		if req.StudentID <= 0 || req.CourseID <= 0 {
			writeError(w, http.StatusBadRequest, "student_id and course_id are required", "VALIDATION_ERROR")
			return
		}

		enrollment, err := models.CreateEnrollment(db, req.StudentID, req.CourseID)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrNotFound):
				writeError(w, http.StatusNotFound, "Student or course not found", "NOT_FOUND")
			case errors.Is(err, models.ErrDuplicate):
				writeError(w, http.StatusConflict, "Student is already enrolled in this course", "DUPLICATE_ENROLLMENT")
			default:
				log.Printf("createEnrollment: %v", err)
				writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			}
			return
		}
		writeJSON(w, http.StatusCreated, enrollment)
	}
}

func deleteEnrollment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid enrollment ID", "BAD_REQUEST")
			return
		}
		if err := models.DeleteEnrollment(db, id); err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Enrollment not found", "NOT_FOUND")
				return
			}
			log.Printf("deleteEnrollment: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getCoursesForStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid student ID", "BAD_REQUEST")
			return
		}
		courses, err := models.GetCoursesForStudent(db, id)
		if err != nil {
			log.Printf("getCoursesForStudent: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, courses)
	}
}

func getStudentsForCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid course ID", "BAD_REQUEST")
			return
		}
		students, err := models.GetStudentsForCourse(db, id)
		if err != nil {
			log.Printf("getStudentsForCourse: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, students)
	}
}
