package container

import (
	"flex-service/config"
	"flex-service/pkg/cache"
	"flex-service/pkg/database"
	"flex-service/pkg/logger"
	"flex-service/pkg/mail"
	"flex-service/pkg/rate_limit"
	"flex-service/pkg/secure"

	"github.com/gin-gonic/gin"
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

// CreateRateLimit creates rate limit instance

func (f *ContainerFactory) CreateRateLimit(cache cache.Cache) (rate_limit.RateLimit, error) {
	rateLimitConfig := &rate_limit.RateLimitConfig{
		Limit:  f.config.Ratelimit.Limit,
		Window: f.config.Ratelimit.Window,
		Skip: func(c *gin.Context) bool {
			return f.config.Env == "development"
		},
	}

	rateLimit, err := rate_limit.NewRateLimit(cache, rateLimitConfig)
	if err != nil {
		logger.Error("Failed to create rate limit instance", zap.Error(err))
		return nil, err
	}

	logger.Info("Rate limit instance created successfully")
	return rateLimit, nil
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

	// Create rate limit (required)
	deps.RateLimit, err = f.CreateRateLimit(deps.Cache)
	if err != nil {
		return nil, err
	}

	return deps, nil
}

// AllDependencies holds all created dependencies
type AllDependencies struct {
	Database  database.Database
	Cache     cache.Cache
	Mail      *mail.Mailer
	Secure    *secure.Secure
	RateLimit rate_limit.RateLimit
}
