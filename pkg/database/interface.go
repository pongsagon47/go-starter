package database

import (
	"gorm.io/gorm"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DBTypeMySQL      DatabaseType = "mysql"
	DBTypePostgreSQL DatabaseType = "postgresql"
	DBTypeSQLite     DatabaseType = "sqlite"
)

// Database interface defines the contract for all database implementations
type Database interface {
	// Connection management
	Connect() error
	GetDB() *gorm.DB
	Close() error
	HealthCheck() error
	GetDatabaseStats() error

	// Migration operations
	RunMigrations() error
	RollbackMigrations(count string) error
	GetMigrationStatus() error

	// Seeder operations
	SeedData(seederName string) error
	RunSpecificSeeder(seederName string) error
	ListSeeders() error

	// Database specific information
	GetDatabaseType() DatabaseType
	GetConnectionString() string
}

// DatabaseConfig interface for configuration
type DatabaseConfig interface {
	GetDatabaseType() DatabaseType
	Validate() error
}

// ConnectionPool configuration
type ConnectionPoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // in minutes
}

// Common database configuration fields
type BaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	LogLevel string
	Pool     ConnectionPoolConfig
}

func (c *BaseConfig) Validate() error {
	if c.Host == "" {
		return ErrInvalidHost
	}
	if c.Port <= 0 {
		return ErrInvalidPort
	}
	if c.User == "" {
		return ErrInvalidUser
	}
	if c.Name == "" {
		return ErrInvalidDatabase
	}
	return nil
}
