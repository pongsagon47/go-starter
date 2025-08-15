package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-starter/config"
	"go-starter/internal/container"
	"go-starter/internal/router"
	"go-starter/pkg/logger"

	appTime "go-starter/pkg/time"

	// Import to register migrations and seeders
	_ "go-starter/internal/migrations"
	_ "go-starter/internal/seeders"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting application",
		zap.String("env", cfg.Env),
		zap.String("app_name", cfg.AppName),
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.String("database_type", string(cfg.Database.Type)),
	)

	// Initialize timezone
	if err := appTime.InitTimezone(cfg.Timezone); err != nil {
		logger.Fatal("Failed to initialize timezone", zap.Error(err))
	}

	// Initialize dependency injection container (includes database setup)
	containerInstance, err := container.NewContainer(cfg)
	if err != nil {
		logger.Fatal("Failed to create container", zap.Error(err))
	}

	// Run migrations if in development mode
	if cfg.Env == "development" {
		logger.Info("Running migrations in development mode")
		if err := containerInstance.RunMigrations(); err != nil {
			logger.Warn("Failed to run migrations", zap.Error(err))
		}

		// Seed data in development
		logger.Info("Running seeders in development mode")
		if err := containerInstance.SeedData(""); err != nil {
			logger.Warn("Failed to seed data", zap.Error(err))
		}
	}

	// Setup routes
	routerInstance := router.SetupRouter(containerInstance)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      routerInstance,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("address", server.Addr),
			zap.String("database", string(containerInstance.GetDatabaseType())),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Server started successfully",
		zap.String("address", server.Addr),
		zap.String("database_type", string(containerInstance.GetDatabaseType())))

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// Close container resources (includes database)
	if err := containerInstance.Close(); err != nil {
		logger.Error("Failed to close container resources", zap.Error(err))
	}

	logger.Info("Server exited")
}
