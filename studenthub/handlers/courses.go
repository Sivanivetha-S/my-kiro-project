package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"studenthub/models"
	"studenthub/validators"
)

// RegisterCourseRoutes registers all /api/v1/courses routes on mux.
func RegisterCourseRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("POST /api/v1/courses", createCourse(db))
	mux.HandleFunc("GET /api/v1/courses", listCourses(db))
	mux.HandleFunc("GET /api/v1/courses/{id}", getCourse(db))
	mux.HandleFunc("PUT /api/v1/courses/{id}", updateCourse(db))
	mux.HandleFunc("DELETE /api/v1/courses/{id}", deleteCourse(db))
}

func createCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c models.Course
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}
		if errs := validators.ValidateCourse(c); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}
		created, err := models.CreateCourse(db, c)
		if err != nil {
			if errors.Is(err, models.ErrDuplicate) {
				writeError(w, http.StatusConflict, "Course code already exists", "DUPLICATE_COURSE_CODE")
				return
			}
			log.Printf("createCourse: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func listCourses(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courses, err := models.GetAllCourses(db)
		if err != nil {
			log.Printf("listCourses: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		if courses == nil {
			courses = []models.Course{}
		}
		writeJSON(w, http.StatusOK, courses)
	}
}

func getCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid course ID", "BAD_REQUEST")
			return
		}
		course, err := models.GetCourseByID(db, id)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Course not found", "NOT_FOUND")
				return
			}
			log.Printf("getCourse: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, course)
	}
}

func updateCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid course ID", "BAD_REQUEST")
			return
		}
		var c models.Course
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}
		if errs := validators.ValidateCourse(c); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}
		updated, err := models.UpdateCourse(db, id, c)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrNotFound):
				writeError(w, http.StatusNotFound, "Course not found", "NOT_FOUND")
			case errors.Is(err, models.ErrDuplicate):
				writeError(w, http.StatusConflict, "Course code already exists", "DUPLICATE_COURSE_CODE")
			default:
				log.Printf("updateCourse: %v", err)
				writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			}
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func deleteCourse(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid course ID", "BAD_REQUEST")
			return
		}
		if err := models.DeleteCourse(db, id); err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Course not found", "NOT_FOUND")
				return
			}
			log.Printf("deleteCourse: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
