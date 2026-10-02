---
inclusion: always
---

# StudentHub — Approved Technology Stack

This document defines the approved technology choices for StudentHub. All implementation work must stay within these boundaries. Introducing a technology not listed here requires an explicit decision and an update to this document first.

---

## Backend

**Language: Go**

- Use the version specified in `go.mod`. Do not upgrade the Go version without reviewing compatibility.
- Use only the Go standard library for HTTP serving (`net/http`), JSON encoding (`encoding/json`), SQL access (`database/sql`), testing (`testing`, `testing/quick`), and configuration (`os.Getenv`).
- Go 1.22 or later is required. The router uses `net/http.ServeMux` with method-pattern routing (`GET /api/v1/students`, `POST /api/v1/students`), which requires Go 1.22.
- Do not introduce a third-party HTTP router or web framework (e.g., Gin, Echo, Chi, Fiber). The standard library `ServeMux` is sufficient for this project's route count and complexity.

**Database: SQLite via `mattn/go-sqlite3`**

- `github.com/mattn/go-sqlite3` is the only approved database driver.
- No ORM. All database interaction uses raw SQL through `database/sql` with parameterized statements (`?` placeholders).
- Do not introduce GORM, sqlx, sqlc, or any query-builder library.
- The database file is `studenthub.db` stored at the path configured in `DB_PATH`. It must be excluded from version control via `.gitignore`.

**No other external Go dependencies are approved for v1.**

The only entry in `go.mod` beyond the standard library should be `github.com/mattn/go-sqlite3`. If a new dependency is genuinely needed, discuss and document the decision before adding it.

---

## Frontend

**Languages: HTML5, CSS3, Vanilla JavaScript (ES Modules)**

- No JavaScript frameworks. React, Vue, Angular, Svelte, and similar libraries are not approved for v1.
- No CSS frameworks. Tailwind, Bootstrap, and similar libraries are not approved. All styles are written in plain CSS using the design system defined in `design.md § 6.4`.
- No build tools, bundlers, or transpilers (no Webpack, Vite, Babel, or TypeScript). The frontend runs directly in the browser without a build step.
- JavaScript is written as ES Modules (`type="module"` on the script tag). Use `import`/`export` syntax between files.
- No `npm`, `package.json`, or `node_modules` directory in the frontend. There are no frontend dependencies to install.

**Browser target:** Modern evergreen browsers (Chrome, Firefox, Safari, Edge — latest two major versions). Internet Explorer is not a target.

---

## Communication

**REST API with JSON**

- The frontend communicates with the backend exclusively through the REST API defined in `design.md § 4`.
- All API request and response bodies use `application/json`.
- The API base path is `/api/v1`. Do not add new top-level paths outside this prefix without updating the spec.
- WebSockets, GraphQL, gRPC, and server-sent events are not approved for v1.
- The Go server serves both the API and the static frontend files. No separate static file server or CDN is needed.

---

## Testing

**Go standard library testing only**

- Backend tests use the `testing` package exclusively.
- Property-based tests use `testing/quick` from the standard library. Do not introduce `gopter`, `rapid`, or other property-based testing libraries.
- Handler integration tests use `net/http/httptest` from the standard library.
- No third-party test assertion libraries (no testify, gomega, or similar). Use plain `t.Errorf` and `t.Fatalf`.
- All tests run with `go test ./...` from the project root.

**Frontend testing**

- Client-side validation functions are testable as pure functions. A lightweight in-browser test runner may be used for frontend unit tests, but no npm-based test framework (Jest, Vitest, Mocha) is approved without an explicit decision.

---

## Configuration and Secrets

- Runtime configuration is read from environment variables: `PORT`, `DB_PATH`, `ATTENDANCE_THRESHOLD`.
- Default values are defined in `config/config.go` and apply when the environment variable is not set.
- No `.env` files, no secrets managers, no configuration files checked into source control.
- Never hardcode ports, file paths, thresholds, or any value that might differ between environments.

---

## Development Environment

- The Go server is started with `go run main.go` from the `studenthub/` directory.
- No Docker, Docker Compose, or containerization is required for local development.
- No Makefile is required, though one may be added as a convenience if the team finds it useful.
- CGo is required by `mattn/go-sqlite3`. A C compiler (gcc on Linux/macOS, MinGW or TDM-GCC on Windows) must be available on the developer's machine.

---

## Out of Scope — Technology (v1)

The following technologies are explicitly excluded from v1. Do not introduce them:

| Technology | Reason excluded |
|---|---|
| PostgreSQL, MySQL | SQLite is sufficient; multi-server DB adds unnecessary complexity |
| Redis, Memcached | No caching layer needed at this scale |
| Authentication libraries (JWT, OAuth) | Auth is out of scope for v1 |
| Frontend frameworks (React, Vue, etc.) | Vanilla JS keeps the project simple and learnable |
| ORM (GORM, sqlx) | Raw SQL is readable and avoids magic at this scale |
| Third-party Go routers | Standard `ServeMux` is adequate |
| Message queues, event buses | No async processing needed |
| Cloud SDKs | Local SQLite only in v1; cloud deployment is a future concern |

If a future version (v2) introduces cloud deployment via AWS Blocks, that will be documented as a separate spec.
