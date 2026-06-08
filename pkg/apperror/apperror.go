// Package apperror defines typed application errors used across all layers.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Code represents a well-known error classification.
type Code string

const (
	CodeNotFound       Code = "NOT_FOUND"
	CodeConflict       Code = "CONFLICT"
	CodeValidation     Code = "VALIDATION_ERROR"
	CodeUnauthorized   Code = "UNAUTHORIZED"
	CodeForbidden      Code = "FORBIDDEN"
	CodeInternal       Code = "INTERNAL_ERROR"
	CodeBadRequest     Code = "BAD_REQUEST"
	CodeUnprocessable  Code = "UNPROCESSABLE"
)

// AppError is the canonical error type for e-Dossier.
type AppError struct {
	Code    Code           `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	Cause   error          `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// HTTPStatus maps an AppError code to an HTTP status code.
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeValidation, CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// ─── constructors ────────────────────────────────────────────────────────────

func New(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

func NotFound(entity, id string) *AppError {
	return &AppError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s with id %q not found", entity, id),
	}
}

func Conflict(message string) *AppError {
	return &AppError{Code: CodeConflict, Message: message}
}

func Validation(fields map[string]any) *AppError {
	return &AppError{
		Code:    CodeValidation,
		Message: "request validation failed",
		Details: fields,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: CodeForbidden, Message: message}
}

func Internal(cause error) *AppError {
	return &AppError{Code: CodeInternal, Message: "an internal error occurred", Cause: cause}
}

func BadRequest(message string) *AppError {
	return &AppError{Code: CodeBadRequest, Message: message}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// Is reports whether any error in the chain is an *AppError with the given code.
func Is(err error, code Code) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code == code
	}
	return false
}

// AsAppError extracts an *AppError from an error chain, or wraps it as Internal.
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return Internal(err)
}
