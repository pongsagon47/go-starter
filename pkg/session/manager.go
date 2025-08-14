package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// SessionManager implements Manager interface
type SessionManager struct {
	store  Store
	config *Config
}

// NewManager creates a new session manager
func NewManager(store Store, config *Config) Manager {
	if config == nil {
		config = DefaultConfig()
	}
	return &SessionManager{
		store:  store,
		config: config,
	}
}

// StartSession creates and starts a new session
func (m *SessionManager) StartSession(ctx context.Context, userID string, data map[string]interface{}) (*Session, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Generate session ID
	sessionID, err := m.generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Create session
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Data:      data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.config.Expiration),
	}

	// Store session
	if err := m.store.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session and validates it
func (m *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, ErrInvalidSessionID
	}

	return m.store.Get(ctx, sessionID)
}

// ValidateSession validates a session and returns session data
func (m *SessionManager) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		m.store.Delete(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	return session, nil
}

// EndSession ends a session
func (m *SessionManager) EndSession(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

// EndAllUserSessions ends all sessions for a user
func (m *SessionManager) EndAllUserSessions(ctx context.Context, userID string) error {
	return m.store.DeleteByUserID(ctx, userID)
}

// SetSessionData sets data in a session
func (m *SessionManager) SetSessionData(ctx context.Context, sessionID string, key string, value interface{}) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Data == nil {
		session.Data = make(map[string]interface{})
	}

	session.Data[key] = value
	return m.store.Update(ctx, session)
}

// GetSessionData gets data from a session
func (m *SessionManager) GetSessionData(ctx context.Context, sessionID string, key string) (interface{}, error) {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, nil
	}

	return session.Data[key], nil
}

// RemoveSessionData removes data from a session
func (m *SessionManager) RemoveSessionData(ctx context.Context, sessionID string, key string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Data != nil {
		delete(session.Data, key)
		return m.store.Update(ctx, session)
	}

	return nil
}

// RefreshSession extends session expiration
func (m *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	return m.store.Refresh(ctx, sessionID, m.config.Expiration)
}

// IsSessionValid checks if a session is valid
func (m *SessionManager) IsSessionValid(ctx context.Context, sessionID string) bool {
	_, err := m.ValidateSession(ctx, sessionID)
	return err == nil
}

// Store interface methods (delegated to store)

func (m *SessionManager) Create(ctx context.Context, session *Session) error {
	return m.store.Create(ctx, session)
}

func (m *SessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
	return m.store.Get(ctx, sessionID)
}

func (m *SessionManager) Update(ctx context.Context, session *Session) error {
	return m.store.Update(ctx, session)
}

func (m *SessionManager) Delete(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

func (m *SessionManager) DeleteByUserID(ctx context.Context, userID string) error {
	return m.store.DeleteByUserID(ctx, userID)
}

func (m *SessionManager) Exists(ctx context.Context, sessionID string) (bool, error) {
	return m.store.Exists(ctx, sessionID)
}

func (m *SessionManager) Refresh(ctx context.Context, sessionID string, duration time.Duration) error {
	return m.store.Refresh(ctx, sessionID, duration)
}

func (m *SessionManager) GetByUserID(ctx context.Context, userID string) ([]*Session, error) {
	return m.store.GetByUserID(ctx, userID)
}

func (m *SessionManager) Cleanup(ctx context.Context) error {
	return m.store.Cleanup(ctx)
}

func (m *SessionManager) Count(ctx context.Context) (int64, error) {
	return m.store.Count(ctx)
}

func (m *SessionManager) CountByUserID(ctx context.Context, userID string) (int64, error) {
	return m.store.CountByUserID(ctx, userID)
}

// Helper methods

// generateSessionID generates a cryptographically secure session ID
func (m *SessionManager) generateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
