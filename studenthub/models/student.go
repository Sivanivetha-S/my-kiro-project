package models

import (
	"database/sql"
	"fmt"
	"strings"
)

// Student represents a student record. All fields map directly to the
// `students` table defined in db/migrations/001_init.sql.
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

// StudentFilter holds optional query parameters for GetAllStudents.
// Zero values mean "no filter applied".
type StudentFilter struct {
	Search     string // partial match on full_name, email, or exact student_id
	Department string
	Year       int // 0 = no year filter
}

// CreateStudent inserts a new student record and returns the created student
// including its generated id and created_at timestamp.
//
// Returns ErrDuplicate if the student_id or email is already in use.
func CreateStudent(db *sql.DB, s Student) (Student, error) {
	const query = `
		INSERT INTO students (student_id, full_name, email, phone, department, year, section, dob)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, student_id, full_name, email, phone, department, year, section, dob, created_at`

	phone := sql.NullString{String: s.Phone, Valid: s.Phone != ""}
	row := db.QueryRow(query,
		s.StudentID, s.FullName, s.Email, phone,
		s.Department, s.Year, s.Section, s.DOB,
	)

	var created Student
	var ph sql.NullString
	err := row.Scan(
		&created.ID, &created.StudentID, &created.FullName, &created.Email,
		&ph, &created.Department, &created.Year, &created.Section,
		&created.DOB, &created.CreatedAt,
	)
	if err != nil {
		if isDuplicateError(err) {
			return Student{}, ErrDuplicate
		}
		return Student{}, fmt.Errorf("CreateStudent: %w", err)
	}
	if ph.Valid {
		created.Phone = ph.String
	}
	return created, nil
}

// GetAllStudents returns all students, optionally filtered by search term,
// department, and/or year. All filters are ANDed together.
func GetAllStudents(db *sql.DB, f StudentFilter) ([]Student, error) {
	query := `SELECT id, student_id, full_name, email, phone, department, year, section, dob, created_at
	          FROM students WHERE 1=1`
	var args []any

	if f.Search != "" {
		query += ` AND (full_name LIKE ? OR email LIKE ? OR student_id = ?)`
		like := "%" + f.Search + "%"
		args = append(args, like, like, f.Search)
	}
	if f.Department != "" {
		query += ` AND department = ?`
		args = append(args, f.Department)
	}
	if f.Year > 0 {
		query += ` AND year = ?`
		args = append(args, f.Year)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetAllStudents: %w", err)
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		var ph sql.NullString
		if err := rows.Scan(
			&s.ID, &s.StudentID, &s.FullName, &s.Email,
			&ph, &s.Department, &s.Year, &s.Section,
			&s.DOB, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetAllStudents scan: %w", err)
		}
		if ph.Valid {
			s.Phone = ph.String
		}
		students = append(students, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllStudents rows: %w", err)
	}
	return students, nil
}

// GetStudentByID returns a single student by primary key.
// Returns ErrNotFound if no student with that id exists.
func GetStudentByID(db *sql.DB, id int) (Student, error) {
	const query = `SELECT id, student_id, full_name, email, phone, department, year, section, dob, created_at
	               FROM students WHERE id = ?`

	var s Student
	var ph sql.NullString
	err := db.QueryRow(query, id).Scan(
		&s.ID, &s.StudentID, &s.FullName, &s.Email,
		&ph, &s.Department, &s.Year, &s.Section,
		&s.DOB, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return Student{}, ErrNotFound
	}
	if err != nil {
		return Student{}, fmt.Errorf("GetStudentByID: %w", err)
	}
	if ph.Valid {
		s.Phone = ph.String
	}
	return s, nil
}

// UpdateStudent replaces all mutable fields of a student record.
// Returns ErrNotFound if no student with that id exists.
// Returns ErrDuplicate if the new student_id or email conflicts with another record.
func UpdateStudent(db *sql.DB, id int, s Student) (Student, error) {
	const query = `
		UPDATE students
		SET student_id=?, full_name=?, email=?, phone=?, department=?, year=?, section=?, dob=?
		WHERE id=?`

	phone := sql.NullString{String: s.Phone, Valid: s.Phone != ""}
	result, err := db.Exec(query,
		s.StudentID, s.FullName, s.Email, phone,
		s.Department, s.Year, s.Section, s.DOB, id,
	)
	if err != nil {
		if isDuplicateError(err) {
			return Student{}, ErrDuplicate
		}
		return Student{}, fmt.Errorf("UpdateStudent: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Student{}, fmt.Errorf("UpdateStudent rows affected: %w", err)
	}
	if n == 0 {
		return Student{}, ErrNotFound
	}
	return GetStudentByID(db, id)
}

// DeleteStudent removes a student by primary key.
// Cascade deletes in the DB handle enrollments, marks, and attendance.
// Returns ErrNotFound if no student with that id exists.
func DeleteStudent(db *sql.DB, id int) error {
	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteStudent: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteStudent rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// isDuplicateError reports whether an SQLite error represents a UNIQUE
// constraint violation. We detect this by inspecting the error message
// because mattn/go-sqlite3 does not export a typed error for this case.
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "unique violation")
}
