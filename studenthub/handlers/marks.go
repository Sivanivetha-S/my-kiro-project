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

// RegisterMarksRoutes registers all marks-related routes on mux.
func RegisterMarksRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.HandleFunc("POST /api/v1/marks", createMarks(db))
	mux.HandleFunc("PUT /api/v1/marks/{id}", updateMarks(db))
	mux.HandleFunc("GET /api/v1/students/{id}/marks", getMarksForStudent(db))
}

func createMarks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Marks
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}

		if errs := validators.ValidateMarks(m); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}

		// Verify the student is enrolled in the course.
		enrolled, err := models.IsEnrolled(db, m.StudentID, m.CourseID)
		if err != nil {
			log.Printf("createMarks IsEnrolled: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		if !enrolled {
			writeError(w, http.StatusBadRequest, "Student is not enrolled in this course", "NOT_ENROLLED")
			return
		}

		created, err := models.CreateMarks(db, m)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrDuplicate):
				writeError(w, http.StatusConflict, "Marks already recorded for this student and course", "DUPLICATE_MARKS")
			default:
				log.Printf("createMarks: %v", err)
				writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			}
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func updateMarks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid marks ID", "BAD_REQUEST")
			return
		}

		var req struct {
			Marks float64 `json:"marks"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}

		if errs := validators.ValidateMarksUpdate(req.Marks); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}

		updated, err := models.UpdateMarks(db, id, req.Marks)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Marks record not found", "NOT_FOUND")
				return
			}
			log.Printf("updateMarks: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func getMarksForStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid student ID", "BAD_REQUEST")
			return
		}

		resp, err := models.GetMarksForStudent(db, id)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Student not found", "NOT_FOUND")
				return
			}
			log.Printf("getMarksForStudent: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}
