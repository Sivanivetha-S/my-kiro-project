package models

import "errors"

// ErrNotFound is returned by model functions when a requested record does
// not exist in the database. Handlers translate this to HTTP 404.
var ErrNotFound = errors.New("record not found")

// ErrDuplicate is returned by model functions when a unique-constraint
// violation is detected (duplicate student_id, email, course_code, etc.).
// Handlers translate this to HTTP 409.
var ErrDuplicate = errors.New("duplicate record")
