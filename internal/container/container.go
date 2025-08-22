package container

import (
	"context"
	"flex-service/config"
	"flex-service/internal/user_auth"

	"flex-service/pkg/cache"
	"flex-service/pkg/database"
	"flex-service/pkg/logger"
	"flex-service/pkg/mail"
	"flex-service/pkg/rate_limit"
	"flex-service/pkg/secure"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Container implements ContainerInterface and holds all application dependencies
type Container struct {
	Config *config.Config

	// Core infrastructure
	Database  database.Database
	Cache     cache.Cache
	Mail      *mail.Mailer
	Secure    *secure.Secure
	RateLimit rate_limit.RateLimit

	// Backward compatibility (deprecated, use Database interface instead)
	DB *gorm.DB

	// Application services (registered via ServiceRegistry)
	UserAuthRepo    user_auth.UserAuthRepository
	UserAuthUsecase user_auth.UserAuthUsecase
	UserAuthHandler *user_auth.UserAuthHandler
}

// NewContainer creates a new container with all dependencies using the factory pattern
func NewContainer(cfg *config.Config) (*Container, error) {
	// Create factory
	factory := NewContainerFactory(cfg)

	// Create all dependencies
	deps, err := factory.CreateAll()
	if err != nil {
		logger.Error("Failed to create container dependencies", zap.Error(err))
		return nil, err
	}

	// Create container with core dependencies
	container := &Container{
		Config:    cfg,
		Database:  deps.Database,
		Cache:     deps.Cache,
		Mail:      deps.Mail,
		Secure:    deps.Secure,
		DB:        deps.Database.GetDB(), // Backward compatibility
		RateLimit: deps.RateLimit,
	}

	// Register application services
	registry := NewServiceRegistry(container)
	if err := registry.RegisterAll(); err != nil {
		logger.Error("Failed to register services", zap.Error(err))
		return nil, err
	}

	logger.Info("Container created successfully")
	return container, nil
}

// MustNewContainer creates a new container or panics (for backward compatibility)
func MustNewContainer(cfg *config.Config) *Container {
	container, err := NewContainer(cfg)
	if err != nil {
		logger.Fatal("Failed to create container", zap.Error(err))
	}
	return container
}

// =============================================================================
// ContainerInterface Implementation
// =============================================================================

// GetDatabase returns the database instance
func (c *Container) GetDatabase() database.Database {
	return c.Database
}

// GetCache returns the cache instance
func (c *Container) GetCache() cache.Cache {
	return c.Cache
}

// GetMail returns the mailer instance
func (c *Container) GetMail() *mail.Mailer {
	return c.Mail
}

// GetSecure returns the secure instance
func (c *Container) GetSecure() *secure.Secure {
	return c.Secure
}

// GetDatabaseType returns the current database type
func (c *Container) GetDatabaseType() database.DatabaseType {
	return c.Database.GetDatabaseType()
}

// HealthCheck performs health check on all critical dependencies
func (c *Container) HealthCheck(ctx context.Context) error {
	// Check database health
	if err := c.Database.HealthCheck(); err != nil {
		logger.Error("Database health check failed", zap.Error(err))
		return err
	}

	// Check cache health (if available)
	if c.Cache != nil {
		// TODO: Implement cache health check when available in cache interface
		logger.Debug("Cache health check skipped (not implemented)")
	}

	// Check mail connection (optional)
	if c.Mail != nil {
		if err := c.Mail.TestConnection(); err != nil {
			logger.Warn("Mail health check failed", zap.Error(err))
			// Mail failure shouldn't fail the entire health check
		}
	}

	logger.Info("Container health check passed")
	return nil
}

// RunMigrations runs database migrations
func (c *Container) RunMigrations() error {
	logger.Info("Running migrations", zap.String("database_type", string(c.GetDatabaseType())))
	return c.Database.RunMigrations()
}

// SeedData runs database seeders
func (c *Container) SeedData(seederName string) error {
	logger.Info("Running seeders",
		zap.String("database_type", string(c.GetDatabaseType())),
		zap.String("seeder", seederName))
	return c.Database.SeedData(seederName)
}

// Close closes all container resources gracefully
func (c *Container) Close() error {
	logger.Info("Closing container resources")

	var lastError error

	// Close cache connection if available
	if c.Cache != nil {
		if err := c.Cache.Close(); err != nil {
			logger.Error("Failed to close cache", zap.Error(err))
			lastError = err
		}
	}

	// Close database connection
	if err := c.Database.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		lastError = err
	}

	if lastError == nil {
		logger.Info("Container resources closed successfully")
	} else {
		logger.Error("Container resources closed with errors", zap.Error(lastError))
	}

	return lastError
}
