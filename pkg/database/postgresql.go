package database

import (
	"fmt"
	"time"

	"go-starter/pkg/logger"
	"go-starter/pkg/migration"
	"go-starter/pkg/seeder"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type PostgreSQL struct {
	DB     *gorm.DB
	config *PostgreSQLConfig
}

// NewPostgreSQL creates a new PostgreSQL database connection
func NewPostgreSQL(config *PostgreSQLConfig) (*PostgreSQL, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL configuration: %w", err)
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
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL database", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB", zap.Error(err))
		return nil, err
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(config.Pool.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Pool.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Pool.ConnMaxLifetime) * time.Minute)

	logger.Info("Successfully connected to PostgreSQL database",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Name),
		zap.Int("max_idle_conns", config.Pool.MaxIdleConns),
		zap.Int("max_open_conns", config.Pool.MaxOpenConns))

	return &PostgreSQL{
		DB:     db,
		config: config,
	}, nil
}

// Connect establishes database connection (already done in constructor)
func (p *PostgreSQL) Connect() error {
	return p.HealthCheck()
}

// GetDB returns the GORM database instance
func (p *PostgreSQL) GetDB() *gorm.DB {
	return p.DB
}

// Close closes the database connection
func (p *PostgreSQL) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDatabaseType returns the database type
func (p *PostgreSQL) GetDatabaseType() DatabaseType {
	return DBTypePostgreSQL
}

// GetConnectionString returns the connection string (masked password)
func (p *PostgreSQL) GetConnectionString() string {
	maskedConfig := *p.config
	maskedConfig.Password = "****"
	return maskedConfig.GetConnectionString()
}

// RunMigrations runs database migrations using Laravel-style migration system
func (p *PostgreSQL) RunMigrations() error {
	logger.Info("Starting PostgreSQL migrations...")

	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(p.DB, config)

	// Run migrations
	if err := migrationManager.RunMigrations(); err != nil {
		logger.Error("Failed to run PostgreSQL migrations", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	logger.Info("PostgreSQL migrations completed successfully")
	return nil
}

// RollbackMigrations rolls back the specified number of migrations
func (p *PostgreSQL) RollbackMigrations(count string) error {
	logger.Info("Starting PostgreSQL migration rollback...", zap.String("count", count))

	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(p.DB, config)

	// Rollback migrations
	if err := migrationManager.RollbackMigrations(count); err != nil {
		logger.Error("Failed to rollback PostgreSQL migrations", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	logger.Info("PostgreSQL migration rollback completed successfully")
	return nil
}

// GetMigrationStatus returns the current migration status
func (p *PostgreSQL) GetMigrationStatus() error {
	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(p.DB, config)

	// Get migration status
	if err := migrationManager.GetMigrationStatus(); err != nil {
		logger.Error("Failed to get PostgreSQL migration status", zap.Error(err))
		return err
	}

	return nil
}

// SeedData seeds the database with initial data using Laravel-style seeders
func (p *PostgreSQL) SeedData(seederName string) error {
	logger.Info("Starting PostgreSQL database seeding...")

	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(p.DB, config)

	// Run seeders
	if err := seederManager.RunSeeders(seederName); err != nil {
		logger.Error("Failed to run PostgreSQL seeders", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrSeederFailed, err)
	}

	logger.Info("PostgreSQL database seeding completed successfully")
	return nil
}

// RunSpecificSeeder runs a specific seeder
func (p *PostgreSQL) RunSpecificSeeder(seederName string) error {
	logger.Info("Running specific PostgreSQL seeder...", zap.String("seeder", seederName))

	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(p.DB, config)

	// Run specific seeder
	if err := seederManager.RunSpecificSeeder(seederName); err != nil {
		logger.Error("Failed to run specific PostgreSQL seeder", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrSeederFailed, err)
	}

	logger.Info("Specific PostgreSQL seeder completed successfully")
	return nil
}

// ListSeeders lists all registered seeders
func (p *PostgreSQL) ListSeeders() error {
	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(p.DB, config)

	// List seeders
	return seederManager.ListSeeders()
}

// HealthCheck checks the database connection health
func (p *PostgreSQL) HealthCheck() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL ping failed: %w", err)
	}

	return nil
}

// GetDatabaseStats returns database connection statistics
func (p *PostgreSQL) GetDatabaseStats() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	logger.Info("PostgreSQL Connection Statistics",
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
