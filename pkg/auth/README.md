# 🔐 Auth Package

Comprehensive authentication and authorization system with JWT tokens, role-based access control (RBAC), password hashing, and middleware for securing API endpoints.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [JWT Authentication](#jwt-authentication)
- [Password Handling](#password-handling)
- [Permissions & Roles](#permissions--roles)
- [Middleware](#middleware)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/auth"
```

## ⚡ Quick Start

### Basic JWT Authentication

```go
package main

import (
    "time"
    "go-starter/pkg/auth"
)

func main() {
    // Initialize JWT
    jwtAuth := auth.NewJWT(
        "your-secret-key",     // Secret key (use strong secret in production)
        24*time.Hour,          // Access token TTL
        7*24*time.Hour,        // Refresh token TTL
        "your-app-name",       // Issuer
    )

    // Generate access token
    accessToken, err := jwtAuth.GenerateAccessToken(
        "user-123",            // User ID
        "user@example.com",    // Email
        []string{"user"},      // Roles
        []string{"user:read"}, // Permissions
    )
    if err != nil {
        panic(err)
    }

    // Validate token
    claims, err := jwtAuth.ValidateToken(accessToken)
    if err != nil {
        panic(err)
    }

    fmt.Printf("User ID: %s\n", claims.UserID)
    fmt.Printf("Email: %s\n", claims.Email)
}
```

## 🔑 JWT Authentication

### **JWT Structure**

```go
type JWTClaims struct {
    UserID      string   `json:"user_id"`
    Email       string   `json:"email"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    TokenType   string   `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}
```

### **Token Generation**

```go
// Access Token (short-lived, contains user data)
accessToken, err := jwtAuth.GenerateAccessToken(
    userID,
    email,
    []string{"admin", "user"},
    []string{"user:create", "user:read", "admin:*"},
)

// Refresh Token (long-lived, for token renewal)
refreshToken, err := jwtAuth.GenerateRefreshToken(userID)
```

### **Token Validation**

```go
// Validate and parse token
claims, err := jwtAuth.ValidateToken(tokenString)
if err != nil {
    // Handle invalid token
    return err
}

// Extract user ID quickly
userID, err := jwtAuth.ExtractUserID(tokenString)

// Get token info without validation (for debugging)
info, err := jwtAuth.GetTokenInfo(tokenString)
```

### **Token Refresh**

```go
// Refresh access token using refresh token
newAccessToken, err := jwtAuth.RefreshAccessToken(
    refreshToken,
    updatedRoles,
    updatedPermissions,
)
```

## 🔒 Password Handling

### **Password Hashing**

```go
// Hash password before storing
hashedPassword, err := auth.HashPassword("userPassword123!")
if err != nil {
    return err
}

// Store hashedPassword in database
user.Password = hashedPassword
```

### **Password Verification**

```go
// Verify password during login
isValid := auth.VerifyPassword("userPassword123!", user.Password)
if !isValid {
    return errors.New("invalid password")
}
```

### **Password Validation**

```go
// Validate password strength
err := auth.ValidatePassword("userPassword123!")
if err != nil {
    // Password doesn't meet requirements
    return err
}

// Requirements:
// - At least 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
// - At least one special character
```

## 👥 Permissions & Roles

### **Permission System**

```go
// Initialize permission checker
permChecker := auth.NewPermissionChecker()

// Add permissions
permChecker.AddPermission(auth.Permission{
    ID:          "user:create",
    Name:        "Create User",
    Description: "Ability to create new users",
    Resource:    "user",
    Action:      "create",
})

// Add roles with permissions
permChecker.AddRole(auth.Role{
    ID:   "admin",
    Name: "Administrator",
    Permissions: []auth.Permission{
        {ID: "user:create"},
        {ID: "user:read"},
        {ID: "user:update"},
        {ID: "user:delete"},
    },
})
```

### **Permission Checking**

```go
// Check specific permission
hasPermission := permChecker.HasPermission(
    []string{"user:read", "post:create"},
    "user:read",
)

// Check role
hasRole := permChecker.HasRole(
    []string{"admin", "user"},
    "admin",
)

// Get all permissions for roles
permissions := permChecker.GetRolePermissions([]string{"admin", "moderator"})
```

### **Permission Format**

```go
// Permissions follow "resource:action" format
"user:create"     // Create users
"user:read"       // Read user data
"user:update"     // Update users
"user:delete"     // Delete users
"admin:*"         // All admin actions (wildcard)

// Validate permission format
err := auth.ValidatePermissionFormat("user:create") // Valid
err := auth.ValidatePermissionFormat("invalid")     // Error
```

### **Common Permissions & Roles**

```go
// Predefined permissions
auth.CommonPermissions.UserRead    // "user:read"
auth.CommonPermissions.UserCreate  // "user:create"
auth.CommonPermissions.AdminList   // "admin:list"
auth.CommonPermissions.SystemLogs  // "system:logs"

// Predefined roles
auth.CommonRoles.SuperAdmin  // "super_admin"
auth.CommonRoles.Admin       // "admin"
auth.CommonRoles.User        // "user"
auth.CommonRoles.Guest       // "guest"
```

## 🛡️ Middleware

### **Setup Middleware**

```go
// Initialize auth middleware
jwtAuth := auth.NewJWT(secretKey, accessTTL, refreshTTL, issuer)
permChecker := auth.NewPermissionChecker()
authMiddleware := auth.NewAuthMiddleware(jwtAuth, permChecker)

// Setup router
router := gin.Default()
```

### **Authentication Middleware**

```go
// Require authentication for all routes
router.Use(authMiddleware.RequireAuth())

// Optional authentication (user info if token present)
router.GET("/public", authMiddleware.OptionalAuth(), publicHandler)

// Protected route
router.GET("/profile", authMiddleware.RequireAuth(), getProfileHandler)
```

### **Permission-Based Middleware**

```go
// Require specific permission
router.POST("/users",
    authMiddleware.RequirePermission("user:create"),
    createUserHandler,
)

// Require specific role
router.GET("/admin/users",
    authMiddleware.RequireRole("admin"),
    adminListUsersHandler,
)

// Require any of specified roles
router.GET("/admin",
    authMiddleware.RequireAnyRole("admin", "super_admin"),
    adminDashboardHandler,
)
```

### **Convenience Middleware**

```go
// Admin only routes
adminRoutes := router.Group("/admin")
adminRoutes.Use(authMiddleware.AdminOnly())
{
    adminRoutes.GET("/dashboard", adminDashboardHandler)
    adminRoutes.GET("/users", adminUsersHandler)
}

// Super admin only routes
superAdminRoutes := router.Group("/super-admin")
superAdminRoutes.Use(authMiddleware.SuperAdminOnly())
{
    superAdminRoutes.GET("/system", systemHandler)
    superAdminRoutes.POST("/config", configHandler)
}
```

## 💡 Examples

### **1. User Registration & Login**

```go
// Registration endpoint
func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request", nil)
        return
    }

    // Validate password strength
    if err := auth.ValidatePassword(req.Password); err != nil {
        response.Error(c, 400, "WEAK_PASSWORD", err.Error(), nil)
        return
    }

    // Hash password
    hashedPassword, err := auth.HashPassword(req.Password)
    if err != nil {
        response.Error(c, 500, "HASH_ERROR", "Failed to hash password", nil)
        return
    }

    // Create user
    user := &User{
        Email:    req.Email,
        Password: hashedPassword,
        Roles:    []string{"user"}, // Default role
    }

    if err := userService.CreateUser(user); err != nil {
        response.Error(c, 500, "CREATE_ERROR", "Failed to create user", nil)
        return
    }

    response.Success(c, 201, "User created successfully", gin.H{
        "user_id": user.ID,
        "email":   user.Email,
    })
}

// Login endpoint
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request", nil)
        return
    }

    // Find user
    user, err := userService.GetByEmail(req.Email)
    if err != nil {
        response.Error(c, 401, "INVALID_CREDENTIALS", "Invalid email or password", nil)
        return
    }

    // Verify password
    if !auth.VerifyPassword(req.Password, user.Password) {
        response.Error(c, 401, "INVALID_CREDENTIALS", "Invalid email or password", nil)
        return
    }

    // Get user permissions
    permissions := permissionService.GetUserPermissions(user.ID)

    // Generate tokens
    accessToken, err := jwtAuth.GenerateAccessToken(
        user.ID, user.Email, user.Roles, permissions,
    )
    if err != nil {
        response.Error(c, 500, "TOKEN_ERROR", "Failed to generate token", nil)
        return
    }

    refreshToken, err := jwtAuth.GenerateRefreshToken(user.ID)
    if err != nil {
        response.Error(c, 500, "TOKEN_ERROR", "Failed to generate refresh token", nil)
        return
    }

    response.Success(c, 200, "Login successful", gin.H{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "expires_in":    24 * 3600, // 24 hours in seconds
        "user": gin.H{
            "id":    user.ID,
            "email": user.Email,
            "roles": user.Roles,
        },
    })
}
```

### **2. Token Refresh Endpoint**

```go
func RefreshTokenHandler(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request", nil)
        return
    }

    // Validate refresh token
    claims, err := jwtAuth.ValidateToken(req.RefreshToken)
    if err != nil {
        response.Error(c, 401, "INVALID_REFRESH_TOKEN", "Invalid refresh token", nil)
        return
    }

    if claims.TokenType != "refresh" {
        response.Error(c, 401, "INVALID_TOKEN_TYPE", "Token is not a refresh token", nil)
        return
    }

    // Get updated user data
    user, err := userService.GetByID(claims.UserID)
    if err != nil {
        response.Error(c, 404, "USER_NOT_FOUND", "User not found", nil)
        return
    }

    // Get current permissions
    permissions := permissionService.GetUserPermissions(user.ID)

    // Generate new access token
    newAccessToken, err := jwtAuth.GenerateAccessToken(
        user.ID, user.Email, user.Roles, permissions,
    )
    if err != nil {
        response.Error(c, 500, "TOKEN_ERROR", "Failed to generate new token", nil)
        return
    }

    response.Success(c, 200, "Token refreshed successfully", gin.H{
        "access_token": newAccessToken,
        "expires_in":   24 * 3600,
    })
}
```

### **3. Protected Endpoints**

```go
// Get current user profile
func GetProfileHandler(c *gin.Context) {
    userID := auth.GetCurrentUserID(c)

    user, err := userService.GetByID(userID)
    if err != nil {
        response.Error(c, 404, "USER_NOT_FOUND", "User not found", nil)
        return
    }

    response.Success(c, 200, "Profile retrieved", user)
}

