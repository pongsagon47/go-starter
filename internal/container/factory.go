package container

import (
	"errors"
	"go-starter/config"
	pkgAuth "go-starter/pkg/auth"
	"go-starter/pkg/cache"
	"go-starter/pkg/database"
	"go-starter/pkg/logger"
	"go-starter/pkg/mail"
	"go-starter/pkg/secure"
	"time"

	"go.uber.org/zap"
)

// ContainerFactory implements CompositeFactory
type ContainerFactory struct {
	config *config.Config
}

// NewContainerFactory creates a new container factory
func NewContainerFactory(cfg *config.Config) *ContainerFactory {
	return &ContainerFactory{
		config: cfg,
	}
}

// CreateDatabase creates database instance with proper error handling
func (f *ContainerFactory) CreateDatabase() (database.Database, error) {
	factory := database.NewDatabaseFactory()
	dbConfig := f.config.GetDatabaseConfig()

	db, err := factory.CreateDatabase(dbConfig)
	if err != nil {
		logger.Error("Failed to create database",
			zap.Error(err),
			zap.String("database_type", string(f.config.Database.Type)))
		return nil, err
	}

	// Test database connection
	if err := db.HealthCheck(); err != nil {
		logger.Error("Database health check failed",
			zap.Error(err),
			zap.String("database_type", string(f.config.Database.Type)))
		return nil, err
	}

	logger.Info("Database connected successfully",
		zap.String("type", string(f.config.Database.Type)),
		zap.String("connection", db.GetConnectionString()))

	return db, nil
}

// CreateCache creates cache instance with optional fallback
func (f *ContainerFactory) CreateCache() (cache.Cache, error) {
	// Skip cache creation in development without Redis config
	if f.config.Env != "production" && f.config.Redis.Host == "" {
		logger.Info("Cache disabled (development mode without Redis host)")
		return nil, nil // This is expected, not an error
	}

	cacheInstance, err := cache.NewCache(&f.config.Redis)
	if err != nil {
		logger.Warn("Failed to initialize Redis cache",
			zap.Error(err),
			zap.String("host", f.config.Redis.Host),
			zap.Int("port", f.config.Redis.Port))

		// In production, cache failure should be an error
		if f.config.Env == "production" {
			return nil, err
		}

		// In development, continue without cache
		return nil, nil
	}

	logger.Info("Redis cache connected successfully",
		zap.String("host", f.config.Redis.Host),
		zap.Int("port", f.config.Redis.Port))

	return cacheInstance, nil
}

// CreateMailer creates mail instance with connection test
func (f *ContainerFactory) CreateMailer() (*mail.Mailer, error) {
	mailer, err := mail.NewGomail(&f.config.Email)
	if err != nil {
		logger.Error("Failed to create mailer", zap.Error(err))
		return nil, err
	}

	if err := mailer.TestConnection(); err != nil {
		logger.Error("Failed to test email connection", zap.Error(err))
		return nil, err
	}

	logger.Info("Email connection successful")
	return mailer, nil
}

// CreateSecure creates secure instance
func (f *ContainerFactory) CreateSecure() (*secure.Secure, error) {
	secure, err := secure.NewSecure(&f.config.Secure)
	if err != nil {
		logger.Error("Failed to create secure instance", zap.Error(err))
		return nil, err
	}

	logger.Info("Secure instance created successfully")
	return secure, nil
}

// CreateJWT creates JWT instance from configuration
func (f *ContainerFactory) CreateJWT() (*pkgAuth.JWT, error) {
	if f.config.JWT.Secret == "" {
		return nil, errors.New("JWT secret is required")
	}

	// Use configured TTL values with defaults
	accessHours := f.config.JWT.ExpirationHours
	if accessHours == 0 {
		accessHours = 24 // default 24 hours
	}
	accessTTL := time.Duration(accessHours) * time.Hour

	refreshHours := f.config.JWT.RefreshExpirationHours
	if refreshHours == 0 {
		refreshHours = 720 // default 30 days (720 hours)
	}
	refreshTTL := time.Duration(refreshHours) * time.Hour

	issuer := f.config.AppName // Default issuer (could be made configurable later)

	jwt := pkgAuth.NewJWT(f.config.JWT.Secret, accessTTL, refreshTTL, issuer)

	logger.Info("JWT instance created successfully",
		zap.Duration("access_ttl", accessTTL),
		zap.Duration("refresh_ttl", refreshTTL),
		zap.String("issuer", issuer))

	return jwt, nil
}

// CreateAll creates all dependencies at once
func (f *ContainerFactory) CreateAll() (*AllDependencies, error) {
	deps := &AllDependencies{}
	var err error

	// Create database (required)
	deps.Database, err = f.CreateDatabase()
	if err != nil {
		return nil, err
	}

	// Create cache (optional)
	deps.Cache, err = f.CreateCache()
	if err != nil {
		return nil, err
	}

	// Create mailer (required)
	deps.Mail, err = f.CreateMailer()
	if err != nil {
		return nil, err
	}

	// Create secure (required)
	deps.Secure, err = f.CreateSecure()
	if err != nil {
		return nil, err
	}

	// Create JWT (required)
	deps.JWT, err = f.CreateJWT()
	if err != nil {
		return nil, err
	}

	return deps, nil
}

// AllDependencies holds all created dependencies
type AllDependencies struct {
	Database database.Database
	Cache    cache.Cache
	Mail     *mail.Mailer
	Secure   *secure.Secure
	JWT      *pkgAuth.JWT
}
