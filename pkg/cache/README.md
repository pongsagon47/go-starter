# 💾 Cache Package

High-performance Redis caching system with helper functions and tag-based cache management.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Interface](#interface)
- [Redis Implementation](#redis-implementation)
- [Cache Helpers](#cache-helpers)
- [Tag-based Caching](#tag-based-caching)
- [Configuration](#configuration)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/cache"
```

## ⚡ Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "time"
    "go-starter/pkg/cache"
    "go-starter/config"
)

func main() {
    // Create cache instance
    cfg := &config.RedisConfig{
        Host: "localhost",
        Port: 6379,
    }

    cache, err := cache.NewCache(cfg)
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Set value
    cache.Set(ctx, "user:123", "John Doe", time.Hour)

    // Get value
    value, err := cache.Get(ctx, "user:123")
    if err != nil {
        // Handle cache miss or error
    }

    fmt.Println(value) // "John Doe"
}
```

## 🔧 Interface

### Cache Interface

```go
type Cache interface {
    // Basic operations
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, keys ...string) (int64, error)

    // TTL operations
    Expire(ctx context.Context, key string, ttl time.Duration) error
    TTL(ctx context.Context, key string) (time.Duration, error)

    // Counter operations
    Incr(ctx context.Context, key string) (int64, error)
    IncrBy(ctx context.Context, key string, value int64) (int64, error)

    // JSON operations
    GetJSON(ctx context.Context, key string, dest interface{}) error
    SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

    // Utility operations
    Ping(ctx context.Context) error
    Close() error
    FlushAll(ctx context.Context) error
}
```

## 🎯 Cache Helpers

### Remember Pattern

The most common caching pattern - cache the result of expensive operations:

```go
helper := cache.NewCacheHelper(cacheInstance)

// Remember pattern - executes function only on cache miss
user, err := helper.Remember(ctx, "user:123", time.Hour, func() (interface{}, error) {
    // This function only runs if cache is empty
    return database.GetUser("123")
})

// String version
username, err := helper.RememberString(ctx, "username:123", time.Hour, func() (string, error) {
    return database.GetUsername("123")
})
```

### Cache Management

```go
// Forget single key
helper.Forget(ctx, "user:123")

// Forget multiple keys
helper.ForgetMany(ctx, "user:123", "user:456", "user:789")

// Cache forever (use with caution)
helper.Forever(ctx, "config:app", appConfig)
```

## 🏷️ Tag-based Caching

Group related cache entries for easy bulk operations:

```go
helper := cache.NewCacheHelper(cacheInstance)

// Cache with tags
products, err := helper.Tag("products").Remember(ctx, "category:electronics", time.Hour, func() (interface{}, error) {
    return database.GetProductsByCategory("electronics")
})

// Cache more items with same tag
categories, err := helper.Tag("products").Remember(ctx, "categories:all", time.Hour, func() (interface{}, error) {
    return database.GetAllCategories()
})

// Flush all cache entries with "products" tag
helper.Tag("products").Flush(ctx) // Removes both "category:electronics" and "categories:all"
```

### Tag Use Cases

```go
// User-related caches
helper.Tag("users").Remember(ctx, "user:profile:123", time.Hour, getUserProfile)
helper.Tag("users").Remember(ctx, "user:permissions:123", time.Hour, getUserPermissions)
helper.Tag("users").Flush(ctx) // Clear all user caches

// Product-related caches
helper.Tag("products").Remember(ctx, "product:123", time.Hour, getProduct)
helper.Tag("products").Remember(ctx, "products:featured", 30*time.Minute, getFeaturedProducts)
helper.Tag("products").Flush(ctx) // Clear all product caches
```

## ⚙️ Configuration

### RedisConfig

```go
type RedisConfig struct {
    Host         string        // Redis host (default: "localhost")
    Port         int           // Redis port (default: 6379)
    Password     string        // Redis password (default: "")
    DB           int           // Redis database (default: 0)
    MaxRetries   int           // Max retry attempts (default: 3)
    PoolSize     int           // Connection pool size (default: 10)
    MinIdleConns int           // Minimum idle connections (default: 5)
    DialTimeout  time.Duration // Connection timeout (default: 5s)
    ReadTimeout  time.Duration // Read timeout (default: 3s)
    WriteTimeout time.Duration // Write timeout (default: 3s)
}
```

### CacheConfig

```go
type CacheConfig struct {
    DefaultTTL time.Duration // Default expiration (default: 1 hour)
    KeyPrefix  string        // Key prefix (default: "go-starter:")
}
```

## 💡 Examples

### Database Query Caching

```go
func (s *UserService) GetUser(userID string) (*User, error) {
    helper := cache.NewCacheHelper(s.cache)

    user, err := helper.Remember(ctx, "user:"+userID, time.Hour, func() (interface{}, error) {
        return s.database.GetUser(userID) // Only called on cache miss
    })

    if err != nil {
        return nil, err
    }

    return user.(*User), nil
}
```

### API Response Caching

```go
func (h *ProductHandler) GetProducts(c *gin.Context) {
    category := c.Query("category")
    helper := cache.NewCacheHelper(h.cache)

    products, err := helper.Remember(ctx, "products:category:"+category, 30*time.Minute, func() (interface{}, error) {
        return h.service.GetProductsByCategory(category)
    })

    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get products"})
        return
    }

    c.JSON(200, products)
}
```

### Counter Operations

```go
// Page view counter
views, err := cache.Incr(ctx, "page:views:/products")

// Rate limiting
requests, err := cache.IncrBy(ctx, "rate_limit:user:123", 1)
if requests == 1 {
    cache.Expire(ctx, "rate_limit:user:123", time.Minute)
}
if requests > 100 {
    return errors.New("rate limit exceeded")
}
```

### Session Storage

```go
// Store session data
sessionData := map[string]interface{}{
    "user_id": 123,
    "role":    "admin",
    "permissions": []string{"read", "write"},
}
cache.SetJSON(ctx, "session:"+sessionID, sessionData, 24*time.Hour)

// Retrieve session data
var session map[string]interface{}
err := cache.GetJSON(ctx, "session:"+sessionID, &session)
```

## 🎯 Best Practices

### 1. **Key Naming Convention**

```go
// Good: Hierarchical and descriptive
"user:profile:123"
"product:details:456"
"cache:api:products:category:electronics"

// Bad: Unclear and flat
"u123"
"prod456"
"data"
```

### 2. **TTL Strategy**

```go
// Short TTL for frequently changing data
cache.Set(ctx, "stock:product:123", stock, 5*time.Minute)

// Medium TTL for semi-static data
cache.Set(ctx, "user:profile:123", profile, time.Hour)

// Long TTL for rarely changing data
cache.Set(ctx, "config:app", config, 24*time.Hour)
```

### 3. **Error Handling**

```go
// Always handle cache misses gracefully
user, err := cache.Get(ctx, "user:123")
if err != nil {
    if err == cache.ErrCacheMiss {
        // Cache miss is normal, get from database
        user = getUserFromDatabase("123")
        cache.Set(ctx, "user:123", user, time.Hour)
    } else {
        // Log error but don't fail the operation
        log.Printf("Cache error: %v", err)
        user = getUserFromDatabase("123")
    }
}
```

### 4. **Memory Management**

```go
// Use appropriate TTL to prevent memory bloat
cache.Set(ctx, "temp:data", data, 5*time.Minute) // Not forever

// Clean up when data changes
cache.Del(ctx, "user:profile:123") // When user updates profile

// Use tags for bulk cleanup
helper.Tag("products").Flush(ctx) // When products are updated
```

### 5. **Performance Optimization**

```go
// Batch operations when possible
keys := []string{"user:1", "user:2", "user:3"}
cache.Del(ctx, keys...) // Better than multiple Del() calls

// Use JSON for complex data
cache.SetJSON(ctx, "user:123", complexUserObject, time.Hour)

// Use strings for simple data
cache.Set(ctx, "counter:views", "1234", time.Hour)
```

## 🚨 Error Handling

### Common Errors

```go
// Cache miss (not an error, normal behavior)
if err == cache.ErrCacheMiss {
    // Get data from primary source
}

// Cache unavailable (Redis down)
if err == cache.ErrCacheUnavailable {
    // Fallback to database, log warning
}

// Serialization failed
if err == cache.ErrSerializationFailed {
    // Data couldn't be JSON marshaled
}
```

### Graceful Degradation

```go
func GetUserWithCache(userID string) (*User, error) {
    // Try cache first
    var user User
    err := cache.GetJSON(ctx, "user:"+userID, &user)
    if err == nil {
        return &user, nil
    }

    // Cache miss or error - use database
    user, err = database.GetUser(userID)
    if err != nil {
        return nil, err
    }

    // Try to cache for next time (ignore cache errors)
    cache.SetJSON(ctx, "user:"+userID, user, time.Hour)

    return &user, nil
}
```

## 🔗 Related Packages

- [`pkg/session`](../session/) - Redis-based session management
- [`pkg/database`](../database/) - Multi-database support
- [`internal/middleware/rate_limit.go`](../../internal/middleware/) - Rate limiting using cache

## 📚 Additional Resources

- [Redis Documentation](https://redis.io/documentation)
- [Go-Redis Client](https://github.com/go-redis/redis)
- [Caching Best Practices](https://redis.io/docs/manual/patterns/)
