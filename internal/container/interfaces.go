package container

import (
	"context"

	pkgAuth "go-starter/pkg/auth"
	"go-starter/pkg/cache"
	"go-starter/pkg/database"
	"go-starter/pkg/mail"
	"go-starter/pkg/secure"
)

// ContainerInterface defines the main container contract
type ContainerInterface interface {
	// Core dependencies
	GetDatabase() database.Database
	GetCache() cache.Cache
	GetMail() *mail.Mailer
	GetSecure() *secure.Secure
	GetJWT() *pkgAuth.JWT

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
	RegisterUser() error
	RegisterProduct() error
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

type JWTFactory interface {
	CreateJWT() (*pkgAuth.JWT, error)
}

// CompositeFactor combines all factories
type CompositeFactory interface {
	DatabaseFactory
	CacheFactory
	MailFactory
	SecureFactory
	JWTFactory
}
