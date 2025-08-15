# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-01-15

### 🎉 Major Features Added

#### **🏭 Modern Container Design with Factory Pattern**

- **Added** Factory Pattern for dependency creation with proper error handling
- **Added** Service Registry for modular service registration
- **Added** ContainerInterface for loose coupling and better testability
- **Added** Built-in health checks for all dependencies
- **Added** Graceful error handling and resource cleanup
- **Improved** Non-fatal error recovery (e.g., Redis cache failure doesn't crash app)

#### **🔧 Enhanced CLI with Multi-Database Support**

- **Added** Dynamic migration discovery - no more manual artisan rebuilds
- **Added** Multi-database migration templates (SQLite, MySQL, PostgreSQL)
- **Added** Primary key strategies: `int`, `uuid`, and `dual` (int + uuid)
- **Added** `STRATEGY` parameter to `make:model` and `make:migration` commands
- **Improved** Database-specific GORM tag generation
- **Fixed** SQLite compatibility issues with timestamp functions

#### **🚀 Error Helper Functions**

- **Added** Convenience functions: `errors.NotFound()`, `errors.BadRequest()`, etc.
- **Added** Wrapping functions: `errors.WrapDatabase()`, `errors.WrapInternal()`, etc.
- **Added** Auth-specific helpers: `errors.UserNotFound()`, `errors.InvalidCredentials()`, etc.
- **Added** Database-specific helpers for better error categorization
- **Improved** Error handling consistency across the codebase

#### **🔐 Complete Authentication System**

- **Added** JWT-based authentication with access and refresh tokens
- **Added** User registration, login, logout, and token refresh endpoints
- **Added** Role-based authorization with permissions
- **Added** Password hashing with bcrypt
- **Added** Authentication middleware for protected routes
- **Added** User profile management (`/api/v1/auth/me`)

### 🛠️ Improvements

#### **📦 Entity & Migration Enhancements**

- **Added** Support for dual primary key strategy (int ID as primary + UUID for public APIs)
- **Added** Auto-generated UUID fields with BeforeCreate hooks
- **Added** Database-specific migration templates
- **Improved** Entity generation with configurable primary key strategies
- **Improved** Primary key consistency: `int` and `dual` strategies use `ID` as primary key
- **Fixed** GORM tag synchronization between entities and migrations

#### **🏗️ Architecture Improvements**

- **Refactored** Container from hardcoded dependencies to factory-based creation
- **Added** Interface-based design for better testing and mocking
- **Improved** Separation of concerns with Service Registry pattern
- **Enhanced** Resource management with proper cleanup procedures

#### **📚 Documentation Updates**

- **Updated** README.md with new features and examples
- **Added** Container design documentation (`internal/container/README.md`)
- **Enhanced** Error handling documentation (`pkg/errors/README.md`)
- **Added** Authentication examples and API documentation
- **Improved** Primary key strategy documentation and examples

### 🔄 Breaking Changes

#### **Container Usage**

```go
// OLD (v1.x)
container := container.NewContainer(cfg) // Could panic

// NEW (v2.0)
container, err := container.NewContainer(cfg)
if err != nil {
    // Handle error gracefully
}
```

#### **Migration CLI**

```bash
# OLD (v1.x)
make make-migration NAME=create_users_table

# NEW (v2.0) - with strategy support
make make-migration NAME=create_users_table STRATEGY=dual CREATE=true TABLE=users
```

### 🛡️ Security Enhancements

- **Added** Complete JWT authentication system
- **Added** Password hashing with configurable salt rounds
- **Added** Role and permission-based authorization
- **Added** Secure token refresh mechanism
- **Improved** Error messages to avoid information leakage

### 🐛 Bug Fixes

- **Fixed** SQLite migration compatibility issues with `CURRENT_TIMESTAMP(3)`
- **Fixed** Dynamic migration discovery and registration
- **Fixed** Environment variable handling in artisan CLI
- **Fixed** GORM tag consistency between entities and migrations
- **Fixed** Resource cleanup in container close operations

### 📈 Performance Improvements

- **Improved** Database connection management with proper pooling
- **Added** Graceful degradation for optional dependencies (Redis cache)
- **Optimized** Migration discovery with generated import files
- **Enhanced** Error handling performance with helper functions

### 🧪 Testing & Development

- **Enhanced** Container testability with interface-based design
- **Improved** Mock capabilities for all dependencies
- **Added** Comprehensive authentication flow testing
- **Updated** Development workflow with new CLI features

---

## [1.0.0] - 2025-01-01

### Initial Release

- **Added** Multi-database support (MySQL, PostgreSQL, SQLite)
- **Added** Laravel-style CLI (Artisan) for migrations and seeders
- **Added** Clean Architecture with Repository pattern
- **Added** Basic dependency injection container
- **Added** Database factory pattern
- **Added** Migration and seeder systems
- **Added** Health check endpoints
- **Added** Docker support
- **Added** Makefile with development commands

---

## Migration Guide from v1.x to v2.0

### Container Usage

Update your container initialization to handle errors:

```go
// Before
container := container.NewContainer(cfg)

// After
container, err := container.NewContainer(cfg)
if err != nil {
    logger.Fatal("Failed to create container", zap.Error(err))
}
defer container.Close() // Add proper cleanup
```

### CLI Commands

Update your Makefile targets to use the new strategy parameter:

```bash
# Before
make make-model NAME=User TABLE=users

# After (with primary key strategy)
make make-model NAME=User TABLE=users STRATEGY=dual
```

### Error Handling

Replace verbose error creation with helper functions:

```go
// Before
err := errors.New("USER_NOT_FOUND", "User not found", 404)

// After
err := errors.UserNotFound()
```

### Authentication Integration

The authentication system is ready to use out of the box. Update your routes to use the new auth endpoints and middleware.

---

For more details on specific changes, see the documentation in each package's README file.
