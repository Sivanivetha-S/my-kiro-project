package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the StudentHub server.
// Values are read from environment variables with documented defaults.
// No value is hardcoded — all configuration is externalised as required
// by coding-standards.md (never hardcode ports, paths, or thresholds).
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	// Read from PORT. Default: "8080".
	Port string

	// DBPath is the file path to the SQLite database.
	// Read from DB_PATH. Default: "./studenthub.db".
	DBPath string

	// AttendanceThreshold is the percentage below which a student is
	// considered to have low attendance and appears in dashboard alerts.
	// Read from ATTENDANCE_THRESHOLD. Default: 75.0.
	AttendanceThreshold float64
}

// Load reads configuration from environment variables and returns a Config
// with defaults applied for any variable that is not set or is invalid.
func Load() Config {
	cfg := Config{
		Port:                "8080",
		DBPath:              "./studenthub.db",
		AttendanceThreshold: 75.0,
	}

	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}

	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}

	if v := os.Getenv("ATTENDANCE_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.AttendanceThreshold = f
		}
		// If the value is set but not parseable, silently keep the default.
		// This prevents an invalid environment variable from crashing startup.
	}

	return cfg
}
