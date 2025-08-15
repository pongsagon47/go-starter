# 🔒 Session Package

Secure Redis-based session management with user indexing, automatic cleanup, and session limits.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Interfaces](#interfaces)
- [Session Management](#session-management)
- [Configuration](#configuration)
- [Examples](#examples)
- [Security Features](#security-features)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/session"
```

## ⚡ Quick Start

### Basic Session Usage

```go
package main

import (
    "context"
    "time"
    "go-starter/pkg/cache"
    "go-starter/pkg/session"
)

func main() {
    // Setup cache (Redis)
    cache, err := cache.NewCache(&config.RedisConfig{
        Host: "localhost",
        Port: 6379,
    })
    if err != nil {
        panic(err)
    }

    // Create session manager
    config := session.DefaultConfig()
    store := session.NewRedisStore(cache, config)
    manager := session.NewManager(store, config)

    ctx := context.Background()

    // Start a new session
    sessionData := map[string]interface{}{
        "role": "admin",
        "permissions": []string{"read", "write"},
    }

    sess, err := manager.StartSession(ctx, "user123", sessionData)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Session ID: %s\n", sess.ID)

    // Validate session
    validSession, err := manager.ValidateSession(ctx, sess.ID)
    if err != nil {
        // Session invalid or expired
    }

    // End session
    manager.EndSession(ctx, sess.ID)
}
```

## 🔧 Interfaces

### Session Structure

```go
type Session struct {
    ID        string                 `json:"id"`        // Unique session ID
    UserID    string                 `json:"user_id"`   // Associated user ID
    Data      map[string]interface{} `json:"data"`      // Session data
    CreatedAt time.Time              `json:"created_at"` // Creation timestamp
    ExpiresAt time.Time              `json:"expires_at"` // Expiration timestamp
    IPAddress string                 `json:"ip_address,omitempty"` // Client IP
    UserAgent string                 `json:"user_agent,omitempty"` // Client User-Agent
}
```

### Store Interface

```go
type Store interface {
    Create(ctx context.Context, session *Session) error
    Get(ctx context.Context, sessionID string) (*Session, error)
    Update(ctx context.Context, session *Session) error
    Delete(ctx context.Context, sessionID string) error
    DeleteByUserID(ctx context.Context, userID string) error
    Exists(ctx context.Context, sessionID string) (bool, error)
    Refresh(ctx context.Context, sessionID string, duration time.Duration) error
    GetByUserID(ctx context.Context, userID string) ([]*Session, error)
    Cleanup(ctx context.Context) error
    Count(ctx context.Context) (int64, error)
    CountByUserID(ctx context.Context, userID string) (int64, error)
}
```

### Manager Interface

```go
type Manager interface {
    Store // Includes all Store methods

    // High-level session management
    StartSession(ctx context.Context, userID string, data map[string]interface{}) (*Session, error)
    GetSession(ctx context.Context, sessionID string) (*Session, error)
    ValidateSession(ctx context.Context, sessionID string) (*Session, error)
    EndSession(ctx context.Context, sessionID string) error
    EndAllUserSessions(ctx context.Context, userID string) error

    // Session data management
    SetSessionData(ctx context.Context, sessionID string, key string, value interface{}) error
    GetSessionData(ctx context.Context, sessionID string, key string) (interface{}, error)
    RemoveSessionData(ctx context.Context, sessionID string, key string) error

    // Session utilities
    RefreshSession(ctx context.Context, sessionID string) error
    IsSessionValid(ctx context.Context, sessionID string) bool
}
```

## 🎯 Session Management

### Creating Sessions

```go
// Start session with initial data
sessionData := map[string]interface{}{
    "user_id":     123,
    "role":        "admin",
    "permissions": []string{"read", "write", "delete"},
    "login_time":  time.Now(),
}

session, err := manager.StartSession(ctx, "user123", sessionData)
if err != nil {
    return err
}

// Session ID can be used in cookies/headers
fmt.Printf("Session ID: %s", session.ID)
```

### Session Validation

```go
func ValidateRequest(sessionID string) (*Session, error) {
    session, err := manager.ValidateSession(ctx, sessionID)
    if err != nil {
        switch err {
        case session.ErrSessionNotFound:
            return nil, errors.New("session not found")
        case session.ErrSessionExpired:
            return nil, errors.New("session expired")
        default:
            return nil, errors.New("session validation failed")
        }
    }

    return session, nil
}
```

### Session Data Operations

```go
// Set session data
manager.SetSessionData(ctx, sessionID, "last_activity", time.Now())
manager.SetSessionData(ctx, sessionID, "cart_items", []string{"item1", "item2"})

// Get session data
lastActivity, err := manager.GetSessionData(ctx, sessionID, "last_activity")
if err != nil {
    // Handle error
}

cartItems, err := manager.GetSessionData(ctx, sessionID, "cart_items")
if err != nil {
    // Handle error
}

// Remove session data
manager.RemoveSessionData(ctx, sessionID, "cart_items")
```

### Session Lifecycle

```go
// Refresh session (extend expiration)
manager.RefreshSession(ctx, sessionID) // Extends by default duration

// Check if session is valid
if manager.IsSessionValid(ctx, sessionID) {
    // Session is valid and not expired
}

// End specific session
manager.EndSession(ctx, sessionID)

// End all user sessions (useful for security)
manager.EndAllUserSessions(ctx, "user123")
```

## ⚙️ Configuration

### Session Config

```go
type Config struct {
    Expiration         time.Duration // Session expiration (default: 24h)
    KeyPrefix          string        // Redis key prefix (default: "session:")
    MaxSessionsPerUser int           // Max concurrent sessions per user (default: 5)
    CleanupInterval    time.Duration // Cleanup interval (default: 1h)
    CookieName         string        // Cookie name (default: "session_id")
    CookieSecure       bool          // Secure cookie flag (default: true)
    CookieHTTPOnly     bool          // HTTP-only cookie flag (default: true)
    CookieSameSite     string        // SameSite attribute (default: "Strict")
}

// Default configuration
config := session.DefaultConfig()

// Custom configuration
config := &session.Config{
    Expiration:         12 * time.Hour, // 12 hours
    MaxSessionsPerUser: 3,              // Max 3 sessions per user
    CookieSecure:       true,           // HTTPS only
    CookieHTTPOnly:     true,           // No JavaScript access
}
```

## 💡 Examples

### Web Application Integration

```go
// Login handler
func LoginHandler(c *gin.Context) {
    // Authenticate user...
    userID := authenticateUser(username, password)
    if userID == "" {
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }

    // Create session
    sessionData := map[string]interface{}{
        "user_id":    userID,
        "role":       getUserRole(userID),
        "login_time": time.Now(),
        "ip_address": c.ClientIP(),
        "user_agent": c.GetHeader("User-Agent"),
    }

    session, err := sessionManager.StartSession(ctx, userID, sessionData)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create session"})
        return
    }

    // Set cookie
    c.SetCookie(
        "session_id",           // name
        session.ID,             // value
        int(24*time.Hour.Seconds()), // max age
        "/",                    // path
        "",                     // domain
        true,                   // secure
        true,                   // httpOnly
    )

    c.JSON(200, gin.H{"message": "Login successful"})
}
```

### Authentication Middleware

```go
func SessionMiddleware(sessionManager session.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionID, err := c.Cookie("session_id")
        if err != nil {
            c.JSON(401, gin.H{"error": "No session"})
            c.Abort()
            return
        }

        session, err := sessionManager.ValidateSession(ctx, sessionID)
        if err != nil {
            // Clear invalid cookie
            c.SetCookie("session_id", "", -1, "/", "", false, true)
            c.JSON(401, gin.H{"error": "Invalid session"})
            c.Abort()
            return
        }

        // Add session data to context
        c.Set("session", session)
        c.Set("user_id", session.UserID)

        // Refresh session on activity
        sessionManager.RefreshSession(ctx, sessionID)

        c.Next()
    }
}
```

### Logout Handler

```go
func LogoutHandler(c *gin.Context) {
    sessionID, err := c.Cookie("session_id")
    if err != nil {
        c.JSON(200, gin.H{"message": "Already logged out"})
        return
    }

    // End session
    sessionManager.EndSession(ctx, sessionID)

    // Clear cookie
    c.SetCookie("session_id", "", -1, "/", "", false, true)

    c.JSON(200, gin.H{"message": "Logout successful"})
}
```

### Admin: View User Sessions

```go
func GetUserSessionsHandler(c *gin.Context) {
    userID := c.Param("user_id")

    sessions, err := sessionManager.GetByUserID(ctx, userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get sessions"})
        return
    }

    // Return session info (without sensitive data)
    sessionInfo := make([]map[string]interface{}, len(sessions))
    for i, sess := range sessions {
        sessionInfo[i] = map[string]interface{}{
            "id":         sess.ID,
            "created_at": sess.CreatedAt,
            "expires_at": sess.ExpiresAt,
            "ip_address": sess.IPAddress,
            "user_agent": sess.UserAgent,
        }
    }

    c.JSON(200, gin.H{"sessions": sessionInfo})
}
```

### Security: Force Logout All Sessions

```go
func ForceLogoutHandler(c *gin.Context) {
    userID := c.Param("user_id")

    // End all user sessions (security action)
    err := sessionManager.EndAllUserSessions(ctx, userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to logout user"})
        return
    }

    c.JSON(200, gin.H{"message": "All sessions terminated"})
}
```

## 🔒 Security Features

### 1. **Session Limits**

```go
// Automatically enforced - old sessions removed when limit exceeded
config := &session.Config{
    MaxSessionsPerUser: 3, // Only 3 concurrent sessions per user
}
```

### 2. **Automatic Expiration**

```go
// Sessions automatically expire and are cleaned up
config := &session.Config{
    Expiration: 24 * time.Hour, // 24-hour expiration
}
```

### 3. **User Indexing**

```go
// Efficiently track and manage user sessions
sessions, err := manager.GetByUserID(ctx, "user123")
manager.EndAllUserSessions(ctx, "user123") // Security logout
```

### 4. **Secure Session IDs**

```go
// Cryptographically secure session IDs (256-bit)
// Generated using crypto/rand
sessionID := generateSessionID() // Returns 64-character hex string
```

## 🎯 Best Practices

### 1. **Session Security**

```go
// Always use HTTPS in production
config.CookieSecure = true

// Prevent JavaScript access to session cookies
config.CookieHTTPOnly = true

// Use appropriate SameSite policy
config.CookieSameSite = "Strict" // or "Lax" for cross-site requests
```

### 2. **Session Validation**

```go
// Always validate sessions before use
session, err := manager.ValidateSession(ctx, sessionID)
if err != nil {
    // Handle invalid/expired session
    return
}

// Check session data integrity
if userID, ok := session.Data["user_id"].(string); !ok {
    // Session data corrupted
    manager.EndSession(ctx, sessionID)
    return
}
```

### 3. **Activity Tracking**

```go
// Update last activity on each request
manager.SetSessionData(ctx, sessionID, "last_activity", time.Now())

// Refresh session to extend expiration
manager.RefreshSession(ctx, sessionID)
```

### 4. **Cleanup and Maintenance**

```go
// Periodic cleanup (handled automatically by Redis TTL)
// Manual cleanup if needed
manager.Cleanup(ctx)

// Monitor session counts
activeCount, err := manager.Count(ctx)
userSessionCount, err := manager.CountByUserID(ctx, "user123")
```

### 5. **Error Handling**

```go
// Handle specific session errors
switch err {
case session.ErrSessionNotFound:
    // Redirect to login
case session.ErrSessionExpired:
    // Show "session expired" message
case session.ErrMaxSessionsExceeded:
    // Inform user about session limit
default:
    // General error handling
}
```

## 🚨 Error Types

```go
var (
    ErrSessionNotFound      = errors.New("session not found")
    ErrSessionExpired       = errors.New("session expired")
    ErrSessionInvalid       = errors.New("session invalid")
    ErrInvalidSessionID     = errors.New("invalid session ID")
    ErrInvalidUserID        = errors.New("invalid user ID")
    ErrMaxSessionsExceeded  = errors.New("maximum sessions per user exceeded")
    ErrSessionDataCorrupted = errors.New("session data corrupted")
)
```

## 🔗 Related Packages

- [`pkg/cache`](../cache/) - Redis caching system
- [`internal/middleware`](../../internal/middleware/) - Authentication middleware
- [`pkg/secure`](../secure/) - Security utilities

## 📚 Additional Resources

- [Session Security Best Practices](https://owasp.org/www-community/controls/Session_Management_Cheat_Sheet)
- [Redis Security](https://redis.io/topics/security)
- [HTTP Cookies Security](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#security)
