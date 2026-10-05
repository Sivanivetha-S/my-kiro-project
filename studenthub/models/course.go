package models

import (
	"database/sql"
	"fmt"
)

// Course represents a course record.
// All fields map directly to the `courses` table.
type Course struct {
	ID         int    `json:"id"`
	CourseCode string `json:"course_code"`
	CourseName string `json:"course_name"`
	Credits    int    `json:"credits"`
	Department string `json:"department"`
}

// CreateCourse inserts a new course and returns it with the generated id.
// Returns ErrDuplicate if course_code is already in use.
func CreateCourse(db *sql.DB, c Course) (Course, error) {
	const query = `
		INSERT INTO courses (course_code, course_name, credits, department)
		VALUES (?, ?, ?, ?)
		RETURNING id, course_code, course_name, credits, department`

	row := db.QueryRow(query, c.CourseCode, c.CourseName, c.Credits, c.Department)
	var created Course
	if err := row.Scan(&created.ID, &created.CourseCode, &created.CourseName, &created.Credits, &created.Department); err != nil {
		if isDuplicateError(err) {
			return Course{}, ErrDuplicate
		}
		return Course{}, fmt.Errorf("CreateCourse: %w", err)
	}
	return created, nil
}

// GetAllCourses returns every course ordered by course_code.
func GetAllCourses(db *sql.DB) ([]Course, error) {
	const query = `SELECT id, course_code, course_name, credits, department FROM courses ORDER BY course_code`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("GetAllCourses: %w", err)
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.CourseCode, &c.CourseName, &c.Credits, &c.Department); err != nil {
			return nil, fmt.Errorf("GetAllCourses scan: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

// GetCourseByID returns a single course by primary key.
// Returns ErrNotFound if no course with that id exists.
func GetCourseByID(db *sql.DB, id int) (Course, error) {
	const query = `SELECT id, course_code, course_name, credits, department FROM courses WHERE id = ?`
	var c Course
	err := db.QueryRow(query, id).Scan(&c.ID, &c.CourseCode, &c.CourseName, &c.Credits, &c.Department)
	if err == sql.ErrNoRows {
		return Course{}, ErrNotFound
	}
	if err != nil {
		return Course{}, fmt.Errorf("GetCourseByID: %w", err)
	}
	return c, nil
}

// UpdateCourse replaces all fields of a course record.
// Returns ErrNotFound or ErrDuplicate as appropriate.
func UpdateCourse(db *sql.DB, id int, c Course) (Course, error) {
	const query = `UPDATE courses SET course_code=?, course_name=?, credits=?, department=? WHERE id=?`
	result, err := db.Exec(query, c.CourseCode, c.CourseName, c.Credits, c.Department, id)
	if err != nil {
		if isDuplicateError(err) {
			return Course{}, ErrDuplicate
		}
		return Course{}, fmt.Errorf("UpdateCourse: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Course{}, ErrNotFound
	}
	return GetCourseByID(db, id)
}

// DeleteCourse removes a course by primary key.
// Cascade deletes handle enrollments, marks, and attendance.
func DeleteCourse(db *sql.DB, id int) error {
	result, err := db.Exec("DELETE FROM courses WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteCourse: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
