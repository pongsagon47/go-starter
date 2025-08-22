package container

import (
	"context"

	"flex-service/pkg/cache"
	"flex-service/pkg/database"
	"flex-service/pkg/mail"
	"flex-service/pkg/secure"
)

// ContainerInterface defines the main container contract
type ContainerInterface interface {
	// Core dependencies
	GetDatabase() database.Database
	GetCache() cache.Cache
	GetMail() *mail.Mailer
	GetSecure() *secure.Secure

	// Lifecycle methods
	RunMigrations() error
	SeedData(seederName string) error
	Close() error
	HealthCheck(ctx context.Context) error

	// Database utilities
	GetDatabaseType() database.DatabaseType
}

// ServiceRegistryInterface for registering application services
type ServiceRegistryInterface interface {
	RegisterAuth() error
	// RegisterUser() error
	// RegisterProduct() error
	// Add more service registration methods as needed
}

// Factory interfaces for creating dependencies
type DatabaseFactory interface {
	CreateDatabase() (database.Database, error)
}

type CacheFactory interface {
	CreateCache() (cache.Cache, error)
}

type MailFactory interface {
	CreateMailer() (*mail.Mailer, error)
}

type SecureFactory interface {
	CreateSecure() (*secure.Secure, error)
}

// CompositeFactor combines all factories
type CompositeFactory interface {
	DatabaseFactory
	CacheFactory
	MailFactory
	SecureFactory
}
