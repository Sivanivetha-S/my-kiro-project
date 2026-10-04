# StudentHub

A full-stack Student Management System for managing college student academic information.

Built with a Go REST API backend, SQLite database, and a vanilla HTML/CSS/JavaScript frontend.

## Features

- Student management (add, view, edit, delete, search, filter)
- Course management and student enrollment
- Marks recording with automatic grade calculation
- Attendance tracking with percentage calculation
- Dashboard with key metrics and low-attendance alerts

## How to Run

> The application is not yet fully implemented. Implementation is in progress following the task plan in `.kiro/specs/studenthub/tasks.md`.

Once implementation is complete:

1. Ensure Go 1.22+ and a C compiler (for CGo/SQLite) are installed.
2. From the `studenthub/` directory, install the SQLite driver:
   ```
   go get github.com/mattn/go-sqlite3
   ```
3. Start the server:
   ```
   go run main.go
   ```
4. Open `http://localhost:8080` in a browser.

Configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./studenthub.db` | Path to SQLite database file |
| `ATTENDANCE_THRESHOLD` | `75.0` | Low-attendance alert threshold (%) |

## Project Structure

```
studenthub/       ← Go backend + frontend
.kiro/            ← Kiro specs, steering, hooks, agents, MCP config
studenthub-power/ ← Kiro Power source for StudentHub development
```

## Specification

Full requirements, design, and implementation tasks are in `.kiro/specs/studenthub/`.
