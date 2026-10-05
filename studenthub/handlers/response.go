package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"studenthub/validators"
)

// errorBody is the JSON shape returned for all API errors.
type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// validationErrorBody is returned for HTTP 400 validation failures.
// It includes the per-field errors so the frontend can display them inline.
type validationErrorBody struct {
	Error  string                       `json:"error"`
	Code   string                       `json:"code"`
	Fields []validators.ValidationError `json:"fields"`
}

// writeJSON encodes v as JSON and writes it with the given status code.
// Encoding errors are logged but not surfaced to the client.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, msg, code string) {
	writeJSON(w, status, errorBody{Error: msg, Code: code})
}

// writeValidationErrors writes a 400 response containing all field-level errors.
func writeValidationErrors(w http.ResponseWriter, errs []validators.ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	body := validationErrorBody{
		Error:  "Validation failed",
		Code:   "VALIDATION_ERROR",
		Fields: errs,
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("writeValidationErrors encode error: %v", err)
	}
}
