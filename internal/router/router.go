package router

import (
	"time"

	"flex-service/internal/container"
	"flex-service/internal/middleware"
	"flex-service/pkg/response"

	"github.com/gin-gonic/gin"
)

func SetupRouter(container *container.Container) *gin.Engine {
	// Set Gin mode based on environment
	if container.Config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())
	router.Use(middleware.Logging())
	router.Use(middleware.Helmet())

	// Rate limiting middleware (only if Redis cache is available)
	router.Use(container.RateLimit.IPRateLimit(container.Cache, 100, time.Minute))
	router.Use(middleware.ErrorHandler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, 200, "Server is running", gin.H{
			"status":  "OK",
			"version": "1.0.0",
			"env":     container.Config.Env,
		})
	})

	// Database health check
	router.GET("/health/db", func(c *gin.Context) {
		sqlDB, err := container.DB.DB()
		if err != nil {
			response.Error(c, 500, "DATABASE_ERROR", "Failed to get database connection", nil)
			return
		}

		if err := sqlDB.Ping(); err != nil {
			response.Error(c, 500, "DATABASE_ERROR", "Database connection failed", nil)
			return
		}

		response.Success(c, 200, "Database is healthy", gin.H{
			"status": "OK",
		})
	})

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		response.Error(c, 404, "NOT_FOUND", "Route not found", gin.H{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		userAuthRoutes := v1.Group("/user-auth")
		{
			// ปรับให้เข้มงวดขึ้น (5 ครั้ง/15 นาที แทน 30 ครั้ง/15 นาที)
			userAuthRoutes.POST("/login", container.RateLimit.LoginRateLimit(container.Cache, 5, 15*time.Minute), container.UserAuthHandler.Login)
			userAuthRoutes.POST("/login-social", container.RateLimit.LoginRateLimit(container.Cache, 5, 15*time.Minute), container.UserAuthHandler.LoginWithSocialAccount)
			userAuthRoutes.POST("/register", container.RateLimit.RegisterRateLimit(container.Cache, 15, 1*time.Hour), container.UserAuthHandler.Register)
			userAuthRoutes.POST("/register-social", container.RateLimit.RegisterRateLimit(container.Cache, 15, 1*time.Hour), container.UserAuthHandler.RegisterWithSocialAccount)
			userAuthRoutes.POST("/refresh", container.RateLimit.IPRateLimit(container.Cache, 10, 1*time.Minute), container.UserAuthHandler.RefreshToken)

			// Protected routes with user-based rate limiting
			userAuthProtected := userAuthRoutes.Group("/")
			userAuthProtected.Use(middleware.UserAuthenticate(container.UserAuthUsecase))
			{
				userAuthProtected.POST("/logout", container.RateLimit.UserRateLimit(container.Cache, 10, 1*time.Minute), container.UserAuthHandler.Logout)
				userAuthProtected.GET("/me", container.RateLimit.UserRateLimit(container.Cache, 30, 1*time.Minute), container.UserAuthHandler.Me)
			}
		}
	}

	return router
}
