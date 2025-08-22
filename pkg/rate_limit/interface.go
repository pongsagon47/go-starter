package rate_limit

import (
	"flex-service/pkg/cache"
	"time"

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

type RateLimit interface {
	RateLimitMiddleware(cache cache.Cache, config *RateLimitConfig) gin.HandlerFunc
	IPRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	UserRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	APIKeyRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	EndpointRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	GlobalRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	LoginRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	RegisterRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
	PasswordResetRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc
}

type rateLimit struct {
	cache  cache.Cache
	config *RateLimitConfig
}
