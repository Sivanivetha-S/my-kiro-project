---
inclusion: always
---

# StudentHub — Coding Standards

This document defines the coding standards for StudentHub. These rules apply to every file in the project. When Kiro generates or modifies code, it must follow these standards without being asked.

---

## Go — Formatting and Naming

### Formatting

- All Go code is formatted with `gofmt`. Do not submit unformatted code.
- Run `go vet ./...` before considering any Go change complete. Zero warnings are required.
- Line length has no hard limit, but prefer wrapping at ~100 characters when it improves readability.

### Naming Conventions

Follow standard Go naming conventions. The rules below are specific to this project.

**Packages**

| Package | Name | Notes |
|---|---|---|
| Entry point | `main` | `main.go` only |
| Configuration | `config` | Exports `Config` struct and `Load() Config` |
| Database | `db` | Exports `Open` and `RunMigrations` |
| Domain models | `models` | Exports structs and query functions |
| HTTP handlers | `handlers` | Exports handler functions and `writeJSON`/`writeError` |
| Validators | `validators` | Exports `ValidateXxx` functions and `ValidationError` |
| Middleware | `middleware` | Exports middleware constructor functions |

**Types and functions**

- Exported types use `PascalCase`: `Student`, `Course`, `Enrollment`, `Marks`, `Attendance`, `Config`, `ValidationError`.
- Exported functions use `PascalCase`: `CreateStudent`, `GetAllStudents`, `ValidateStudent`, `CalculateGrade`.
- Unexported helpers use `camelCase`: `writeJSON`, `writeError`, `parseID`, `buildWhereClause`.
- Acronyms follow Go convention: `ID` not `Id`, `URL` not `Url`, `HTTP` not `Http` — but only at word boundaries. `StudentID` is correct. `studentId` is not used in Go code (it is used in JSON tags).

**Variables**

- Use short, descriptive names. Loop variables may be single letters (`i`, `s`, `c`) when the type is clear from context.
- Avoid single-letter names outside loops: `s` for a handler's `*sql.DB` argument is not acceptable; `db` is.
- Boolean variables and functions that return booleans: prefix with `is`, `has`, or `can` when it makes the intent clear (`isValid`, `hasEnrollment`).

**JSON struct tags**

All exported struct fields that are serialized to JSON must have explicit `json` tags using `snake_case`:

```go
type Student struct {
    ID         int    `json:"id"`
    StudentID  string `json:"student_id"`
    FullName   string `json:"full_name"`
    Email      string `json:"email"`
    Phone      string `json:"phone,omitempty"`
    Department string `json:"department"`
    Year       int    `json:"year"`
    Section    string `json:"section"`
    DOB        string `json:"dob"`
    CreatedAt  string `json:"created_at"`
}
```

Use `omitempty` only for genuinely optional fields (e.g., `Phone`). Do not use `omitempty` on fields that should always be present in the response.

---

## Go — Functions and Modules

### Function size and focus

- Each function does one thing. If a function needs a comment to explain what a section of it does, that section should probably be its own function.
- Handler functions should be short: decode → validate → call model → encode. If a handler is growing complex, the complexity belongs in the model.
- Aim for functions under 40 lines. This is a guideline, not a hard limit — clarity takes precedence over brevity.

### Error handling

Follow these rules exactly. Do not deviate for convenience.

**Sentinel errors for known conditions:**

```go
// models/errors.go
var ErrNotFound  = errors.New("record not found")
var ErrDuplicate = errors.New("duplicate record")
```

Return these directly from model functions. Do not wrap them with `fmt.Errorf("%w", ErrNotFound)` — handlers use `==` comparison, not `errors.Is`, for these sentinels.

**Unexpected errors:**

Wrap unexpected errors with context using `fmt.Errorf`:

```go
rows, err := db.Query(query, args...)
if err != nil {
    return nil, fmt.Errorf("GetAllStudents: %w", err)
}
```

This provides a traceable error chain in server logs without leaking details to the client.

**Never ignore errors:**

```go
// Wrong — silently drops the error
json.NewEncoder(w).Encode(result)

// Correct — log if encoding fails
if err := json.NewEncoder(w).Encode(result); err != nil {
    log.Printf("encode response: %v", err)
}
```

**Handler error mapping** — use this mapping consistently in every handler:

