package session

import "errors"

// Session-related errors
var (
	// ErrSessionNotFound indicates that a session was not found
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired indicates that a session has expired
	ErrSessionExpired = errors.New("session expired")

	// ErrSessionInvalid indicates that a session is invalid
	ErrSessionInvalid = errors.New("session invalid")

	// ErrInvalidSessionID indicates that session ID is invalid
	ErrInvalidSessionID = errors.New("invalid session ID")

	// ErrInvalidUserID indicates that user ID is invalid
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrMaxSessionsExceeded indicates that maximum sessions per user exceeded
	ErrMaxSessionsExceeded = errors.New("maximum sessions per user exceeded")

	// ErrSessionDataCorrupted indicates that session data is corrupted
	ErrSessionDataCorrupted = errors.New("session data corrupted")
)
