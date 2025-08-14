package session

import (
	"context"
	"time"
)

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// Store defines the interface for session storage
type Store interface {
	// Create creates a new session
	Create(ctx context.Context, session *Session) error

	// Get retrieves a session by ID
	Get(ctx context.Context, sessionID string) (*Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *Session) error

	// Delete deletes a session by ID
	Delete(ctx context.Context, sessionID string) error

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error

	// Exists checks if a session exists
	Exists(ctx context.Context, sessionID string) (bool, error)

	// Refresh extends session expiration
	Refresh(ctx context.Context, sessionID string, duration time.Duration) error

	// GetByUserID gets all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*Session, error)

	// Cleanup removes expired sessions
	Cleanup(ctx context.Context) error

	// Count returns the number of active sessions
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the number of active sessions for a user
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

// Manager provides high-level session management
type Manager interface {
	Store

	// StartSession creates and starts a new session
	StartSession(ctx context.Context, userID string, data map[string]interface{}) (*Session, error)

	// GetSession retrieves a session and validates it
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// ValidateSession validates a session and returns session data
	ValidateSession(ctx context.Context, sessionID string) (*Session, error)

	// EndSession ends a session
	EndSession(ctx context.Context, sessionID string) error

	// EndAllUserSessions ends all sessions for a user
	EndAllUserSessions(ctx context.Context, userID string) error

	// SetSessionData sets data in a session
	SetSessionData(ctx context.Context, sessionID string, key string, value interface{}) error

	// GetSessionData gets data from a session
	GetSessionData(ctx context.Context, sessionID string, key string) (interface{}, error)

	// RemoveSessionData removes data from a session
	RemoveSessionData(ctx context.Context, sessionID string, key string) error

	// RefreshSession extends session expiration
	RefreshSession(ctx context.Context, sessionID string) error

	// IsSessionValid checks if a session is valid
	IsSessionValid(ctx context.Context, sessionID string) bool
}

// Config holds session configuration
type Config struct {
	// Session expiration duration
	Expiration time.Duration
	// Key prefix for session storage
	KeyPrefix string
	// Maximum sessions per user (0 = unlimited)
	MaxSessionsPerUser int
	// Cleanup interval for expired sessions
	CleanupInterval time.Duration
	// Cookie settings
	CookieName     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
}

// DefaultConfig returns default session configuration
func DefaultConfig() *Config {
	return &Config{
		Expiration:         24 * time.Hour, // 24 hours
		KeyPrefix:          "session:",
		MaxSessionsPerUser: 5, // Maximum 5 concurrent sessions per user
		CleanupInterval:    1 * time.Hour,
		CookieName:         "session_id",
		CookieSecure:       true,
		CookieHTTPOnly:     true,
		CookieSameSite:     "Strict",
	}
}