```go
switch {
case errors.Is(err, models.ErrNotFound):
    writeError(w, http.StatusNotFound, "Resource not found", "NOT_FOUND")
case errors.Is(err, models.ErrDuplicate):
    writeError(w, http.StatusConflict, "Record already exists", "DUPLICATE")
default:
    log.Printf("handler error: %v", err)
    writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")
}
```

The full error is written to `log.Printf` on the server. The client receives only the generic message.

### No global state

- Do not declare package-level variables that hold mutable state (e.g., a package-level `*sql.DB`).
- The database connection is created in `main.go` and passed to handlers via closure or dependency injection.
- Configuration is loaded once in `main.go` and passed where needed. Avoid reading environment variables outside `config/config.go`.

### Imports

Group imports in three blocks, separated by blank lines:

```go
import (
    "database/sql"      // 1. Standard library
    "encoding/json"
    "net/http"

    "github.com/mattn/go-sqlite3"  // 2. Third-party

    "studenthub/models"   // 3. Internal packages
    "studenthub/validators"
)
```

Do not use dot imports (`import . "pkg"`) or blank imports except for the SQLite driver side-effect import in `db/db.go`:

```go
import _ "github.com/mattn/go-sqlite3"
```

---

## Go — Validation Rules

- Validators return **all** validation errors at once, not just the first. This allows the UI to highlight all failing fields in a single response.
- Every validator is a pure function with no side effects.
- Validators check format and range rules only. They do not query the database. The model layer handles uniqueness.
- Validation rules are the source of truth from `requirements.md § 7`. If a rule changes in the spec, the validator must be updated to match.
- Validation error `Field` values must match the JSON field names used in the API request body (e.g., `"student_id"`, `"full_name"`, `"dob"`). The frontend uses these field names to display errors on the correct input.

---

## Go — Testing

### General rules

- Use table-driven tests. Define a `[]struct{ name, input, expected }` slice and range over it.
- Each test is independent. Tests must not share state or depend on execution order.
- Use an in-memory SQLite database (`:memory:`) for all tests that require database access.
- Run the same migration script used in production against the in-memory DB to set up schema.
- Test file naming: `xxx_test.go` in the same package or in a `_test` package for black-box tests.

### What to test

Tests are required for:

- All functions in `validators/` — one test case per validation rule, covering both valid and invalid inputs.
- `CalculateGrade` — all grade boundary values explicitly: 0, 49, 50, 59, 60, 69, 70, 79, 80, 89, 90, 100.
- `CalculateAttendancePercentage` — standard cases plus edge cases: total=0, attended=0, attended=total.
- All HTTP handlers — happy path and at least two error paths per handler, using `net/http/httptest`.

Property-based tests are required for:

- `CalculateGrade`: for any `float64` in [0, 100], the result must be one of `{"A+", "A", "B", "C", "D", "F"}` and must never panic.
- `CalculateAttendancePercentage`: for any valid `(total, attended)` pair where `total >= 1` and `0 <= attended <= total`, the result must be in [0.0, 100.0] and must never panic.

Use `testing/quick` from the standard library. Do not use external property-based testing libraries.

### Test naming

Name tests descriptively:

```go
// Good
func TestValidateStudent_InvalidEmail(t *testing.T) {}
func TestCalculateGrade_BoundaryAt80(t *testing.T) {}
func TestCreateStudent_DuplicateStudentID_Returns409(t *testing.T) {}

// Bad
func TestStudent(t *testing.T) {}
func Test1(t *testing.T) {}
```

---

## HTML Standards

### Structure

- Use semantic HTML5 elements: `<nav>`, `<main>`, `<aside>`, `<section>`, `<article>`, `<header>`, `<footer>`, `<table>`, `<form>`, `<button>`. Do not use `<div>` or `<span>` when a semantic element is appropriate.
- The document must have a `<title>` and `lang` attribute on `<html>`.
- Include a skip-to-content link as the first focusable element in `<body>`:

```html
<a href="#app" class="skip-link">Skip to main content</a>
```

### Forms and accessibility

Every form input must have an associated label. Use the `for`/`id` pairing:

```html
<label for="student-email">Email</label>
<input type="email" id="student-email" name="email" required>
```

Do not use `placeholder` as a substitute for a label. Placeholders may be used as supplementary hints only.

Validation error messages are displayed in a dedicated element below the input:

```html
<label for="student-email">Email</label>
<input type="email" id="student-email" name="email" aria-describedby="student-email-error">
<span id="student-email-error" class="field-error" role="alert" hidden></span>
```

