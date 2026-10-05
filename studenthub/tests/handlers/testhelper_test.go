package handlers_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"studenthub/db"
)

// newTestDB creates an in-memory SQLite database with the full schema applied.
// Each call returns a fresh, isolated database.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("newTestDB Open: %v", err)
	}

	// Locate migrations relative to the project root via this file's path.
	_, filename, _, _ := runtime.Caller(0)
	// tests/handlers/ → ../../db/migrations
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations")

	// Only apply the init migration, not seed.sql.
	// We read the directory and apply *.sql files that are not seed.sql.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("newTestDB ReadDir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() == "seed.sql" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, e.Name()))
		if err != nil {
			t.Fatalf("newTestDB ReadFile %s: %v", e.Name(), err)
		}
		if _, err := database.Exec(string(content)); err != nil {
			t.Fatalf("newTestDB Exec %s: %v", e.Name(), err)
		}
	}

	t.Cleanup(func() { database.Close() })
	return database
}
