package database

import (
	"fmt"
	"time"

	"flex-service/pkg/logger"
	"flex-service/pkg/migration"
	"flex-service/pkg/seeder"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type SQLite struct {
	DB     *gorm.DB
	config *SQLiteConfig
}

// NewSQLite creates a new SQLite database connection
func NewSQLite(config *SQLiteConfig) (*SQLite, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SQLite configuration: %w", err)
	}

	// Configure GORM logger based on config
	var logLevel gormLogger.LogLevel
	switch config.LogLevel {
	case "debug":
		logLevel = gormLogger.Info
	case "info":
		logLevel = gormLogger.Warn
	case "warn":
		logLevel = gormLogger.Error
	default:
		logLevel = gormLogger.Error
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: false,
		CreateBatchSize:                          1000,
	}

	// Connect to database
	dsn := config.GetConnectionString()
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		logger.Error("Failed to connect to SQLite database", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB", zap.Error(err))
		return nil, err
	}

	// Configure connection pool (SQLite-specific limits)
	sqlDB.SetMaxIdleConns(config.Pool.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Pool.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Pool.ConnMaxLifetime) * time.Minute)

	// Enable foreign keys if configured
	if config.ForeignKeys {
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			logger.Warn("Failed to enable foreign keys", zap.Error(err))
		}
	}

	logger.Info("Successfully connected to SQLite database",
		zap.String("file_path", config.FilePath),
		zap.Bool("in_memory", config.InMemory),
		zap.Bool("foreign_keys", config.ForeignKeys),
		zap.String("journal_mode", config.Journal))

	return &SQLite{
		DB:     db,
		config: config,
	}, nil
}

// Connect establishes database connection (already done in constructor)
func (s *SQLite) Connect() error {
	return s.HealthCheck()
}

// GetDB returns the GORM database instance
func (s *SQLite) GetDB() *gorm.DB {
	return s.DB
}

// Close closes the database connection
func (s *SQLite) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDatabaseType returns the database type
func (s *SQLite) GetDatabaseType() DatabaseType {
	return DBTypeSQLite
}

// GetConnectionString returns the connection string
func (s *SQLite) GetConnectionString() string {
	return s.config.GetConnectionString()
}

// RunMigrations runs database migrations using Laravel-style migration system
func (s *SQLite) RunMigrations() error {
	logger.Info("Starting SQLite migrations...")

	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(s.DB, config)

	// Run migrations
	if err := migrationManager.RunMigrations(); err != nil {
		logger.Error("Failed to run SQLite migrations", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	logger.Info("SQLite migrations completed successfully")
	return nil
}

// RollbackMigrations rolls back the specified number of migrations
func (s *SQLite) RollbackMigrations(count string) error {
	logger.Info("Starting SQLite migration rollback...", zap.String("count", count))

	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(s.DB, config)

	// Rollback migrations
	if err := migrationManager.RollbackMigrations(count); err != nil {
		logger.Error("Failed to rollback SQLite migrations", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	logger.Info("SQLite migration rollback completed successfully")
	return nil
}

// GetMigrationStatus returns the current migration status
func (s *SQLite) GetMigrationStatus() error {
	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(s.DB, config)

	// Get migration status
	if err := migrationManager.GetMigrationStatus(); err != nil {
		logger.Error("Failed to get SQLite migration status", zap.Error(err))
		return err
	}

	return nil
}

// SeedData seeds the database with initial data using Laravel-style seeders
func (s *SQLite) SeedData(seederName string) error {
	logger.Info("Starting SQLite database seeding...")

	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(s.DB, config)

	// Run seeders
	if err := seederManager.RunSeeders(seederName); err != nil {
		logger.Error("Failed to run SQLite seeders", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrSeederFailed, err)
	}

	logger.Info("SQLite database seeding completed successfully")
	return nil
}

// RunSpecificSeeder runs a specific seeder
func (s *SQLite) RunSpecificSeeder(seederName string) error {
	logger.Info("Running specific SQLite seeder...", zap.String("seeder", seederName))

	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(s.DB, config)

	// Run specific seeder
	if err := seederManager.RunSpecificSeeder(seederName); err != nil {
		logger.Error("Failed to run specific SQLite seeder", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrSeederFailed, err)
	}

	logger.Info("Specific SQLite seeder completed successfully")
	return nil
}

// ListSeeders lists all registered seeders
func (s *SQLite) ListSeeders() error {
	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(s.DB, config)

	// List seeders
	return seederManager.ListSeeders()
}

// HealthCheck checks the database connection health
func (s *SQLite) HealthCheck() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("SQLite ping failed: %w", err)
	}

	return nil
}

// GetDatabaseStats returns database connection statistics
func (s *SQLite) GetDatabaseStats() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	logger.Info("SQLite Connection Statistics",
		zap.Int("max_open_connections", stats.MaxOpenConnections),
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed))

	return nil
}
