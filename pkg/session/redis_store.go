package session

import (
	"context"
	"fmt"
	"time"

	"go-starter/pkg/cache"
)

// RedisStore implements Store interface using Redis
type RedisStore struct {
	cache  cache.Cache
	config *Config
}

// NewRedisStore creates a new Redis session store
func NewRedisStore(cache cache.Cache, config *Config) Store {
	if config == nil {
		config = DefaultConfig()
	}
	return &RedisStore{
		cache:  cache,
		config: config,
	}
}

// Create creates a new session in Redis
func (r *RedisStore) Create(ctx context.Context, session *Session) error {
	if session.ID == "" {
		return fmt.Errorf("session ID is required")
	}

	// Set creation time if not set
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	// Set expiration if not set
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = session.CreatedAt.Add(r.config.Expiration)
	}

	// Store session data
	sessionKey := r.buildSessionKey(session.ID)
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	if err := r.cache.SetJSON(ctx, sessionKey, session, ttl); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Add to user sessions index
	if session.UserID != "" {
		if err := r.addToUserIndex(ctx, session.UserID, session.ID); err != nil {
			// Clean up session if indexing fails
			r.cache.Del(ctx, sessionKey)
			return fmt.Errorf("failed to index session: %w", err)
		}
	}

	return nil
}

// Get retrieves a session from Redis
func (r *RedisStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := r.buildSessionKey(sessionID)

	var session Session
	err := r.cache.GetJSON(ctx, sessionKey, &session)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		r.Delete(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// Update updates a session in Redis
func (r *RedisStore) Update(ctx context.Context, session *Session) error {
	if session.ID == "" {
		return fmt.Errorf("session ID is required")
	}

	// Check if session exists
	exists, err := r.Exists(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if !exists {
		return ErrSessionNotFound
	}

	// Update session
	sessionKey := r.buildSessionKey(session.ID)
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session expired")
	}

	return r.cache.SetJSON(ctx, sessionKey, session, ttl)
}

// Delete deletes a session from Redis
func (r *RedisStore) Delete(ctx context.Context, sessionID string) error {
	// Get session to find user ID for index cleanup
	session, err := r.Get(ctx, sessionID)
	if err != nil && err != ErrSessionNotFound {
		return fmt.Errorf("failed to get session for deletion: %w", err)
	}

	// Delete session
	sessionKey := r.buildSessionKey(sessionID)
	if err := r.cache.Del(ctx, sessionKey); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user index
	if session != nil && session.UserID != "" {
		if err := r.removeFromUserIndex(ctx, session.UserID, sessionID); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to remove session from user index: %v\n", err)
		}
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *RedisStore) DeleteByUserID(ctx context.Context, userID string) error {
	sessions, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete each session
	sessionKeys := make([]string, len(sessions))
	for i, session := range sessions {
		sessionKeys[i] = r.buildSessionKey(session.ID)
	}

	if len(sessionKeys) > 0 {
		if err := r.cache.Del(ctx, sessionKeys...); err != nil {
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}
	}

	// Clear user index
	userIndexKey := r.buildUserIndexKey(userID)
	return r.cache.Del(ctx, userIndexKey)
}

// Exists checks if a session exists in Redis
func (r *RedisStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	sessionKey := r.buildSessionKey(sessionID)
	count, err := r.cache.Exists(ctx, sessionKey)
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}
	return count > 0, nil
}

// Refresh extends session expiration
func (r *RedisStore) Refresh(ctx context.Context, sessionID string, duration time.Duration) error {
	session, err := r.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session for refresh: %w", err)
	}

	session.ExpiresAt = time.Now().Add(duration)
	return r.Update(ctx, session)
}

// GetByUserID gets all sessions for a user
func (r *RedisStore) GetByUserID(ctx context.Context, userID string) ([]*Session, error) {
	userIndexKey := r.buildUserIndexKey(userID)

	var sessionIDs []string
	err := r.cache.GetJSON(ctx, userIndexKey, &sessionIDs)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to get user session index: %w", err)
	}

	var sessions []*Session
	for _, sessionID := range sessionIDs {
		session, err := r.Get(ctx, sessionID)
		if err != nil {
			if err == ErrSessionNotFound || err == ErrSessionExpired {
				// Remove invalid session from index
				r.removeFromUserIndex(ctx, userID, sessionID)
				continue
			}
			return nil, fmt.Errorf("failed to get session %s: %w", sessionID, err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Cleanup removes expired sessions
func (r *RedisStore) Cleanup(ctx context.Context) error {
	// Redis handles TTL automatically, so this is mostly a no-op
	// In a more complex implementation, you might scan for expired sessions
	// and clean up indexes
	return nil
}

// Count returns the number of active sessions
func (r *RedisStore) Count(ctx context.Context) (int64, error) {
	// This would require scanning Redis keys, which is expensive
	// For now, return 0 and implement proper counting if needed
	return 0, nil
}

// CountByUserID returns the number of active sessions for a user
func (r *RedisStore) CountByUserID(ctx context.Context, userID string) (int64, error) {
	sessions, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int64(len(sessions)), nil
}

// Helper methods

func (r *RedisStore) buildSessionKey(sessionID string) string {
	return r.config.KeyPrefix + sessionID
}

func (r *RedisStore) buildUserIndexKey(userID string) string {
	return r.config.KeyPrefix + "user:" + userID
}

func (r *RedisStore) addToUserIndex(ctx context.Context, userID, sessionID string) error {
	userIndexKey := r.buildUserIndexKey(userID)

	// Get existing session IDs
	var sessionIDs []string
	err := r.cache.GetJSON(ctx, userIndexKey, &sessionIDs)
	if err != nil && err != cache.ErrCacheMiss {
		return fmt.Errorf("failed to get user index: %w", err)
	}

	// Add new session ID
	sessionIDs = append(sessionIDs, sessionID)

	// Enforce max sessions per user
	if r.config.MaxSessionsPerUser > 0 && len(sessionIDs) > r.config.MaxSessionsPerUser {
		// Remove oldest sessions
		oldSessions := sessionIDs[:len(sessionIDs)-r.config.MaxSessionsPerUser]
		for _, oldSessionID := range oldSessions {
			r.Delete(ctx, oldSessionID)
		}
		sessionIDs = sessionIDs[len(sessionIDs)-r.config.MaxSessionsPerUser:]
	}

	// Update index
	return r.cache.SetJSON(ctx, userIndexKey, sessionIDs, r.config.Expiration)
}

func (r *RedisStore) removeFromUserIndex(ctx context.Context, userID, sessionID string) error {
	userIndexKey := r.buildUserIndexKey(userID)

	// Get existing session IDs
	var sessionIDs []string
	err := r.cache.GetJSON(ctx, userIndexKey, &sessionIDs)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return nil // Index doesn't exist, nothing to remove
		}
		return fmt.Errorf("failed to get user index: %w", err)
	}

	// Remove session ID
	var newSessionIDs []string
	for _, id := range sessionIDs {
		if id != sessionID {
			newSessionIDs = append(newSessionIDs, id)
		}
	}

	// Update index
	if len(newSessionIDs) == 0 {
		// Remove index if empty
		return r.cache.Del(ctx, userIndexKey)
	}

	return r.cache.SetJSON(ctx, userIndexKey, newSessionIDs, r.config.Expiration)
}
