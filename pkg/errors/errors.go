package errors

import (
	"fmt"
	"net/http"
)

// AppError represents application-specific errors
type AppError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	StatusCode int         `json:"-"`
	Details    interface{} `json:"details,omitempty"`
	Cause      error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Error codes
const (
	// General errors
	ErrInternal     = "INTERNAL_ERROR"
	ErrNotFound     = "NOT_FOUND"
	ErrBadRequest   = "BAD_REQUEST"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
	ErrConflict     = "CONFLICT"
	ErrValidation   = "VALIDATION_ERROR"

	// Auth errors
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrTokenExpired       = "TOKEN_EXPIRED"
	ErrTokenInvalid       = "TOKEN_INVALID"
	ErrUserExists         = "USER_EXISTS"
	ErrUserNotFound       = "USER_NOT_FOUND"
)

// New creates a new AppError
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      err,
	}
}

// WithDetails adds details to AppError
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// Predefined errors
var (
	ErrInternalServer    = New(ErrInternal, "Internal server error", http.StatusInternalServerError)
	ErrNotFoundError     = New(ErrNotFound, "Resource not found", http.StatusNotFound)
	ErrBadRequestError   = New(ErrBadRequest, "Bad request", http.StatusBadRequest)
	ErrUnauthorizedError = New(ErrUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbiddenError    = New(ErrForbidden, "Forbidden", http.StatusForbidden)

	// Auth errors
	ErrInvalidCredentialsError = New(ErrInvalidCredentials, "Invalid email or password", http.StatusUnauthorized) // used
	ErrTokenExpiredError       = New(ErrTokenExpired, "Token has expired", http.StatusUnauthorized)
	ErrTokenInvalidError       = New(ErrTokenInvalid, "Invalid token", http.StatusUnauthorized)
)
