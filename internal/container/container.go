package container

import (
	"go-starter/config"
	"go-starter/pkg/cache"
	"go-starter/pkg/database"
	"go-starter/pkg/logger"
	"go-starter/pkg/mail"
	"go-starter/pkg/secure"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	Config   *config.Config
	DB       *gorm.DB
	Database database.Database // Add database interface
	Cache    cache.Cache       // Add Redis cache
	Mail     *mail.Mailer
	Secure   *secure.Secure

	// Add your repositories, usecases, and handlers here
	// Example:
	// UserRepository UserRepository
	// UserUsecase    UserUsecase
	// UserHandler    *UserHandler
}

func NewContainer(cfg *config.Config) *Container {
	// Initialize database using factory
	factory := database.NewDatabaseFactory()
	dbConfig := cfg.GetDatabaseConfig()

	db, err := factory.CreateDatabase(dbConfig)
	if err != nil {
		logger.Fatal("Failed to initialize database",
			zap.Error(err),
			zap.String("database_type", string(cfg.Database.Type)))
	}

	// Test database connection
	if err := db.HealthCheck(); err != nil {
		logger.Fatal("Database health check failed",
			zap.Error(err),
			zap.String("database_type", string(cfg.Database.Type)))
	}

	logger.Info("Database connected successfully",
		zap.String("type", string(cfg.Database.Type)),
		zap.String("connection", db.GetConnectionString()))

	// Initialize mail
	mail, err := mail.NewGomail(&cfg.Email)
	if err != nil {
		logger.Fatal("Failed to initialize email", zap.Error(err))
	}

	if err := mail.TestConnection(); err != nil {
		logger.Fatal("Failed to test email connection", zap.Error(err))
	}

	logger.Info("Email connection successful")

	// Initialize Redis cache (optional)
	var cacheInstance cache.Cache
	if cfg.Env == "production" || cfg.Redis.Host != "" {
		var err error
		cacheInstance, err = cache.NewCache(&cfg.Redis)
		if err != nil {
			logger.Warn("Failed to initialize Redis cache, continuing without cache", zap.Error(err))
			cacheInstance = nil
		} else {
			logger.Info("Redis cache connected successfully",
				zap.String("host", cfg.Redis.Host),
				zap.Int("port", cfg.Redis.Port))
		}
	} else {
		logger.Info("Redis cache disabled (development mode without Redis host)")
		cacheInstance = nil
	}

	// Initialize secure
	secure, err := secure.NewSecure(&cfg.Secure)
	if err != nil {
		logger.Fatal("Failed to initialize secure", zap.Error(err))
	}

	// Initialize your dependencies here
	// Example:
	// userRepository := user.NewUserRepository(db.GetDB())
	// userUsecase := user.NewUserUsecase(cfg, userRepository, secure)
	// userHandler := user.NewUserHandler(userUsecase)

	return &Container{
		Config:   cfg,
		DB:       db.GetDB(),    // Keep GORM DB for backward compatibility
		Database: db,            // New database interface
		Cache:    cacheInstance, // Redis cache
		Mail:     mail,
		Secure:   secure,

		// Add your dependencies here
		// Example:
		// UserRepository: userRepository,
		// UserUsecase:    userUsecase,
		// UserHandler:    userHandler,
	}
}

// GetDatabaseType returns the current database type
func (c *Container) GetDatabaseType() database.DatabaseType {
	return c.Database.GetDatabaseType()
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

// Close closes all container resources
func (c *Container) Close() error {
	logger.Info("Closing container resources")

	// Close cache connection if available
	if c.Cache != nil {
		if err := c.Cache.Close(); err != nil {
			logger.Error("Failed to close cache", zap.Error(err))
		}
	}

	// Close database connection
	if err := c.Database.Close(); err != nil {
		logger.Error("Failed to close database", zap.Error(err))
		return err
	}

	logger.Info("Container resources closed successfully")
	return nil
}
