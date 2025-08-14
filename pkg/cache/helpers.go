package cache

import (
	"context"
	"fmt"
	"time"
)

// CacheHelper provides high-level caching operations
type CacheHelper struct {
	cache Cache
}

// NewCacheHelper creates a new cache helper
func NewCacheHelper(cache Cache) *CacheHelper {
	return &CacheHelper{cache: cache}
}

// Remember caches the result of a function for a given key and TTL
// If the key exists, it returns the cached value
// If not, it executes the function, caches the result, and returns it
func (h *CacheHelper) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	var result interface{}
	err := h.cache.GetJSON(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	// If not cache miss, return the error
	if err != ErrCacheMiss {
		return nil, fmt.Errorf("cache error for key %s: %w", key, err)
	}

	// Execute function to get fresh data
	data, err := fn()
	if err != nil {
		return nil, fmt.Errorf("function execution failed for key %s: %w", key, err)
	}

	// Cache the result
	if cacheErr := h.cache.SetJSON(ctx, key, data, ttl); cacheErr != nil {
		// Log cache error but don't fail the operation
		// In production, you might want to log this properly
		fmt.Printf("Warning: failed to cache result for key %s: %v\n", key, cacheErr)
	}

	return data, nil
}

// RememberString is like Remember but for string values
func (h *CacheHelper) RememberString(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	// Try to get from cache first
	result, err := h.cache.Get(ctx, key)
	if err == nil {
		return result, nil
	}

	// If not cache miss, return the error
	if err != ErrCacheMiss {
		return "", fmt.Errorf("cache error for key %s: %w", key, err)
	}

	// Execute function to get fresh data
	data, err := fn()
	if err != nil {
		return "", fmt.Errorf("function execution failed for key %s: %w", key, err)
	}

	// Cache the result
	if cacheErr := h.cache.Set(ctx, key, data, ttl); cacheErr != nil {
		// Log cache error but don't fail the operation
		fmt.Printf("Warning: failed to cache result for key %s: %v\n", key, cacheErr)
	}

	return data, nil
}

// Forget removes a key from cache
func (h *CacheHelper) Forget(ctx context.Context, key string) error {
	return h.cache.Del(ctx, key)
}

// ForgetMany removes multiple keys from cache
func (h *CacheHelper) ForgetMany(ctx context.Context, keys ...string) error {
	return h.cache.Del(ctx, keys...)
}

// Forever caches a value without expiration (use with caution)
func (h *CacheHelper) Forever(ctx context.Context, key string, value interface{}) error {
	return h.cache.SetJSON(ctx, key, value, 0) // 0 means no expiration
}

// Tags provides a way to group cache keys for easier invalidation
type CacheTag struct {
	helper *CacheHelper
	tag    string
}

// Tag creates a new cache tag for grouping related cache entries
func (h *CacheHelper) Tag(tag string) *CacheTag {
	return &CacheTag{
		helper: h,
		tag:    tag,
	}
}

// Remember caches with tag
func (t *CacheTag) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	taggedKey := t.buildTaggedKey(key)
	result, err := t.helper.Remember(ctx, taggedKey, ttl, fn)
	if err != nil {
		return nil, err
	}

	// Add key to tag set
	t.addToTagSet(ctx, key)

	return result, nil
}

// Flush removes all cache entries with this tag
func (t *CacheTag) Flush(ctx context.Context) error {
	// Get all keys for this tag
	keys, err := t.getTagKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tag keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Add tagged keys to deletion list
	taggedKeys := make([]string, len(keys))
	for i, key := range keys {
		taggedKeys[i] = t.buildTaggedKey(key)
	}

	// Delete all tagged keys and the tag set itself
	allKeys := append(taggedKeys, t.getTagSetKey())
	return t.helper.cache.Del(ctx, allKeys...)
}

// buildTaggedKey creates a tagged cache key
func (t *CacheTag) buildTaggedKey(key string) string {
	return fmt.Sprintf("tag:%s:%s", t.tag, key)
}

// getTagSetKey returns the key for storing tag members
func (t *CacheTag) getTagSetKey() string {
	return fmt.Sprintf("tagset:%s", t.tag)
}

// addToTagSet adds a key to the tag set
func (t *CacheTag) addToTagSet(ctx context.Context, key string) error {
	// For simplicity, we'll store tag members as a JSON array
	// In a more sophisticated implementation, you might use Redis sets
	tagSetKey := t.getTagSetKey()

	// Get existing keys
	var existingKeys []string
	err := t.helper.cache.GetJSON(ctx, tagSetKey, &existingKeys)
	if err != nil && err != ErrCacheMiss {
		return err
	}

	// Add new key if not already present
	found := false
	for _, existingKey := range existingKeys {
		if existingKey == key {
			found = true
			break
		}
	}

	if !found {
		existingKeys = append(existingKeys, key)
		return t.helper.cache.SetJSON(ctx, tagSetKey, existingKeys, 24*time.Hour)
	}

	return nil
}

// getTagKeys retrieves all keys for this tag
func (t *CacheTag) getTagKeys(ctx context.Context) ([]string, error) {
	var keys []string
	err := t.helper.cache.GetJSON(ctx, t.getTagSetKey(), &keys)
	if err != nil {
		if err == ErrCacheMiss {
			return []string{}, nil
		}
		return nil, err
	}
	return keys, nil
}
