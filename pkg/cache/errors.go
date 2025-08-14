package cache

import "errors"

// Cache-related errors
var (
	// ErrCacheMiss indicates that a key was not found in cache
	ErrCacheMiss = errors.New("cache miss")

	// ErrCacheUnavailable indicates that cache is not available
	ErrCacheUnavailable = errors.New("cache unavailable")

	// ErrInvalidTTL indicates that TTL value is invalid
	ErrInvalidTTL = errors.New("invalid TTL")

	// ErrInvalidKey indicates that key is invalid
	ErrInvalidKey = errors.New("invalid key")

	// ErrSerializationFailed indicates that data serialization failed
	ErrSerializationFailed = errors.New("serialization failed")

	// ErrDeserializationFailed indicates that data deserialization failed
	ErrDeserializationFailed = errors.New("deserialization failed")
)
