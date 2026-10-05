package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"studenthub/config"
	"studenthub/models"
)

// RegisterDashboardRoutes registers the dashboard endpoint on mux.
func RegisterDashboardRoutes(mux *http.ServeMux, db *sql.DB, cfg config.Config) {
	mux.HandleFunc("GET /api/v1/dashboard", getDashboard(db, cfg.AttendanceThreshold))
}

// RegisterDashboardFromDB registers the dashboard route with an explicit threshold.
// Used in tests to avoid importing config.
func RegisterDashboardFromDB(mux *http.ServeMux, db *sql.DB, threshold float64) {
	mux.HandleFunc("GET /api/v1/dashboard", getDashboard(db, threshold))
}

func getDashboard(db *sql.DB, threshold float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := models.GetDashboard(db, threshold)
		if err != nil {
			log.Printf("getDashboard: %v", err)
			writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}