- Use `aria-describedby` to associate the error element with the input.
- Use `role="alert"` on the error element so screen readers announce the message when it appears.
- Set `hidden` attribute by default; remove it when showing an error.

Icon-only buttons must have an `aria-label`:

```html
<button type="button" aria-label="Delete student Aisha Rajan">
  <svg>...</svg>
</button>
```

### Interactive elements

- Use `<button type="button">` for actions that do not submit a form.
- Use `<button type="submit">` for form submission buttons.
- Never use `<div>` or `<span>` as a clickable element. Use `<button>` or `<a>`.
- All interactive elements must be reachable and operable via keyboard (Tab to focus, Enter or Space to activate).

---

## CSS Standards

### Design tokens

All colors, spacing, typography, shadows, and border radii are defined as CSS custom properties in the `:root` block in `frontend/css/main.css`. Use these tokens everywhere. Do not hardcode hex values or pixel values in component or page styles.

```css
/* Use tokens */
background-color: var(--color-primary);
border-radius: var(--radius-md);

/* Never do this */
background-color: #2563EB;
border-radius: 8px;
```

The full token set is defined in `design.md § 6.4`.

### Organization

- `main.css` — global reset, `:root` tokens, base typography, layout (sidebar + main).
- `components.css` — reusable patterns: buttons, cards, tables, form inputs, badges, toasts, modals.
- `responsive.css` — media queries only. No new styles here — only overrides to existing classes at breakpoints.

Do not create additional CSS files without a clear reason. Do not write `<style>` blocks inside HTML.

### Specificity and selectors

- Prefer class selectors over element selectors for component styles.
- Avoid ID selectors in CSS (IDs are reserved for JavaScript and `for`/`id` label associations).
- Avoid `!important`. If you need it, the specificity hierarchy is wrong.
- Do not use inline styles in HTML or JavaScript (`element.style.xxx`) except for values that must be computed at runtime and cannot be expressed as a CSS class (e.g., a dynamic width percentage for an attendance bar).

### Responsive layout

The three breakpoints defined in `design.md § 6.5` must be used consistently:

```css
/* Mobile-first base styles — no media query needed */

/* Tablet */
@media (min-width: 640px) { ... }

/* Desktop */
@media (min-width: 1024px) { ... }
```

Every page must be usable at 375px width without horizontal scrolling. Test this before marking any frontend task complete.

---

## JavaScript Standards

### Module structure

- All JavaScript files are ES Modules. Use `import`/`export` syntax.
- `api.js` exports a single `api` object. Every network request goes through it. No other file calls `fetch()`.
- Page modules are the only files that import both `api.js` and component modules. Components do not import `api.js`.
- `utils.js` exports pure functions only. It imports nothing from the project.

### Functions

- Prefer `async`/`await` over raw `.then()` chains for readability.
- Always `await` inside `try`/`catch` when the error must be handled:

```javascript
// Correct
async function loadStudents() {
  try {
    const students = await api.get('/students');
    renderTable(students);
  } catch (err) {
    toast.show('error', 'Failed to load students.');
  }
}

// Wrong — swallows errors silently
api.get('/students').then(renderTable);
```

- Do not use `var`. Use `const` by default; use `let` when reassignment is necessary.
- Prefer arrow functions for callbacks and short utility functions. Use named `function` declarations for top-level page functions, so they appear in stack traces with readable names.

### DOM interaction

- Select elements with `document.getElementById` or `document.querySelector`. Never rely on implicit global variables from element IDs.
- Build dynamic HTML by setting `innerHTML` on a container element, not by repeatedly appending individual nodes. Sanitize any user-provided content before inserting it into `innerHTML` — use `textContent` for user data, not `innerHTML`.

```javascript
// Safe — user data goes through textContent
const td = document.createElement('td');
td.textContent = student.full_name;  // never student.full_name inside innerHTML template

// Acceptable for static structure with no user data embedded
container.innerHTML = `<div class="card"><h2>Dashboard</h2></div>`;
```

- Clean up the `#app` container before rendering a new page. Each page module's render function should start by setting `document.getElementById('app').innerHTML = ''` or by replacing its contents.

### Validation

Client-side validation runs on form `submit` before any API call. The rules must mirror the server-side rules defined in `requirements.md § 7`. If a rule changes in the spec, both the Go validator and the JavaScript validator must be updated.

Validation functions return an array of `{ field, message }` objects:

