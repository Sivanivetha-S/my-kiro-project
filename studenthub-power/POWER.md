---
name: "studenthub-development"
displayName: "StudentHub Development"
description: "Development knowledge for the StudentHub Student Management System — a Go REST API + SQLite backend with a vanilla JS frontend. Provides architecture rules, domain business rules, API conventions, database constraints, and testing requirements specific to this project."
keywords: ["studenthub", "student management", "go backend", "sqlite", "vanilla js", "rest api", "marks", "attendance", "grading", "enrollment", "dashboard", "student", "course"]
author: "Kiro University"
---

# StudentHub Development Power

This Power provides Kiro with project-specific knowledge for the **StudentHub Student Management System** — a Kiro University demonstration project built with a Go REST backend, SQLite database, and vanilla JS frontend.

## Project Location

All StudentHub source code lives under:
```
<workspace>/studenthub/
```

All project specification and steering documents live under:
```
<workspace>/.kiro/specs/studenthub/    ← requirements.md, design.md, tasks.md
<workspace>/.kiro/steering/            ← tech.md, architecture.md, coding-standards.md
```

**Always read the spec and steering documents before implementing any feature.** They are the source of truth. This Power provides a quick-reference summary, not a replacement.

## When to Load Steering Files

Load the appropriate steering file based on what you are doing:

- Implementing business logic (grades, attendance, validation) → `domain-rules.md`
- Writing API handlers, SQL queries, or JSON response shapes → `api-and-db.md`
- Starting a new implementation task, writing tests, or deciding what to build → `dev-workflow.md`

## Project Summary

StudentHub manages four domains: **Students**, **Courses**, **Marks**, and **Attendance**, plus a **Dashboard** that aggregates them.

### Technology (from `.kiro/steering/tech.md`)

| Layer | Technology |
|---|---|
| Frontend | HTML5 + CSS3 + Vanilla JavaScript (ES Modules) |
| Backend | Go 1.22, `net/http` standard library |
| Database | SQLite via `mattn/go-sqlite3` |
| API | REST + JSON, base path `/api/v1` |
| Testing | Go `testing` package + `testing/quick` for property-based tests |

No frontend frameworks. No ORMs. No third-party HTTP routers. No third-party test assertion libraries.

### Backend Package Structure (from `.kiro/steering/architecture.md`)

```
studenthub/
├── main.go           — wire everything, start server
├── config/           — env var config (PORT, DB_PATH, ATTENDANCE_THRESHOLD)
├── db/               — SQLite open + migration runner
├── models/           — domain structs + ALL SQL queries + pure calculations
├── handlers/         — HTTP decode → validate → call model → encode
├── validators/       — pure validation functions, no DB/HTTP
├── middleware/        — CORS only
├── frontend/         — static HTML/CSS/JS served by Go
└── tests/            — tests organised by models/, validators/, handlers/
```

### Key Architectural Rules

1. **No SQL in handlers.** All database interaction lives in `models/`.
2. **No business logic in handlers.** Grade calculation and attendance percentage live in `models/`.
3. **No HTTP in models or validators.** Clean layer separation.
4. **Validators are pure functions.** No side effects, no DB calls.
5. **Computed values are never stored.** Grade and attendance percentage are calculated on every read.
6. **`PRAGMA foreign_keys = ON`** must be set on every SQLite connection.
7. **`SetMaxOpenConns(1)`** must be set — SQLite is single-writer.

## Onboarding Check

When starting a StudentHub development session:

1. Read `.kiro/specs/studenthub/requirements.md` for what the system must do.
2. Read `.kiro/specs/studenthub/design.md` for how it is built (structs, routes, SQL).
3. Read `.kiro/specs/studenthub/tasks.md` for the current implementation plan.
4. Check which task phase is active — do not skip phases.
5. Run `go vet ./...` from `studenthub/` before and after any Go change.