// Update user (requires permission)
func UpdateUserHandler(c *gin.Context) {
    targetUserID := c.Param("id")
    currentUserID := auth.GetCurrentUserID(c)

    // Users can update their own profile, or admin can update anyone
    if targetUserID != currentUserID && !auth.IsAdmin(c) {
        response.Error(c, 403, "FORBIDDEN", "Cannot update other user's profile", nil)
        return
    }

    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request", nil)
        return
    }

    user, err := userService.UpdateUser(targetUserID, req)
    if err != nil {
        response.Error(c, 500, "UPDATE_ERROR", "Failed to update user", nil)
        return
    }

    response.Success(c, 200, "User updated successfully", user)
}

// Admin only endpoint
func AdminStatsHandler(c *gin.Context) {
    // This handler only runs if user has admin role (enforced by middleware)

    stats := adminService.GetSystemStats()
    response.Success(c, 200, "Stats retrieved", stats)
}
```

### **4. Permission-Based Access Control**

```go
// Create post (requires permission)
func CreatePostHandler(c *gin.Context) {
    // Middleware already checked "post:create" permission

    var req CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request", nil)
        return
    }

    userID := auth.GetCurrentUserID(c)

    post := &Post{
        Title:    req.Title,
        Content:  req.Content,
        AuthorID: userID,
    }

    if err := postService.CreatePost(post); err != nil {
        response.Error(c, 500, "CREATE_ERROR", "Failed to create post", nil)
        return
    }

    response.Success(c, 201, "Post created successfully", post)
}

