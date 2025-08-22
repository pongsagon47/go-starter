package rate_limit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"flex-service/pkg/cache"
	"flex-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRateLimit(cache cache.Cache, cfg *RateLimitConfig) (RateLimit, error) {
	rateLimitConfig := DefaultRateLimitConfig()
	if cfg != nil {
		rateLimitConfig.Limit = cfg.Limit
		rateLimitConfig.Window = cfg.Window
		rateLimitConfig.Skip = cfg.Skip
	}

	return &rateLimit{
		cache:  cache,
		config: rateLimitConfig,
	}, nil
}

// RateLimitMiddleware creates a rate limiting middleware using Redis
func (r *rateLimit) RateLimitMiddleware(cache cache.Cache, config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		logger.Debug("Rate limit skipped - no cache available")
		config = DefaultRateLimitConfig()
	}

	// Merge custom config with instance config
	mergedConfig := r.mergeConfig(config)

	return func(c *gin.Context) {
		if cache == nil {
			c.Next()
			return
		}

		// Skip if skip function is defined and returns true
		if mergedConfig.Skip != nil && mergedConfig.Skip(c) {
			c.Next()
			return
		}

		// Generate cache key
		key := mergedConfig.KeyGenerator(c)
		ctx := context.Background()

		// Get current count
		count, err := cache.Incr(ctx, key)
		if err != nil {
			// Log error but don't block request if cache is unavailable
			logger.Warn("Rate limit cache error, allowing request",
				zap.Error(err),
				zap.String("key", key))
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			if err := cache.Expire(ctx, key, mergedConfig.Window); err != nil {
				logger.Warn("Failed to set cache expiration",
					zap.Error(err),
					zap.String("key", key))
			}
		}

		// Check if limit exceeded
		if count > int64(mergedConfig.Limit) {
			// Get TTL for rate limit reset time
			ttl, err := cache.TTL(ctx, key)
			if err != nil {
				ttl = mergedConfig.Window
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(mergedConfig.Limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			// Call custom handler if provided
			if mergedConfig.OnRateLimited != nil {
				mergedConfig.OnRateLimited(c, mergedConfig.Limit, mergedConfig.Window)
				return
			}

			// Default response
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     mergedConfig.Message,
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		remaining := mergedConfig.Limit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(mergedConfig.Limit))
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
func (r *rateLimit) IPRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			return "rate_limit:ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v allowed.", limit, window),
	}
	return r.RateLimitMiddleware(cache, config)
}

// UserRateLimit creates a user-based rate limiter (requires authentication)
func (r *rateLimit) UserRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
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
	return r.RateLimitMiddleware(cache, config)
}

// APIKeyRateLimit creates an API key-based rate limiter
func (r *rateLimit) APIKeyRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
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
	return r.RateLimitMiddleware(cache, config)
}

// EndpointRateLimit creates an endpoint-specific rate limiter
func (r *rateLimit) EndpointRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			endpoint := c.Request.Method + ":" + c.FullPath()
			return "rate_limit:endpoint:" + endpoint + ":ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Rate limit exceeded for this endpoint. Maximum %d requests per %v allowed.", limit, window),
	}
	return r.RateLimitMiddleware(cache, config)
}

// GlobalRateLimit creates a global rate limiter across all requests
func (r *rateLimit) GlobalRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			return "rate_limit:global"
		},
		Message: fmt.Sprintf("Global rate limit exceeded. Maximum %d requests per %v allowed across all users.", limit, window),
	}
	return r.RateLimitMiddleware(cache, config)
}

// LoginRateLimit creates a login-specific rate limiter to prevent brute force attacks
func (r *rateLimit) LoginRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			// Use username + IP for login attempts
			username := c.PostForm("username")
			if username == "" {
				// Try to get from JSON body
				var loginData struct {
					Username string `json:"username"`
				}
				if err := c.ShouldBindJSON(&loginData); err == nil {
					username = loginData.Username
				}
			}

			if username == "" {
				// Fallback to IP-only if no username
				return "rate_limit:login:ip:" + c.ClientIP()
			}

			return "rate_limit:login:" + username + ":ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Too many login attempts. Please try again in %v.", window),
		OnRateLimited: func(c *gin.Context, limit int, window time.Duration) {
			// Custom response for login rate limiting
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":          "TOO_MANY_LOGIN_ATTEMPTS",
				"message":        fmt.Sprintf("Too many login attempts. Please try again in %v.", window),
				"retry_after":    int(window.Seconds()),
				"account_locked": true,
			})
			c.Abort()
		},
	}
	return r.RateLimitMiddleware(cache, config)
}

// RegisterRateLimit creates a registration-specific rate limiter
func (r *rateLimit) RegisterRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			// Use IP for registration attempts
			return "rate_limit:register:ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Too many registration attempts. Please try again in %v.", window),
		OnRateLimited: func(c *gin.Context, limit int, window time.Duration) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "TOO_MANY_REGISTRATION_ATTEMPTS",
				"message":     fmt.Sprintf("Too many registration attempts. Please try again in %v.", window),
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
		},
	}
	return r.RateLimitMiddleware(cache, config)
}

// PasswordResetRateLimit creates a password reset-specific rate limiter
func (r *rateLimit) PasswordResetRateLimit(cache cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyGenerator: func(c *gin.Context) string {
			// Use email + IP for password reset attempts
			email := c.PostForm("email")
			if email == "" {
				var resetData struct {
					Email string `json:"email"`
				}
				if err := c.ShouldBindJSON(&resetData); err == nil {
					email = resetData.Email
				}
			}

			if email == "" {
				return "rate_limit:password_reset:ip:" + c.ClientIP()
			}

			return "rate_limit:password_reset:" + email + ":ip:" + c.ClientIP()
		},
		Message: fmt.Sprintf("Too many password reset attempts. Please try again in %v.", window),
		OnRateLimited: func(c *gin.Context, limit int, window time.Duration) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "TOO_MANY_PASSWORD_RESET_ATTEMPTS",
				"message":     fmt.Sprintf("Too many password reset attempts. Please try again in %v.", window),
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
		},
	}
	return r.RateLimitMiddleware(cache, config)
}
