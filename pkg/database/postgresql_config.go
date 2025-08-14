package database

import "fmt"

// PostgreSQLConfig configuration for PostgreSQL database
type PostgreSQLConfig struct {
	BaseConfig
	SSLMode          string
	TimeZone         string
	ConnectTimeout   int
	StatementTimeout int
}

// GetDatabaseType returns the database type
func (c *PostgreSQLConfig) GetDatabaseType() DatabaseType {
	return DBTypePostgreSQL
}

// Validate validates the PostgreSQL configuration
func (c *PostgreSQLConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}

	// PostgreSQL specific validations
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	if c.TimeZone == "" {
		c.TimeZone = "UTC"
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 30
	}

	return nil
}

// GetConnectionString builds the PostgreSQL DSN
func (c *PostgreSQLConfig) GetConnectionString() string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode,
		c.TimeZone,
	)

	if c.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", c.ConnectTimeout)
	}

	if c.StatementTimeout > 0 {
		dsn += fmt.Sprintf(" statement_timeout=%dms", c.StatementTimeout)
	}

	return dsn
}

// DefaultPostgreSQLConfig returns a default PostgreSQL configuration
func DefaultPostgreSQLConfig() *PostgreSQLConfig {
	return &PostgreSQLConfig{
		BaseConfig: BaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "flex_service",
			LogLevel: "warn",
			Pool: ConnectionPoolConfig{
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
		},
		SSLMode:        "disable",
		TimeZone:       "UTC",
		ConnectTimeout: 30,
	}
}