// Delete post (requires permission + ownership check)
func DeletePostHandler(c *gin.Context) {
    postID := c.Param("id")
    userID := auth.GetCurrentUserID(c)

    post, err := postService.GetByID(postID)
    if err != nil {
        response.Error(c, 404, "POST_NOT_FOUND", "Post not found", nil)
        return
    }

    // Check if user owns the post or has admin permission
    if post.AuthorID != userID && !auth.HasPermission(c, "admin:*") {
        response.Error(c, 403, "FORBIDDEN", "Cannot delete other user's post", nil)
        return
    }

    if err := postService.DeletePost(postID); err != nil {
        response.Error(c, 500, "DELETE_ERROR", "Failed to delete post", nil)
        return
    }

    response.Success(c, 200, "Post deleted successfully", nil)
}
```

### **5. Role Management**

```go
// Assign role to user (admin only)
func AssignRoleHandler(c *gin.Context) {
    userID := c.Param("id")

    var req AssignRoleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, "Invalid request", nil)
        return
    }

    // Validate role exists
    if !isValidRole(req.Role) {
        response.Error(c, 400, "INVALID_ROLE", "Invalid role specified", nil)
        return
    }

    // Super admin role can only be assigned by super admin
    if req.Role == auth.CommonRoles.SuperAdmin && !auth.IsSuperAdmin(c) {
        response.Error(c, 403, "FORBIDDEN", "Only super admin can assign super admin role", nil)
        return
    }

    if err := userService.AssignRole(userID, req.Role); err != nil {
        response.Error(c, 500, "ASSIGN_ERROR", "Failed to assign role", nil)
        return
    }

    response.Success(c, 200, "Role assigned successfully", gin.H{
        "user_id": userID,
        "role":    req.Role,
    })
}
```

## 🎯 Best Practices

### **1. Token Security**

```go
// ✅ DO: Use strong secret keys
secretKey := os.Getenv("JWT_SECRET") // Generate with: openssl rand -base64 32

