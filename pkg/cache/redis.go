package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	config *CacheConfig
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, config *CacheConfig) Cache {
	if config == nil {
		config = DefaultCacheConfig()
	}
	return &RedisCache{
		client: client,
		config: config,
	}
}

// Get retrieves a value from Redis cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrCacheMiss
		}
		return "", fmt.Errorf("failed to get key %s: %w", fullKey, err)
	}
	return result, nil
}

// Set stores a value in Redis cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.buildKey(key)
	if ttl == 0 {
		ttl = r.config.DefaultTTL
	}

	err := r.client.Set(ctx, fullKey, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", fullKey, err)
	}
	return nil
}

// Del deletes keys from Redis cache
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	err := r.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists checks if keys exist in Redis cache
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	result, err := r.client.Exists(ctx, fullKeys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check existence of keys: %w", err)
	}
	return result, nil
}

// Expire sets TTL for a key in Redis cache
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := r.buildKey(key)
	err := r.client.Expire(ctx, fullKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiry for key %s: %w", fullKey, err)
	}
	return nil
}

// TTL returns the remaining TTL of a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", fullKey, err)
	}
	return result, nil
}

// Incr increments a counter in Redis cache
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", fullKey, err)
	}
	return result, nil
}

// IncrBy increments a counter by value in Redis cache
func (r *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.IncrBy(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", fullKey, value, err)
	}
	return result, nil
}

// GetJSON retrieves and unmarshals JSON data from Redis cache
func (r *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}
	return nil
}

// SetJSON marshals and stores JSON data in Redis cache
func (r *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	return r.Set(ctx, key, string(data), ttl)
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Ping checks if Redis is available
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// FlushAll clears all Redis cache data (use with caution)
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush all cache: %w", err)
	}
	return nil
}

// buildKey creates a full key with prefix
func (r *RedisCache) buildKey(key string) string {
	if r.config.KeyPrefix == "" {
		return key
	}
	return r.config.KeyPrefix + key
}
