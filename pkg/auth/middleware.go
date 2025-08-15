package auth

import (
	"context"
	"net/http"
	"strings"

	"go-starter/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwt               *JWT
	permissionChecker *PermissionChecker
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwt *JWT, permissionChecker *PermissionChecker) *AuthMiddleware {
	return &AuthMiddleware{
		jwt:               jwt,
		permissionChecker: permissionChecker,
	}
}

// RequireAuth middleware that requires valid authentication
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "MISSING_TOKEN", "Authentication token is required", nil)
			c.Abort()
			return
		}

		claims, err := am.jwt.ValidateToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token", nil)
			c.Abort()
			return
		}

		// Add user info to context
		am.setUserContext(c, claims)
		c.Next()
	}
}

// RequirePermission middleware that requires specific permission
func (am *AuthMiddleware) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userPermissions := am.getUserPermissions(c)
		if !am.permissionChecker.HasPermission(userPermissions, requiredPermission) {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS",
				"You don't have permission to access this resource", gin.H{
					"required_permission": requiredPermission,
					"user_permissions":    userPermissions,
				})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (am *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userRoles := am.getUserRoles(c)
		if !am.permissionChecker.HasRole(userRoles, requiredRole) {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_ROLE",
				"You don't have the required role to access this resource", gin.H{
					"required_role": requiredRole,
					"user_roles":    userRoles,
				})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (am *AuthMiddleware) RequireAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		userRoles := am.getUserRoles(c)
		if !am.permissionChecker.HasAnyRole(userRoles, requiredRoles) {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_ROLE",
				"You don't have any of the required roles to access this resource", gin.H{
					"required_roles": requiredRoles,
					"user_roles":     userRoles,
				})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that extracts user info if token is present but doesn't require it
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token != "" {
			claims, err := am.jwt.ValidateToken(token)
			if err == nil {
				am.setUserContext(c, claims)
			}
		}
		c.Next()
	}
}

// AdminOnly middleware shortcut for admin-only endpoints
func (am *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return am.RequireAnyRole(CommonRoles.Admin, CommonRoles.SuperAdmin)
}

// SuperAdminOnly middleware shortcut for super admin only endpoints
func (am *AuthMiddleware) SuperAdminOnly() gin.HandlerFunc {
	return am.RequireRole(CommonRoles.SuperAdmin)
}

// extractToken extracts JWT token from Authorization header or query parameter
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Fallback to query parameter (less secure, use sparingly)
	return c.Query("token")
}

// setUserContext sets user information in the request context
func (am *AuthMiddleware) setUserContext(c *gin.Context, claims *JWTClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("user_email", claims.Email)
	c.Set("user_roles", claims.Roles)
	c.Set("user_permissions", claims.Permissions)
	c.Set("token_type", claims.TokenType)

	// Also set in regular context for use with context-aware functions
	ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "user_email", claims.Email)
	ctx = context.WithValue(ctx, "user_roles", claims.Roles)
	ctx = context.WithValue(ctx, "user_permissions", claims.Permissions)
	c.Request = c.Request.WithContext(ctx)
}

// Helper functions to extract user info from context

// GetCurrentUserID returns the current user ID from context
func GetCurrentUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetCurrentUserEmail returns the current user email from context
func GetCurrentUserEmail(c *gin.Context) string {
	if email, exists := c.Get("user_email"); exists {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}

// getUserRoles returns user roles from context
func (am *AuthMiddleware) getUserRoles(c *gin.Context) []string {
	if roles, exists := c.Get("user_roles"); exists {
		if r, ok := roles.([]string); ok {
			return r
		}
	}
	return []string{}
}

// getUserPermissions returns user permissions from context
func (am *AuthMiddleware) getUserPermissions(c *gin.Context) []string {
	if permissions, exists := c.Get("user_permissions"); exists {
		if p, ok := permissions.([]string); ok {
			return p
		}
	}
	return []string{}
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	return GetCurrentUserID(c) != ""
}

// HasPermission checks if current user has specific permission
func HasPermission(c *gin.Context, permission string) bool {
	if permissions, exists := c.Get("user_permissions"); exists {
		if p, ok := permissions.([]string); ok {
			for _, perm := range p {
				if perm == permission || strings.HasSuffix(perm, "*") {
					if strings.HasSuffix(perm, "*") {
						prefix := strings.TrimSuffix(perm, "*")
						if strings.HasPrefix(permission, prefix) {
							return true
						}
					} else if perm == permission {
						return true
					}
				}
			}
		}
	}
	return false
}

// HasRole checks if current user has specific role
func HasRole(c *gin.Context, role string) bool {
	if roles, exists := c.Get("user_roles"); exists {
		if r, ok := roles.([]string); ok {
			for _, userRole := range r {
				if userRole == role {
					return true
				}
			}
		}
	}
	return false
}

// IsAdmin checks if current user is an admin
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, CommonRoles.Admin) || HasRole(c, CommonRoles.SuperAdmin)
}

// IsSuperAdmin checks if current user is a super admin
func IsSuperAdmin(c *gin.Context) bool {
	return HasRole(c, CommonRoles.SuperAdmin)
}
