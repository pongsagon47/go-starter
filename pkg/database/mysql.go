package database

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"go-starter/pkg/logger"
	"go-starter/pkg/migration"
	"go-starter/pkg/seeder"

	mysqldriver "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type MySQL struct {
	DB     *gorm.DB
	config *MySQLConfig
}

// NewMySQL creates a new MySQL database connection
func NewMySQL(config *MySQLConfig) (*MySQL, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid MySQL configuration: %w", err)
	}

	// Handle SSL configuration if provided
	var dsn string
	if config.ClientCert != "" && config.ClientKey != "" && config.CA != "" {
		if err := setupMySQLTLS(config); err != nil {
			return nil, fmt.Errorf("failed to setup MySQL TLS: %w", err)
		}
		dsn = config.GetConnectionString()
	} else {
		dsn = config.GetConnectionString()
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
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		logger.Error("Failed to connect to MySQL database", zap.Error(err))
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

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping MySQL database", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	logger.Info("Successfully connected to MySQL database",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Name),
		zap.Int("max_idle_conns", config.Pool.MaxIdleConns),
		zap.Int("max_open_conns", config.Pool.MaxOpenConns))

	return &MySQL{
		DB:     db,
		config: config,
	}, nil
}

// setupMySQLTLS configures TLS for MySQL connection
func setupMySQLTLS(config *MySQLConfig) error {
	wd, _ := os.Getwd()
	clientCertPath := wd + config.ClientCert
	clientKeyPath := wd + config.ClientKey
	caPath := wd + config.CA

	rootCert, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM(rootCert); !ok {
		return fmt.Errorf("failed to append CA certificate")
	}

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:            rootCertPool,
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: true,
	}

	return mysqldriver.RegisterTLSConfig("certConfig", tlsConfig)
}

// Connect establishes database connection (already done in constructor)
func (m *MySQL) Connect() error {
	return m.HealthCheck()
}

// GetDB returns the GORM database instance
func (m *MySQL) GetDB() *gorm.DB {
	return m.DB
}

// Close closes the database connection
func (m *MySQL) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDatabaseType returns the database type
func (m *MySQL) GetDatabaseType() DatabaseType {
	return DBTypeMySQL
}

// GetConnectionString returns the connection string (masked password)
func (m *MySQL) GetConnectionString() string {
	maskedConfig := *m.config
	maskedConfig.Password = "****"
	return maskedConfig.GetConnectionString()
}

// RunMigrations runs database migrations using Laravel-style migration system
func (m *MySQL) RunMigrations() error {
	logger.Info("Starting MySQL migrations...")

	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(m.DB, config)

	// Run migrations
	if err := migrationManager.RunMigrations(); err != nil {
		logger.Error("Failed to run MySQL migrations", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	logger.Info("MySQL migrations completed successfully")
	return nil
}

// RollbackMigrations rolls back the specified number of migrations
func (m *MySQL) RollbackMigrations(count string) error {
	logger.Info("Starting MySQL migration rollback...", zap.String("count", count))

	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(m.DB, config)

	// Rollback migrations
	if err := migrationManager.RollbackMigrations(count); err != nil {
		logger.Error("Failed to rollback MySQL migrations", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}

	logger.Info("MySQL migration rollback completed successfully")
	return nil
}

// GetMigrationStatus returns the current migration status
func (m *MySQL) GetMigrationStatus() error {
	// Create migration manager with global migrations
	config := migration.DefaultMigrationConfig()
	migrationManager := migration.NewManagerWithGlobalMigrations(m.DB, config)

	// Get migration status
	if err := migrationManager.GetMigrationStatus(); err != nil {
		logger.Error("Failed to get MySQL migration status", zap.Error(err))
		return err
	}

	return nil
}

// SeedData seeds the database with initial data using Laravel-style seeders
func (m *MySQL) SeedData(seederName string) error {
	logger.Info("Starting MySQL database seeding...")

	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(m.DB, config)

	// Run seeders
	if err := seederManager.RunSeeders(seederName); err != nil {
		logger.Error("Failed to run MySQL seeders", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrSeederFailed, err)
	}

	logger.Info("MySQL database seeding completed successfully")
	return nil
}

// RunSpecificSeeder runs a specific seeder
func (m *MySQL) RunSpecificSeeder(seederName string) error {
	logger.Info("Running specific MySQL seeder...", zap.String("seeder", seederName))

	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(m.DB, config)

	// Run specific seeder
	if err := seederManager.RunSpecificSeeder(seederName); err != nil {
		logger.Error("Failed to run specific MySQL seeder", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrSeederFailed, err)
	}

	logger.Info("Specific MySQL seeder completed successfully")
	return nil
}

// ListSeeders lists all registered seeders
func (m *MySQL) ListSeeders() error {
	// Create seeder manager with global seeders
	config := seeder.DefaultSeederConfig()
	seederManager := seeder.NewManagerWithGlobalSeeders(m.DB, config)

	// List seeders
	return seederManager.ListSeeders()
}

// HealthCheck checks the database connection health
func (m *MySQL) HealthCheck() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("MySQL ping failed: %w", err)
	}

	return nil
}

// GetDatabaseStats returns database connection statistics
func (m *MySQL) GetDatabaseStats() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	logger.Info("MySQL Connection Statistics",
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
