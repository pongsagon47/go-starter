package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Del deletes a key from cache
	Del(ctx context.Context, keys ...string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets TTL for a key
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL returns the remaining TTL of a key
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Incr increments a counter
	Incr(ctx context.Context, key string) (int64, error)

	// IncrBy increments a counter by value
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	// GetJSON retrieves and unmarshals JSON data
	GetJSON(ctx context.Context, key string, dest interface{}) error

	// SetJSON marshals and stores JSON data
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Close closes the cache connection
	Close() error

	// Ping checks if cache is available
	Ping(ctx context.Context) error

	// FlushAll clears all cache data (use with caution)
	FlushAll(ctx context.Context) error
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	DefaultTTL time.Duration
	KeyPrefix  string
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL: 1 * time.Hour,
		KeyPrefix:  "flex-service:",
	}
}
