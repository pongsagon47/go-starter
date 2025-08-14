package database

import "fmt"

// MySQLConfig configuration for MySQL database
type MySQLConfig struct {
	BaseConfig
	SSLMode    string
	ClientCert string
	ClientKey  string
	CA         string
	Charset    string
	ParseTime  bool
	Loc        string
}

// GetDatabaseType returns the database type
func (c *MySQLConfig) GetDatabaseType() DatabaseType {
	return DBTypeMySQL
}

// Validate validates the MySQL configuration
func (c *MySQLConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}

	// MySQL specific validations
	if c.Charset == "" {
		c.Charset = "utf8mb4"
	}
	if c.Loc == "" {
		c.Loc = "Local"
	}

	return nil
}

// GetConnectionString builds the MySQL DSN
func (c *MySQLConfig) GetConnectionString() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.Charset,
		c.ParseTime,
		c.Loc,
	)

	// Add SSL configuration if provided
	if c.ClientCert != "" && c.ClientKey != "" && c.CA != "" {
		dsn += "&tls=certConfig"
	}

	return dsn
}

// DefaultMySQLConfig returns a default MySQL configuration
func DefaultMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		BaseConfig: BaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "password",
			Name:     "flex_service",
			LogLevel: "warn",
			Pool: ConnectionPoolConfig{
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 60,
			},
		},
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}
}