// ✅ DO: Set appropriate token expiry
accessTokenTTL := 15 * time.Minute  // Short-lived
refreshTokenTTL := 7 * 24 * time.Hour // Longer-lived

// ✅ DO: Include token type in claims
if claims.TokenType != "access" {
    return errors.New("invalid token type")
}

// ❌ DON'T: Use weak secrets or long-lived access tokens
```

### **2. Password Security**

```go
// ✅ DO: Always validate password strength
if err := auth.ValidatePassword(password); err != nil {
    return err
}

// ✅ DO: Hash passwords before storing
hashedPassword, err := auth.HashPassword(password)

// ✅ DO: Use constant-time comparison
isValid := auth.VerifyPassword(plainPassword, hashedPassword)

// ❌ DON'T: Store plain text passwords
// ❌ DON'T: Log passwords in any form
```

### **3. Permission Design**

```go
// ✅ DO: Use consistent permission naming
"user:create"     // resource:action
"post:read"       // resource:action
"admin:*"         // wildcard for admin

// ✅ DO: Implement least privilege principle
userRoles := []string{"user"}  // Start with minimal permissions
adminRoles := []string{"admin", "user"}  // Inherit base permissions

// ✅ DO: Validate permissions
if err := auth.ValidatePermissionFormat(permission); err != nil {
    return err
}
```

### **4. Middleware Usage**

```go
// ✅ DO: Use specific middleware for each endpoint
router.POST("/admin/users",
    authMiddleware.RequirePermission("user:create"),
    createUserHandler,
)

// ✅ DO: Layer middleware appropriately
adminGroup := router.Group("/admin")
adminGroup.Use(authMiddleware.RequireAuth())
adminGroup.Use(authMiddleware.RequireRole("admin"))

// ✅ DO: Handle errors gracefully in middleware
if !auth.IsAuthenticated(c) {
    response.Error(c, 401, "UNAUTHORIZED", "Authentication required", nil)
    return
}
```

### **5. Error Handling**

```go
// ✅ DO: Use specific error codes
response.Error(c, 401, "TOKEN_EXPIRED", "Token has expired", nil)
response.Error(c, 403, "INSUFFICIENT_PERMISSIONS", "Access denied", nil)

// ✅ DO: Log security events
logger.Warn("Failed login attempt",
    zap.String("email", email),
    zap.String("ip", c.ClientIP()),
)

// ❌ DON'T: Expose sensitive information in errors
// ❌ DON'T: Log tokens or passwords
```

### **6. Testing Authentication**

```go
func TestAuthMiddleware(t *testing.T) {
    // Setup test JWT
    jwtAuth := auth.NewJWT("test-secret", time.Hour, time.Hour, "test")

    // Generate test token
    token, err := jwtAuth.GenerateAccessToken(
        "test-user", "test@example.com",
        []string{"user"}, []string{"user:read"},
    )
    require.NoError(t, err)

    // Test protected endpoint
    req := httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
}
```

## 🔧 Utilities

### **Helper Functions**

```go
// Generate secure API key
apiKey, err := auth.GenerateAPIKey()
// Output: "gsk_a1b2c3d4e5f6..."

// Generate random string
randomStr, err := auth.GenerateRandomString(16)

// Validate email format
isValid := auth.IsValidEmail("user@example.com")

// Sanitize user input for logging
sanitized := auth.SanitizeUserInput(map[string]interface{}{
    "email":    "user@example.com",
    "password": "secret123",  // Will be "[REDACTED]"
})
```

### **Context Helpers**

```go
// In handlers, extract user information
userID := auth.GetCurrentUserID(c)
email := auth.GetCurrentUserEmail(c)

// Check authentication status
if !auth.IsAuthenticated(c) {
    response.Error(c, 401, "UNAUTHORIZED", "Login required", nil)
    return
}

// Check permissions
if !auth.HasPermission(c, "user:update") {
    response.Error(c, 403, "FORBIDDEN", "Insufficient permissions", nil)
    return
}

// Check roles
if !auth.IsAdmin(c) {
    response.Error(c, 403, "FORBIDDEN", "Admin access required", nil)
    return
}
```

## 🔗 Related Packages

- [`pkg/response`](../response/) - API response formatting
- [`pkg/validator`](../validator/) - Input validation
- [`pkg/errors`](../errors/) - Error handling
- [`pkg/logger`](../logger/) - Security event logging
- [`config`](../../config/) - JWT configuration

## 📚 Additional Resources

- [JWT Best Practices](https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [bcrypt Documentation](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
