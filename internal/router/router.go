package router

import (
	"time"

	"go-starter/internal/container"
	"go-starter/internal/middleware"
	"go-starter/pkg/response"

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
	if container.Cache != nil {
		router.Use(middleware.IPRateLimit(container.Cache, 100, time.Minute))
	}
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
		// Authentication routes (public)
		// authRoutes := v1.Group("/auth")
		// {
		// 	authRoutes.POST("/register", container.AuthHandler.Register)
		// 	authRoutes.POST("/login", container.AuthHandler.Login)
		// 	authRoutes.POST("/refresh", container.AuthHandler.RefreshToken)
		// }

		// Protected routes (require authentication)
		// protected := v1.Group("/")
		// protected.Use(auth.AuthMiddleware(*container.JWT))
		// {
		// 	Auth related protected routes
		// 	authProtected := protected.Group("/auth")
		// 	{
		// 		authProtected.POST("/logout", container.AuthHandler.Logout)
		// 		authProtected.GET("/me", container.AuthHandler.Me)
		// 	}

		// 	Add other protected routes here
		// 	Example:
		// 	users := protected.Group("/users")
		// 	{
		// 	    users.GET("/", container.UserHandler.List)
		// 	    users.GET("/:id", container.UserHandler.GetByID)
		// 	    users.PUT("/:id", container.UserHandler.Update)
		// 	    users.DELETE("/:id", container.UserHandler.Delete)
		// 	}
		// }

		// Demo endpoint (public)
		v1.GET("/demo", func(c *gin.Context) {
			response.Success(c, 200, "Demo endpoint", gin.H{
				"message": "This is a demo endpoint for the starter project",
				"authentication_endpoints": gin.H{
					"register": "POST /api/v1/auth/register",
					"login":    "POST /api/v1/auth/login",
					"refresh":  "POST /api/v1/auth/refresh",
					"logout":   "POST /api/v1/auth/logout (requires auth)",
					"me":       "GET /api/v1/auth/me (requires auth)",
				},
				"tips": []string{
					"Use make make-package NAME=User to create a complete package",
					"Use make make-migration to create database migrations",
					"Use make make-seeder to create database seeders",
					"Check the README.md for more commands",
				},
			})
		})
	}

	return router
}
