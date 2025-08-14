package database

import (
	"fmt"
)

// DatabaseFactory creates database instances based on configuration
type DatabaseFactory struct{}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{}
}

// CreateDatabase creates a database instance based on the provided configuration
func (f *DatabaseFactory) CreateDatabase(config DatabaseConfig) (Database, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	switch config.GetDatabaseType() {
	case DBTypeMySQL:
		mysqlConfig, ok := config.(*MySQLConfig)
		if !ok {
			return nil, fmt.Errorf("invalid MySQL configuration type")
		}
		return NewMySQL(mysqlConfig)

	case DBTypePostgreSQL:
		postgresConfig, ok := config.(*PostgreSQLConfig)
		if !ok {
			return nil, fmt.Errorf("invalid PostgreSQL configuration type")
		}
		return NewPostgreSQL(postgresConfig)

	case DBTypeSQLite:
		sqliteConfig, ok := config.(*SQLiteConfig)
		if !ok {
			return nil, fmt.Errorf("invalid SQLite configuration type")
		}
		return NewSQLite(sqliteConfig)

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDatabaseType, config.GetDatabaseType())
	}
}

// CreateDatabaseFromType creates a database with default configuration for the specified type
func (f *DatabaseFactory) CreateDatabaseFromType(dbType DatabaseType) (Database, error) {
	var config DatabaseConfig

	switch dbType {
	case DBTypeMySQL:
		config = DefaultMySQLConfig()
	case DBTypePostgreSQL:
		config = DefaultPostgreSQLConfig()
	case DBTypeSQLite:
		config = DefaultSQLiteConfig()
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDatabaseType, dbType)
	}

	return f.CreateDatabase(config)
}

// GetSupportedDatabaseTypes returns list of supported database types
func (f *DatabaseFactory) GetSupportedDatabaseTypes() []DatabaseType {
	return []DatabaseType{DBTypeMySQL, DBTypePostgreSQL, DBTypeSQLite}
}

// ValidateDatabaseType checks if the database type is supported
func (f *DatabaseFactory) ValidateDatabaseType(dbType DatabaseType) bool {
	supportedTypes := f.GetSupportedDatabaseTypes()
	for _, supportedType := range supportedTypes {
		if dbType == supportedType {
			return true
		}
	}
	return false
}
