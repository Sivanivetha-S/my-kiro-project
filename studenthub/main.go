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

	// Resolve the frontend directory relative to main.go's source location.
	// This keeps static file serving correct regardless of where `go run` is
	// invoked from.
	frontendDir := filepath.Join(projectRoot, "frontend")
	fs := http.FileServer(http.Dir(frontendDir))

	// Top-level mux routes:
	//   /api/  → CORS middleware wrapping apiMux
	//   /css/  → static CSS files from frontend/css/
	//   /js/   → static JS files from frontend/js/
	//   /      → serves frontend/index.html (SPA shell)
	mux := http.NewServeMux()

	mux.Handle("/api/", middleware.CORS(apiMux))

	// Static asset routes — strip the leading path prefix so the file server
	// looks inside the correct subdirectory of frontend/.
	mux.Handle("/css/", fs)
	mux.Handle("/js/", fs)

	// Root handler: serve index.html for all non-API, non-asset paths.
	// This supports the SPA pattern — any unknown path returns the shell,
	// and client-side routing handles the rest.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
	})

	// 6. Start the HTTP server.
	addr := ":" + cfg.Port
	log.Printf("Listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