```javascript
// js/components/form.js
export function displayErrors(errors) {
  clearErrors();
  errors.forEach(({ field, message }) => {
    const el = document.querySelector(`[data-error="${field}"]`);
    if (el) {
      el.textContent = message;
      el.hidden = false;
    }
  });
}
```

Never block submission on a network-round-trip check (e.g., do not pre-check uniqueness via an API call before submitting the form). Submit the form and handle the 409 response from the server.

### Global state

- Minimize global state. Page-level state (e.g., currently selected student ID) lives as a local variable inside the page module, not on `window`.
- Do not attach data to `window` except for the router in `app.js`, which must be globally accessible.
- Do not store sensitive data (even the student records themselves) in `localStorage` or `sessionStorage`.

---

## Security Standards

These rules apply to every file in the project, not just specific layers.

### No hardcoded secrets or credentials

- No passwords, API keys, tokens, or secrets of any kind appear in source code.
- No connection strings, database file paths, or port numbers are hardcoded in source files. They come from environment variables via `config/config.go`.
- The `.gitignore` file must include `studenthub.db` and any future `.env` files.

### SQL injection prevention

All SQL queries use parameterized statements. This rule has no exceptions:

```go
// Correct
row := db.QueryRow("SELECT * FROM students WHERE id = ?", id)

// Never acceptable
row := db.QueryRow("SELECT * FROM students WHERE id = " + id)
row := db.QueryRow(fmt.Sprintf("SELECT * FROM students WHERE id = %d", id))
```

### No internal errors exposed to clients

The server logs full error details internally. The API response contains only a generic message:

```go
// Correct
log.Printf("CreateStudent: %v", err)
writeError(w, http.StatusInternalServerError, "An unexpected error occurred", "INTERNAL_ERROR")

// Never acceptable
writeError(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
```

### Input length caps

All string inputs are validated for maximum length before being written to the database. This prevents both accidental and deliberate oversized input from degrading the system. Maximum lengths are defined in `requirements.md § 7`.

### Out-of-scope security features

Authentication, authorization, HTTPS, rate limiting, and audit logging are out of scope for v1. Do not implement partial versions of these features — a partial implementation gives false confidence without providing real protection. They will be addressed as complete, spec-driven features if StudentHub is extended.

---

## Accessibility Standards

These are the minimum accessibility requirements for every UI change.

**Labels:** Every `<input>`, `<select>`, and `<textarea>` has an associated `<label>` using `for`/`id`. No exceptions.

**Keyboard navigation:** Every interactive element — buttons, links, form inputs, modal close controls — is reachable via Tab and operable via Enter or Space. Test this manually before marking a frontend task complete.

**Focus management:**
- When a modal opens, focus moves to the first focusable element inside it.
- When a modal closes, focus returns to the element that triggered it.
- Focus must not leave the modal while it is open (focus trap).

**Color:** Color is never the only indicator of meaning. Grade badges use both color and a text label. Attendance bars use both color and a percentage value. Low-attendance rows use both a background highlight and a visible percentage.

**Contrast:** All body text and interactive labels meet WCAG 2.1 AA minimum contrast ratio of 4.5:1 against their background. Large text (18pt+) meets the 3:1 minimum. Use the color tokens from `main.css`; they are designed to pass these ratios.

**ARIA:** Use ARIA attributes only when a semantic HTML element is insufficient. Prefer native elements. When ARIA is used, use it correctly — an incorrect `role` or `aria-*` attribute is worse than none.

---

## Responsiveness Standards

- Write mobile-first CSS. Base styles target the smallest screen (375px). Enhancements are added at larger breakpoints.
- Tables that may overflow on mobile must be wrapped in a container with `overflow-x: auto`.
- The sidebar collapses to a hamburger menu below 640px. At no breakpoint should the sidebar overlap or obscure main content without a close control.
- Test every page at 375px, 768px, and 1280px before marking a frontend task complete. No horizontal scrollbar at any of these widths is acceptable.

---

## What Not to Add

Do not introduce the following unless they are explicitly requested and added to the spec:

- User authentication or login functionality
- Role-based access control
- Deployment configuration (Dockerfile, CI/CD pipelines, cloud provider configs)
- A database other than SQLite
- A frontend framework or build tool
- Pagination on the backend API (the frontend handles display limiting in v1)
- Soft delete (all deletes are hard deletes in v1)
- Audit logging
- Email or notification functionality
- File upload
- Report or PDF generation

Adding unrequested features increases complexity and maintenance burden without delivering requested value. If a feature seems obviously useful, raise it as a spec change rather than implementing it silently.
