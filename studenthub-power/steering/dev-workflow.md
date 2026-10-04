# StudentHub — Development Workflow

**Source of truth:** `.kiro/specs/studenthub/tasks.md`, `.kiro/steering/coding-standards.md`
**Load this file when:** starting a new implementation task, writing tests, making architectural decisions, or deciding whether a feature is in scope.

---

## Implementation Phase Order

Tasks must be completed in phase order. Do not start Phase 2 until all Phase 1 tasks pass. Each task in `tasks.md` lists its dependencies explicitly.

| Phase | Tasks | Focus |
|---|---|---|
| 0 | T-01 – T-05 | Project scaffold, DB, config, server bootstrap |
| 1 | T-06 – T-09 | Student CRUD backend + tests |
| 2 | T-10 – T-11 | Course CRUD backend + tests |
| 3 | T-12 – T-17 | Enrollment, Marks, Attendance backend + all unit tests |
| 4 | T-18 – T-25 | Full frontend — all 6 pages, forms, validation |
| 5 | T-26 – T-30 | Accessibility audit, steering docs, hooks, final verification |

**Current state (after Lesson 4):** Phase 0 is partially complete. `go.mod`, `models/marks.go`, and `models/attendance.go` exist. The remaining Phase 0 scaffold files (`main.go`, `config/`, `db/`, `middleware/`) are not yet created.

---

## Before Writing Any Code

1. Read the relevant task entry in `tasks.md` — it lists steps, acceptance criteria, and dependencies.
2. Read the spec section cited by that task — do not rely on memory or this Power alone.
3. Check which files already exist to avoid overwriting prior work.
4. For Go files: run `go vet ./...` after every save (the `studenthub-go-quality` hook does this automatically if Kiro hooks are active).

---

## Go Coding Rules (Summary)

Source: `.kiro/steering/coding-standards.md` — read the full document for complete rules.

### Naming
- Exported types and functions: `PascalCase` — `Student`, `CalculateGrade`, `ValidateStudent`
- Unexported helpers: `camelCase` — `writeJSON`, `parseID`
- Acronyms: `ID` not `Id`, `URL` not `Url`
- JSON struct tags: `snake_case` — `json:"student_id"`, `json:"full_name"`

### Imports — three groups, blank line between each
```go
import (
    "database/sql"      // 1. Standard library
    "encoding/json"

    "github.com/mattn/go-sqlite3"  // 2. Third-party (only this one in v1)

    "studenthub/models"   // 3. Internal packages
    "studenthub/validators"
)
```

### Error handling pattern
```go
// Model functions return sentinel errors for known conditions
var ErrNotFound  = errors.New("record not found")   // models/errors.go
var ErrDuplicate = errors.New("duplicate record")

// Unexpected errors are wrapped with context for server logs
return nil, fmt.Errorf("CreateStudent: %w", err)

// Never ignore errors
if err := json.NewEncoder(w).Encode(result); err != nil {
    log.Printf("encode response: %v", err)
}
```

### Function focus
- Handlers: decode → validate → call model → encode. Nothing else.
- Models: SQL only. No HTTP, no validation logic.
- Validators: pure functions. No DB, no HTTP, no side effects.

---

## Testing Requirements

Source: `coding-standards.md`, `tasks.md T-17`

### What must be tested

| Target | Test type | Location |
|---|---|---|
| `CalculateGrade` | Table-driven + property-based | `tests/models/marks_test.go` ✅ exists |
| `CalculateAttendancePercentage` | Table-driven + property-based | `tests/models/attendance_test.go` ✅ exists |
| All validators | Table-driven unit tests | `tests/validators/` |
| All HTTP handlers | Integration tests with `httptest` | `tests/handlers/` |

### Property-based testing rules (testing/quick)
- Use `testing/quick` from the standard library — no external libraries.
- Use `t.Errorf`/`t.Fatalf` for assertions — no testify or gomega.
- Test names must be descriptive: `TestCalculateGrade_BoundaryAt80`, not `Test1`.
- Use table-driven tests: define `[]struct{ name, input, expected }`, range over it.
- Each test is independent. Tests must not share state or depend on execution order.
- Use in-memory SQLite (`:memory:`) for any test that requires a database.

### Run tests
```powershell
# From studenthub/ directory
go test ./tests/models/...          # property-based tests (already implemented)
go test ./tests/validators/...      # validator unit tests
go test ./tests/handlers/...        # handler integration tests
go test ./...                       # all tests
go test ./... -v -race -cover       # with verbosity, race detector, coverage
```

---

## Frontend Development Rules (Summary)

Source: `.kiro/steering/coding-standards.md`

