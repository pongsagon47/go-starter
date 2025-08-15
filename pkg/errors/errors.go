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
	ErrInvalidCredentialsError = New(ErrInvalidCredentials, "Invalid email or password", http.StatusUnauthorized)
	ErrTokenExpiredError       = New(ErrTokenExpired, "Token has expired", http.StatusUnauthorized)
	ErrTokenInvalidError       = New(ErrTokenInvalid, "Invalid token", http.StatusUnauthorized)
)

// =============================================================================
// Convenience Helper Functions
// =============================================================================

// NotFound creates a not found error with custom message
func NotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return New(ErrNotFound, message, http.StatusNotFound)
}

// BadRequest creates a bad request error with custom message
func BadRequest(message string) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return New(ErrBadRequest, message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error with custom message
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return New(ErrUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error with custom message
func Forbidden(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return New(ErrForbidden, message, http.StatusForbidden)
}

// Conflict creates a conflict error with custom message
func Conflict(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return New(ErrConflict, message, http.StatusConflict)
}

// Internal creates an internal server error with custom message
func Internal(message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return New(ErrInternal, message, http.StatusInternalServerError)
}

// Validation creates a validation error with custom message
func Validation(message string) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return New(ErrValidation, message, http.StatusBadRequest)
}

// =============================================================================
// Wrapping Helper Functions
// =============================================================================

// WrapInternal wraps an error as internal server error
func WrapInternal(err error, message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return Wrap(err, ErrInternal, message, http.StatusInternalServerError)
}

// WrapNotFound wraps an error as not found error
func WrapNotFound(err error, message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return Wrap(err, ErrNotFound, message, http.StatusNotFound)
}

// WrapBadRequest wraps an error as bad request error
func WrapBadRequest(err error, message string) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return Wrap(err, ErrBadRequest, message, http.StatusBadRequest)
}

// WrapUnauthorized wraps an error as unauthorized error
func WrapUnauthorized(err error, message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return Wrap(err, ErrUnauthorized, message, http.StatusUnauthorized)
}

// =============================================================================
// Database-specific Helper Functions
// =============================================================================

// DatabaseError creates a database error (usually 500)
func DatabaseError(message string) *AppError {
	if message == "" {
		message = "Database operation failed"
	}
	return New("DATABASE_ERROR", message, http.StatusInternalServerError)
}

// WrapDatabase wraps a database error
func WrapDatabase(err error, message string) *AppError {
	if message == "" {
		message = "Database operation failed"
	}
	return Wrap(err, "DATABASE_ERROR", message, http.StatusInternalServerError)
}

// =============================================================================
// Auth-specific Helper Functions
// =============================================================================

// InvalidCredentials creates invalid credentials error
func InvalidCredentials() *AppError {
	return New(ErrInvalidCredentials, "Invalid email or password", http.StatusUnauthorized)
}

// TokenExpired creates token expired error
func TokenExpired() *AppError {
	return New(ErrTokenExpired, "Token has expired", http.StatusUnauthorized)
}

// TokenInvalid creates invalid token error
func TokenInvalid() *AppError {
	return New(ErrTokenInvalid, "Invalid token", http.StatusUnauthorized)
}

// UserExists creates user already exists error
func UserExists(field string) *AppError {
	message := "User already exists"
	if field != "" {
		message = fmt.Sprintf("%s already exists", field)
	}
	return New(ErrUserExists, message, http.StatusConflict)
}

// UserNotFound creates user not found error
func UserNotFound() *AppError {
	return New(ErrUserNotFound, "User not found", http.StatusNotFound)
}

// TokenError creates token generation/processing error
func TokenError(message string) *AppError {
	if message == "" {
		message = "Token operation failed"
	}
	return New("TOKEN_ERROR", message, http.StatusInternalServerError)
}

// WrapTokenError wraps a token-related error
func WrapTokenError(err error, message string) *AppError {
	if message == "" {
		message = "Token operation failed"
	}
	return Wrap(err, "TOKEN_ERROR", message, http.StatusInternalServerError)
}

// AccountDisabled creates account disabled error
func AccountDisabled() *AppError {
	return New(ErrUnauthorized, "Account is disabled", http.StatusUnauthorized)
}
