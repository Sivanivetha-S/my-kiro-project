package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Open opens a SQLite database at the given path and applies the required
// connection settings.
//
// Two settings are applied immediately after opening, as required by the
// StudentHub architecture (architecture.md):
//
//   - PRAGMA foreign_keys = ON: SQLite does not enforce foreign key
//     constraints by default. This must be set on every new connection
//     because it is not persisted in the database file.
//
//   - SetMaxOpenConns(1): SQLite supports only one concurrent writer.
//     Using a single connection prevents "database is locked" errors under
//     concurrent HTTP requests.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("db.Open: %w", err)
	}

	// Verify the connection is actually usable.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Open ping: %w", err)
	}

	// Enable foreign key enforcement. This pragma is not persisted in the
	// database file and must be set on every new connection.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Open foreign_keys pragma: %w", err)
	}

	// Restrict to a single open connection. SQLite allows multiple concurrent
	// readers but only one writer; a single connection avoids write contention.
	db.SetMaxOpenConns(1)

	return db, nil
}

// RunMigrations reads all *.sql files from migrationsDir in filename order,
// skips any already recorded in the schema_migrations table, and applies the
// rest inside individual transactions. If any migration fails, its transaction
// is rolled back and an error is returned — no further migrations are applied.
//
// Calling RunMigrations a second time on an already-migrated database is safe:
// every previously applied migration is skipped, making the function idempotent.
//
// The schema_migrations table is created inline as part of 001_init.sql.
// RunMigrations itself does not create it — it simply records into it after
// each successful application.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Read all .sql files from the migrations directory.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("RunMigrations: read dir %q: %w", migrationsDir, err)
	}

	// Collect and sort migration filenames. os.ReadDir already returns entries
	// in alphabetical order on most platforms, but we sort explicitly to
	// guarantee consistent application order across all OS implementations.
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, filename := range files {
		applied, err := isMigrationApplied(db, filename)
		if err != nil {
			return fmt.Errorf("RunMigrations: check %q: %w", filename, err)
		}
		if applied {
			continue // already applied on a previous startup — skip
		}

		sqlPath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(sqlPath)
		if err != nil {
			return fmt.Errorf("RunMigrations: read %q: %w", sqlPath, err)
		}

		if err := applyMigration(db, filename, string(content)); err != nil {
			return fmt.Errorf("RunMigrations: apply %q: %w", filename, err)
		}
	}

	return nil
}

// isMigrationApplied checks whether a migration filename has already been
// recorded in schema_migrations. Returns false (not applied) if the
// schema_migrations table does not yet exist, which is the expected state
// before 001_init.sql creates it.
func isMigrationApplied(db *sql.DB, filename string) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE filename = ?",
		filename,
	).Scan(&count)
	if err != nil {
		// schema_migrations does not exist yet — treat as not applied.
		if strings.Contains(err.Error(), "no such table") {
			return false, nil
		}
		return false, fmt.Errorf("isMigrationApplied %q: %w", filename, err)
	}
	return count > 0, nil
}

// applyMigration runs the given SQL content inside a single transaction and,
// on success, records the filename in schema_migrations. Any error causes a
// full rollback — the database is left unchanged.
func applyMigration(db *sql.DB, filename, sqlContent string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Ensure the transaction is rolled back if we return with an error
	// before reaching the Commit call.
	committed := false
	defer func() {
		if !committed {
			tx.Rollback() //nolint:errcheck — best-effort cleanup
		}
	}()

	if _, err := tx.Exec(sqlContent); err != nil {
		return fmt.Errorf("exec SQL: %w", err)
	}

	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (filename) VALUES (?)",
		filename,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	committed = true
	return nil
}
