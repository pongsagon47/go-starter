package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-starter/pkg/cache"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Requests per window
	Limit int
	// Time window duration
	Window time.Duration
	// Key generator function (default: IP-based)
	KeyGenerator func(c *gin.Context) string
	// Skip function to bypass rate limiting for certain requests
	Skip func(c *gin.Context) bool
	// Custom error message
	Message string
	// Custom error handler
	OnRateLimited func(c *gin.Context, limit int, window time.Duration)
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Limit:  100,         // 100 requests
		Window: time.Minute, // per minute
		KeyGenerator: func(c *gin.Context) string {
			return "rate_limit:ip:" + c.ClientIP()
		},
		Skip:          nil,
		Message:       "Rate limit exceeded. Please try again later.",
		OnRateLimited: nil,
	}
}

// RateLimitMiddleware creates a rate limiting middleware using Redis
func RateLimitMiddleware(cache cache.Cache, config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(c *gin.Context) {
		// Skip if skip function is defined and returns true
		if config.Skip != nil && config.Skip(c) {
			c.Next()
			return
		}

		// Generate cache key
		key := config.KeyGenerator(c)
		ctx := context.Background()

		// Get current count
		count, err := cache.Incr(ctx, key)
		if err != nil {
			// Log error but don't block request if cache is unavailable
			fmt.Printf("Rate limit cache error: %v\n", err)
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			if err := cache.Expire(ctx, key, config.Window); err != nil {
				fmt.Printf("Rate limit expire error: %v\n", err)
			}
		}

		// Check if limit exceeded
		if count > int64(config.Limit) {
			// Get TTL for rate limit reset time
			ttl, err := cache.TTL(ctx, key)
			if err != nil {
				ttl = config.Window
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.Limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			// Call custom handler if provided
			if config.OnRateLimited != nil {
				config.OnRateLimited(c, config.Limit, config.Window)
				return
			}

			// Default response
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     config.Message,
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		remaining := config.Limit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		// Get TTL for reset time
		ttl, err := cache.TTL(ctx, key)
		if err == nil {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
		}

		c.Next()
	}
}

// IPRateLimit creates an IP-based rate limiter
func IPRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			return "rate_limit:ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v allowed.", limit, window),
	}
	return RateLimitMiddleware(cache, config)
}

// UserRateLimit creates a user-based rate limiter (requires authentication)
func UserRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			// Try to get user ID from context (set by auth middleware)
			userID, exists := c.Get("user_id")
			if !exists {
				// Fallback to IP-based if no user ID
				return "rate_limit:ip:" + c.ClientIP()
			}
			return "rate_limit:user:" + fmt.Sprintf("%v", userID)
		},
		Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v allowed per user.", limit, window),
	}
	return RateLimitMiddleware(cache, config)
}

// APIKeyRateLimit creates an API key-based rate limiter
func APIKeyRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.Query("api_key")
			}
			if apiKey == "" {
				// Fallback to IP-based if no API key
				return "rate_limit:ip:" + c.ClientIP()
			}
			return "rate_limit:apikey:" + apiKey
		},
		Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v allowed per API key.", limit, window),
	}
	return RateLimitMiddleware(cache, config)
}

// EndpointRateLimit creates an endpoint-specific rate limiter
func EndpointRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			endpoint := c.Request.Method + ":" + c.FullPath()
			return "rate_limit:endpoint:" + endpoint + ":ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Rate limit exceeded for this endpoint. Maximum %d requests per %v allowed.", limit, window),
	}
	return RateLimitMiddleware(cache, config)
}

// GlobalRateLimit creates a global rate limiter across all requests
func GlobalRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			return "rate_limit:global"
		},
		Message: fmt.Sprintf("Global rate limit exceeded. Maximum %d requests per %v allowed across all users.", limit, window),
	}
	return RateLimitMiddleware(cache, config)
}
