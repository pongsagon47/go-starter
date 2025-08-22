# 📦 Container Package

Modern dependency injection container with Factory Pattern, Service Registry, and Interface-based design for the flex-service project.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Components](#components)
- [Usage Examples](#usage-examples)
- [Migration Guide](#migration-guide)
- [Best Practices](#best-practices)

## 🏗️ Overview

The container package provides a **modern dependency injection system** that follows industry best practices:

- **🏭 Factory Pattern**: Centralized dependency creation with proper error handling
- **📋 Service Registry**: Modular service registration system
- **🔌 Interface-based Design**: Loose coupling through well-defined interfaces
- **⚡ Graceful Error Handling**: Non-fatal error recovery where possible
- **🔄 Health Checks**: Built-in health monitoring for all dependencies
- **🧹 Resource Management**: Proper cleanup and resource lifecycle

## 🏛️ Architecture

### Before (Legacy)

```go
// Hardcoded, tightly coupled
container := container.NewContainer(cfg) // Could panic
jwt := pkgAuth.NewJWT("hardcoded-secret", 24*time.Hour, 720*time.Hour, "hardcoded")
authRepo := auth.NewAuthRepository(db)
// ... manual dependency wiring
```

### After (Modern)

```go
// Factory-based, configurable, error-handled
container, err := container.NewContainer(cfg)
if err != nil {
    // Handle error gracefully
}

// All dependencies created via factories
// Service registration is automatic
// Health checks are built-in
```

## 🧩 Components

### 1. **ContainerInterface**

```go
type ContainerInterface interface {
    // Core dependencies
    GetDatabase() database.Database
    GetCache() cache.Cache
    GetMail() *mail.Mailer
    GetSecure() *secure.Secure
    GetJWT() *pkgAuth.JWT

    // Lifecycle methods
    RunMigrations() error
    SeedData(seederName string) error
    Close() error
    HealthCheck(ctx context.Context) error

    // Database utilities
    GetDatabaseType() database.DatabaseType
}
```

### 2. **Factory Pattern**

```go
type ContainerFactory struct {
    config *config.Config
}

// Creates dependencies with proper error handling
func (f *ContainerFactory) CreateDatabase() (database.Database, error)
func (f *ContainerFactory) CreateCache() (cache.Cache, error)
func (f *ContainerFactory) CreateMailer() (*mail.Mailer, error)
func (f *ContainerFactory) CreateSecure() (*secure.Secure, error)
func (f *ContainerFactory) CreateJWT() (*pkgAuth.JWT, error)
```

### 3. **Service Registry**

```go
type ServiceRegistry struct {
    container *Container
}

// Modular service registration
func (r *ServiceRegistry) RegisterAuth() error
func (r *ServiceRegistry) RegisterUser() error
func (r *ServiceRegistry) RegisterProduct() error
```

### 4. **Modern Container**

```go
type Container struct {
    Config *config.Config

    // Core infrastructure
    Database database.Database
    Cache    cache.Cache
    Mail     *mail.Mailer
    Secure   *secure.Secure
    JWT      *pkgAuth.JWT

    // Backward compatibility
    DB *gorm.DB

    // Application services
    AuthRepo    auth.AuthRepository
    AuthUsecase auth.AuthUsecase
    AuthHandler *auth.AuthHandler
}
```

## 🚀 Usage Examples

### Basic Container Creation

```go
package main

import (
    "flex-service/config"
    "flex-service/internal/container"
    "flex-service/pkg/logger"
)

func main() {
    cfg := config.Load()

    // Create container with error handling
    container, err := container.NewContainer(cfg)
    if err != nil {
        logger.Fatal("Failed to create container", zap.Error(err))
    }
    defer container.Close()

    // Container is ready to use
    db := container.GetDatabase()
    cache := container.GetCache()
    jwt := container.GetJWT()
}
```

### Health Check Implementation

```go
func healthHandler(container container.ContainerInterface) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        if err := container.HealthCheck(ctx); err != nil {
            c.JSON(500, gin.H{
                "status": "unhealthy",
                "error": err.Error(),
            })
            return
        }

        c.JSON(200, gin.H{
            "status": "healthy",
            "database": container.GetDatabaseType(),
        })
    }
}
```

### Graceful Shutdown

```go
func gracefulShutdown(container *container.Container, server *http.Server) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down server...")

    // Shutdown HTTP server
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Error("Server forced to shutdown", zap.Error(err))
    }

    // Close container resources
    if err := container.Close(); err != nil {
        logger.Error("Failed to close container", zap.Error(err))
    }

    logger.Info("Server exited")
}
```

### Custom Service Registration

```go
// Extending the ServiceRegistry for custom services
func (r *ServiceRegistry) RegisterCustomService() error {
    if r.container.Database == nil {
        return errors.New("Database dependency not available")
    }

    // Create custom service dependencies
    customRepo := custom.NewRepository(r.container.Database.GetDB())
    customUsecase := custom.NewUsecase(customRepo, r.container.JWT)
    customHandler := custom.NewHandler(customUsecase)

    // Register in container (you may need to extend Container struct)
    logger.Info("Custom service registered successfully")
    return nil
}
```

## 🔄 Migration Guide

### From Legacy to Modern Container

#### 1. **Update Container Creation**

```go
// OLD - Could panic, no error handling
container := container.NewContainer(cfg)

// NEW - Proper error handling
container, err := container.NewContainer(cfg)
if err != nil {
    // Handle error appropriately
}
```

#### 2. **Use Interface Methods**

```go
// OLD - Direct field access
db := container.DB
cache := container.Cache

// NEW - Interface methods (recommended)
db := container.GetDatabase().GetDB()
cache := container.GetCache()
```

#### 3. **Add Health Checks**

```go
// NEW - Built-in health monitoring
func healthCheck(container container.ContainerInterface) error {
    ctx := context.Background()
    return container.HealthCheck(ctx)
}
```

#### 4. **Graceful Resource Cleanup**

```go
// NEW - Proper resource management
defer func() {
    if err := container.Close(); err != nil {
        logger.Error("Failed to close container", zap.Error(err))
    }
}()
```

### Backward Compatibility

The new container design maintains **full backward compatibility**:

- ✅ `container.DB` still available for direct GORM access
- ✅ `container.AuthHandler` still available for immediate use
- ✅ All existing code continues to work unchanged
- ✅ Migration can be done incrementally

## ✨ Best Practices

### 1. **Error Handling**

```go
// ✅ DO - Handle container creation errors
container, err := container.NewContainer(cfg)
if err != nil {
    return fmt.Errorf("failed to initialize application: %w", err)
}

// ❌ DON'T - Ignore errors
container := container.MustNewContainer(cfg) // This panics on error
```

### 2. **Resource Management**

```go
// ✅ DO - Always close resources
defer container.Close()

// ✅ DO - Check for close errors in production
if err := container.Close(); err != nil {
    logger.Error("Failed to close container", zap.Error(err))
}
```

### 3. **Health Monitoring**

```go
// ✅ DO - Implement health checks
app.GET("/health", func(c *gin.Context) {
    if err := container.HealthCheck(c.Request.Context()); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "healthy"})
})
```

### 4. **Interface Usage**

```go
// ✅ DO - Use interfaces for loose coupling
func processPayment(db database.Database, cache cache.Cache) error {
    // Implementation
}

// Call with container dependencies
processPayment(container.GetDatabase(), container.GetCache())

// ❌ DON'T - Direct field access in business logic
func processPayment(container *container.Container) error {
    db := container.DB  // Tight coupling
}
```

### 5. **Configuration-Driven**

```go
// ✅ DO - All dependencies configured via config
// JWT TTL, database settings, cache settings all from config

// ❌ DON'T - Hardcode values
jwt := pkgAuth.NewJWT("hardcoded-secret", 24*time.Hour, 720*time.Hour, "hardcoded")
```

## 🎯 Benefits

### **Before vs After Comparison**

| Aspect                   | Before (Legacy)       | After (Modern)                |
| ------------------------ | --------------------- | ----------------------------- |
| **Error Handling**       | Fatal crashes         | Graceful error recovery       |
| **Configuration**        | Hardcoded values      | Config-driven                 |
| **Coupling**             | Tight coupling        | Loose coupling via interfaces |
| **Testing**              | Hard to mock          | Easy to mock with interfaces  |
| **Resource Management**  | Manual, error-prone   | Automatic with proper cleanup |
| **Health Monitoring**    | None                  | Built-in health checks        |
| **Service Registration** | Manual wiring         | Automatic via registry        |
| **Extensibility**        | Modify core container | Add new services via registry |

### **Key Improvements**

1. **🛡️ Resilience**: Non-fatal errors don't crash the application
2. **🔧 Maintainability**: Clear separation of concerns
3. **🧪 Testability**: Easy to mock dependencies for testing
4. **📈 Scalability**: Easy to add new services without modifying core container
5. **🔍 Observability**: Built-in health checks and logging
6. **⚡ Performance**: Proper resource management and cleanup

## 🔗 Related Packages

- [`pkg/database`](../../pkg/database/) - Database factory and interfaces
- [`pkg/cache`](../../pkg/cache/) - Redis cache implementation
- [`pkg/auth`](../../pkg/auth/) - JWT authentication
- [`config`](../../config/) - Configuration management

## 📚 Additional Resources

- [Dependency Injection in Go](https://github.com/uber-go/dig)
- [Factory Pattern](https://refactoring.guru/design-patterns/factory-method)
- [Service Locator Pattern](https://martinfowler.com/articles/injection.html)
