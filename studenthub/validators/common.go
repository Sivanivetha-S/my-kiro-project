// Package validators provides pure validation functions for StudentHub.
// Each function accepts a domain struct and returns a slice of ValidationError
// covering ALL invalid fields — never just the first.
// Validators do not query the database; uniqueness is enforced by the models layer.
package validators

// ValidationError describes a single field-level validation failure.
// Field matches the JSON field name used in the API request body so that the
// frontend can map the error to the correct input element.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Departments is the fixed list of accepted department values.
// Both the Go backend and the JavaScript frontend use this exact set.
var Departments = []string{
	"Computer Science",
	"Information Technology",
	"Electronics and Communication",
	"Mechanical Engineering",
	"Civil Engineering",
	"Business Administration",
	"Mathematics",
	"Physics",
}

// isValidDepartment reports whether dept is in the fixed department list.
func isValidDepartment(dept string) bool {
	for _, d := range Departments {
		if d == dept {
			return true
		}
	}
	return false
}
