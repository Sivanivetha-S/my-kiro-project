package main

import (
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"studenthub/config"
	"studenthub/db"
	"studenthub/middleware"
)

func main() {
	// 1. Load configuration from environment variables.
	cfg := config.Load()
	log.Printf("StudentHub starting on port %s (DB: %s)", cfg.Port, cfg.DBPath)

	// 2. Open the SQLite database connection.
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// 3. Locate the migrations directory relative to the project root.
	//    runtime.Caller(0) gives the source file path of main.go so we can
	//    resolve migrations/ regardless of the working directory when the
	//    server is started.
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filename)
	migrationsDir := filepath.Join(projectRoot, "db", "migrations")

	// 4. Run database migrations (idempotent — safe to call on every startup).
	if err := db.RunMigrations(database, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("Database migrations applied successfully")

	// 5. Build the router.
	//
	//    apiMux holds all /api/v1/... routes. It is wrapped with the CORS
	//    middleware so every API response carries the CORS headers.
	//    API routes are registered here by later tasks (T-08, T-11, etc.).
	apiMux := http.NewServeMux()

	// Top-level mux routes:
	//   /api/  → CORS middleware wrapping apiMux
	//   /      → placeholder (replaced by T-05 with static file serving)
	mux := http.NewServeMux()

	mux.Handle("/api/", middleware.CORS(apiMux))

	// Placeholder root handler — returns a text response to confirm the server
	// is up. T-05 replaces this with static file serving for the SPA shell.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("StudentHub API running")) //nolint:errcheck
	})

	// 6. Start the HTTP server.
	addr := ":" + cfg.Port
	log.Printf("Listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
