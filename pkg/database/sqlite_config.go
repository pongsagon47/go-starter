package database

import "fmt"

// SQLiteConfig configuration for SQLite database
type SQLiteConfig struct {
	FilePath    string
	LogLevel    string
	Pool        ConnectionPoolConfig
	InMemory    bool
	ForeignKeys bool
	Journal     string // WAL, DELETE, TRUNCATE, PERSIST, MEMORY, OFF
	Synchronous string // OFF, NORMAL, FULL, EXTRA
	CacheSize   int    // in KB
	TempStore   string // DEFAULT, FILE, MEMORY
}

// GetDatabaseType returns the database type
func (c *SQLiteConfig) GetDatabaseType() DatabaseType {
	return DBTypeSQLite
}

// Validate validates the SQLite configuration
func (c *SQLiteConfig) Validate() error {
	if !c.InMemory && c.FilePath == "" {
		c.FilePath = "./database.db"
	}

	if c.LogLevel == "" {
		c.LogLevel = "warn"
	}

	// Set defaults
	if c.Journal == "" {
		c.Journal = "WAL"
	}
	if c.Synchronous == "" {
		c.Synchronous = "NORMAL"
	}
	if c.TempStore == "" {
		c.TempStore = "MEMORY"
	}
	if c.CacheSize == 0 {
		c.CacheSize = 10000 // 10MB
	}

	return nil
}

// GetConnectionString builds the SQLite DSN
func (c *SQLiteConfig) GetConnectionString() string {
	if c.InMemory {
		return ":memory:"
	}

	dsn := c.FilePath

	// Add pragma parameters
	params := []string{}

	if c.ForeignKeys {
		params = append(params, "foreign_keys=on")
	}

	if c.Journal != "" {
		params = append(params, fmt.Sprintf("journal_mode=%s", c.Journal))
	}

	if c.Synchronous != "" {
		params = append(params, fmt.Sprintf("synchronous=%s", c.Synchronous))
	}

	if c.CacheSize > 0 {
		params = append(params, fmt.Sprintf("cache_size=%d", c.CacheSize))
	}

	if c.TempStore != "" {
		params = append(params, fmt.Sprintf("temp_store=%s", c.TempStore))
	}

	if len(params) > 0 {
		dsn += "?"
		for i, param := range params {
			if i > 0 {
				dsn += "&"
			}
			dsn += param
		}
	}

	return dsn
}

// DefaultSQLiteConfig returns a default SQLite configuration
func DefaultSQLiteConfig() *SQLiteConfig {
	return &SQLiteConfig{
		FilePath:    "./database.db",
		LogLevel:    "warn",
		InMemory:    false,
		ForeignKeys: true,
		Journal:     "WAL",
		Synchronous: "NORMAL",
		CacheSize:   10000,
		TempStore:   "MEMORY",
		Pool: ConnectionPoolConfig{
			MaxIdleConns:    1, // SQLite works best with single connection
			MaxOpenConns:    1,
			ConnMaxLifetime: 60,
		},
	}
}
