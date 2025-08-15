package auth

import (
	"fmt"
	"strings"
)

// Permission represents a system permission
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// Role represents a user role with permissions
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// PermissionChecker handles permission validation
type PermissionChecker struct {
	permissions map[string]Permission
	roles       map[string]Role
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{
		permissions: make(map[string]Permission),
		roles:       make(map[string]Role),
	}
}

// AddPermission adds a permission to the checker
func (pc *PermissionChecker) AddPermission(permission Permission) {
	pc.permissions[permission.ID] = permission
}

// AddRole adds a role to the checker
func (pc *PermissionChecker) AddRole(role Role) {
	pc.roles[role.ID] = role
}

// HasPermission checks if the given permissions include the required permission
func (pc *PermissionChecker) HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}

		// Check for wildcard permissions
		if strings.HasSuffix(permission, "*") {
			prefix := strings.TrimSuffix(permission, "*")
			if strings.HasPrefix(requiredPermission, prefix) {
				return true
			}
		}
	}
	return false
}

// HasRole checks if the given roles include the required role
func (pc *PermissionChecker) HasRole(userRoles []string, requiredRole string) bool {
	for _, role := range userRoles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the required roles
func (pc *PermissionChecker) HasAnyRole(userRoles []string, requiredRoles []string) bool {
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// HasAllRoles checks if the user has all required roles
func (pc *PermissionChecker) HasAllRoles(userRoles []string, requiredRoles []string) bool {
	userRoleMap := make(map[string]bool)
	for _, role := range userRoles {
		userRoleMap[role] = true
	}

	for _, requiredRole := range requiredRoles {
		if !userRoleMap[requiredRole] {
			return false
		}
	}
	return true
}

// GetRolePermissions returns all permissions for given roles
func (pc *PermissionChecker) GetRolePermissions(roleNames []string) []string {
	permissionSet := make(map[string]bool)

	for _, roleName := range roleNames {
		if role, exists := pc.roles[roleName]; exists {
			for _, permission := range role.Permissions {
				permissionSet[permission.ID] = true
			}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions
}

// ValidatePermissionFormat validates permission format (resource:action)
func ValidatePermissionFormat(permission string) error {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return fmt.Errorf("permission must be in format 'resource:action', got: %s", permission)
	}

	resource := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])

	if resource == "" {
		return fmt.Errorf("permission resource cannot be empty")
	}

	if action == "" {
		return fmt.Errorf("permission action cannot be empty")
	}

	// Validate resource format (alphanumeric, underscore, hyphen)
	if !isValidIdentifier(resource) {
		return fmt.Errorf("permission resource must contain only alphanumeric characters, underscores, and hyphens")
	}

	// Validate action format
	if !isValidIdentifier(action) && action != "*" {
		return fmt.Errorf("permission action must contain only alphanumeric characters, underscores, hyphens, or be '*'")
	}

	return nil
}

// isValidIdentifier checks if a string is a valid identifier
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i, r := range s {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
				return false
			}
		}
	}

	return true
}

// CommonPermissions defines commonly used permissions
var CommonPermissions = struct {
	// User management
	UserRead   string
	UserCreate string
	UserUpdate string
	UserDelete string
	UserList   string

	// Admin permissions
	AdminRead   string
	AdminCreate string
	AdminUpdate string
	AdminDelete string
	AdminList   string

	// System permissions
	SystemConfig  string
	SystemLogs    string
	SystemMetrics string

	// Content management
	PostRead   string
	PostCreate string
	PostUpdate string
	PostDelete string
	PostList   string
}{
	// User management
	UserRead:   "user:read",
	UserCreate: "user:create",
	UserUpdate: "user:update",
	UserDelete: "user:delete",
	UserList:   "user:list",

	// Admin permissions
	AdminRead:   "admin:read",
	AdminCreate: "admin:create",
	AdminUpdate: "admin:update",
	AdminDelete: "admin:delete",
	AdminList:   "admin:list",

	// System permissions
	SystemConfig:  "system:config",
	SystemLogs:    "system:logs",
	SystemMetrics: "system:metrics",

	// Content management
	PostRead:   "post:read",
	PostCreate: "post:create",
	PostUpdate: "post:update",
	PostDelete: "post:delete",
	PostList:   "post:list",
}

// CommonRoles defines commonly used roles
var CommonRoles = struct {
	SuperAdmin string
	Admin      string
	Moderator  string
	User       string
	Guest      string
}{
	SuperAdmin: "super_admin",
	Admin:      "admin",
	Moderator:  "moderator",
	User:       "user",
	Guest:      "guest",
}
