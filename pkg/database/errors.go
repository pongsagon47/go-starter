package database

import "errors"

// Database-specific errors
var (
	ErrUnsupportedDatabaseType = errors.New("unsupported database type")
	ErrInvalidHost             = errors.New("invalid database host")
	ErrInvalidPort             = errors.New("invalid database port")
	ErrInvalidUser             = errors.New("invalid database user")
	ErrInvalidDatabase         = errors.New("invalid database name")
	ErrConnectionFailed        = errors.New("database connection failed")
	ErrMigrationFailed         = errors.New("migration failed")
	ErrSeederFailed            = errors.New("seeder failed")
)
