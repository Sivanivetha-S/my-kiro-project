package db

import (
	"database/sql"
	"fmt"

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