- All `fetch()` calls go through `frontend/js/api.js` only.
- No inline styles in JavaScript — use CSS classes.
- Every `<input>` must have a `<label for="...">`. No exceptions.
- Error messages go in `<span data-error="field-name">` elements with `role="alert"`.
- Mobile-first CSS. Test at 375px, 768px, 1280px before marking a task done.
- Use `const` by default, `let` when reassignment is needed. Never `var`.
- Use `async`/`await` inside `try/catch` — never silent `.then()` chains.

---

## What Is Explicitly Out of Scope

Do not implement these features unless a new spec task is created and explicitly approved. Implementing them silently adds complexity without delivering requested value.

| Feature | Status |
|---|---|
| User authentication / login | Out of scope v1 |
| Role-based access control | Out of scope v1 |
| Multiple assessments per course | Out of scope v1 — one final marks record only |
| Pagination on the backend API | Out of scope v1 |
| Soft delete | Out of scope v1 — hard deletes only |
| File upload (student photos) | Out of scope v1 |
| Email notifications | Out of scope v1 |
| PDF/report generation | Out of scope v1 |
| Docker / deployment config | Out of scope v1 |
| PostgreSQL / MySQL | Not approved — SQLite only |
| Frontend framework (React/Vue/etc.) | Not approved — vanilla JS only |
| ORM (GORM/sqlx) | Not approved — raw SQL only |
| Third-party Go HTTP router | Not approved — net/http ServeMux only |
| Third-party test assertion library | Not approved — standard testing package only |

If a feature seems obviously useful, raise it as a spec change rather than implementing it quietly.

---

## Acceptance Criteria Checklist (Key Items)

Before marking any backend task complete, verify:
- [ ] `go build ./...` succeeds with zero errors
- [ ] `go vet ./...` produces zero warnings
- [ ] `go test ./...` passes with zero failures
- [ ] All SQL uses `?` parameterised placeholders
- [ ] No SQL exists outside the `models/` package
- [ ] No business logic exists in handlers
- [ ] Cascade deletes verified: deleting a student removes their marks and attendance

Before marking any frontend task complete, verify:
- [ ] No horizontal scrollbar at 375px viewport width
- [ ] Every `<input>` has an associated `<label for="...">`
- [ ] All buttons are keyboard-focusable and operable via Enter/Space
- [ ] Success and error states are shown to the user (no silent failures)
- [ ] No raw error strings from the server are displayed to the user

---

## MCP Integration — studenthub-sqlite

StudentHub uses the **`studenthub-sqlite`** MCP server as a development-time tool configured in `mcp.json` at the repository root.

### What it provides

Direct structured access to the StudentHub SQLite database (`studenthub/studenthub.db`) via six tools:

| Tool | What it does |
|---|---|
| `list_tables` | List all tables currently in the database |
| `describe-table` | Show column definitions and types for a specific table |
| `read_query` | Execute a SELECT query and return results |
| `write_query` | Execute INSERT, UPDATE, or DELETE queries |
| `create_table` | Execute a CREATE TABLE statement |
| `append_insight` | Add a note to the in-memory business insights memo |

### When to use it during development

- **After running migrations (T-03):** Use `list_tables` and `describe-table` to confirm the schema was created correctly.
- **After implementing a model function (T-06 through T-14):** Use `read_query` to verify records are actually persisted with the correct values.
- **When debugging a test failure:** Use `read_query` to inspect the actual state of an in-memory or file DB without adding temporary log statements.
- **When verifying cascade deletes:** After deleting a student via the API, use `read_query` to confirm their marks and attendance rows are gone.

### Important constraints

- **The MCP server is a development-time tool only.** It is not part of the production application architecture. It runs locally and must not be used in any deployment pipeline.
- **Do not use `write_query` or `create_table` to bypass application validation or business rules.** All student/course/marks/attendance data must be created through the REST API so that validators run. Direct DB writes produce data that has never been validated and may violate application invariants.
- **Do not perform destructive operations** (`DROP TABLE`, `DELETE FROM students`, etc.) unless explicitly requested and with full understanding of the consequences.
- **The database does not exist until the application is first started** and `db.RunMigrations()` runs. `list_tables` will return an empty result (or an error) until then — this is expected during Phase 0.

### Configuration

The MCP server is configured in `.kiro/settings/mcp.json` at the workspace level:

```json
{
  "mcpServers": {
    "studenthub-sqlite": {
      "command": "C:\\Users\\HP\\AppData\\Local\\Python\\pythoncore-3.14-64\\Scripts\\mcp-server-sqlite.exe",
      "args": [
        "--db-path",
        "D:\\AWS projects\\kiro-university\\my-kiro-project\\studenthub\\studenthub.db"
      ],
      "disabled": false
    }
  }
}
```

The `command` uses the full absolute path to `mcp-server-sqlite.exe` and `args` uses a full absolute path to the database file. `${workspaceFolder}` variable substitution is not supported in MCP `args` — literal paths are required.

On a different machine, install the server with `py -m pip install mcp-server-sqlite==2025.4.25` and update both the `command` path and the `--db-path` value to match the local installation.
