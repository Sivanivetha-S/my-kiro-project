package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"studenthub/config"
	"studenthub/models"
	"studenthub/validators"
)

// RegisterAttendanceRoutes registers all attendance-related routes on mux.
func RegisterAttendanceRoutes(mux *http.ServeMux, db *sql.DB) {
	// NOTE: /api/v1/attendance/low must be registered BEFORE /api/v1/attendance/{id}
	// so the ServeMux matches "low" as a literal path, not as an id.
	mux.HandleFunc("POST /api/v1/attendance", createAttendance(db))
	mux.HandleFunc("GET /api/v1/attendance/low", getLowAttendance(db))
	mux.HandleFunc("PUT /api/v1/attendance/{id}", updateAttendance(db))
	mux.HandleFunc("GET /api/v1/students/{id}/attendance", getAttendanceForStudent(db))
}

// attendanceThreshold holds the configured threshold; set by RegisterAttendanceRoutes.
// Using a package-level variable avoids passing config deep into every handler closure.
var attendanceThreshold float64 = 75.0

// SetAttendanceThreshold is called during startup to configure the threshold.
func SetAttendanceThreshold(t float64) {
	attendanceThreshold = t
}

func createAttendance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var a models.Attendance
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}

		if errs := validators.ValidateAttendance(a); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}

		enrolled, err := models.IsEnrolled(db, a.StudentID, a.CourseID)
		if err != nil {
			log.Printf("createAttendance IsEnrolled: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		if !enrolled {
			writeError(w, http.StatusBadRequest, "Student is not enrolled in this course", "NOT_ENROLLED")
			return
		}

		created, err := models.CreateAttendance(db, a)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrDuplicate):
				writeError(w, http.StatusConflict, "Attendance already recorded for this student and course", "DUPLICATE_ATTENDANCE")
			default:
				log.Printf("createAttendance: %v", err)
				writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			}
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func updateAttendance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid attendance ID", "BAD_REQUEST")
			return
		}

		var req struct {
			TotalClasses int `json:"total_classes"`
			Attended     int `json:"attended"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
			return
		}

		if errs := validators.ValidateAttendanceUpdate(req.TotalClasses, req.Attended); len(errs) > 0 {
			writeValidationErrors(w, errs)
			return
		}

		updated, err := models.UpdateAttendance(db, id, req.TotalClasses, req.Attended)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Attendance record not found", "NOT_FOUND")
				return
			}
			log.Printf("updateAttendance: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func getLowAttendance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		students, err := models.GetLowAttendanceStudents(db, attendanceThreshold)
		if err != nil {
			log.Printf("getLowAttendance: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, students)
	}
}

func getAttendanceForStudent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r, "id")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid student ID", "BAD_REQUEST")
			return
		}

		resp, err := models.GetAttendanceForStudent(db, id)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				writeError(w, http.StatusNotFound, "Student not found", "NOT_FOUND")
				return
			}
			log.Printf("getAttendanceForStudent: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// RegisterAttendanceRoutesWithConfig registers attendance routes and sets the threshold from config.
func RegisterAttendanceRoutesWithConfig(mux *http.ServeMux, db *sql.DB, cfg config.Config) {
	SetAttendanceThreshold(cfg.AttendanceThreshold)
	RegisterAttendanceRoutes(mux, db)
}
