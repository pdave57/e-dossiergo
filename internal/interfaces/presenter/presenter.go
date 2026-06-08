// Package presenter handles all HTTP response serialisation for e-Dossier.
package presenter

import (
	"encoding/json"
	"net/http"

	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/pagination"
)

// envelope wraps every API response for consistency.
type envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *errBody `json:"error,omitempty"`
	Meta    *pagination.Meta `json:"meta,omitempty"`
}

type errBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// JSON writes a successful JSON response.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Success: true, Data: data}) //nolint:errcheck
}

// JSONList writes a paginated list response.
func JSONList(w http.ResponseWriter, status int, data any, meta pagination.Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Success: true, Data: data, Meta: &meta}) //nolint:errcheck
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error translates an error to a structured JSON error response.
func Error(w http.ResponseWriter, err error) {
	ae := apperror.AsAppError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.HTTPStatus())
	json.NewEncoder(w).Encode(envelope{ //nolint:errcheck
		Success: false,
		Error: &errBody{
			Code:    string(ae.Code),
			Message: ae.Message,
			Details: ae.Details,
		},
	})
}

// DecodeJSON decodes the request body into v.
// Returns a BadRequest AppError on failure.
func DecodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return apperror.BadRequest("invalid JSON: " + err.Error())
	}
	return nil
}
