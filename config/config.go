package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go-starter/pkg/database"

	"github.com/joho/godotenv"
)

type Config struct {
	Database MultiDatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	Log      LogConfig
	Email    EmailConfig
	Secure   SecureConfig
	Redis    RedisConfig
	Env      string
	AppName  string
	Timezone string
}

// MultiDatabaseConfig supports multiple database configurations
type MultiDatabaseConfig struct {
	Type       database.DatabaseType // mysql, postgresql, sqlite
	MySQL      MySQLDatabaseConfig
	PostgreSQL PostgreSQLDatabaseConfig
	SQLite     SQLiteDatabaseConfig
}

// MySQLDatabaseConfig for MySQL specific settings
type MySQLDatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	LogLevel        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	ClientCert      string
	ClientKey       string
	CA              string
	Charset         string
	ParseTime       bool
	Loc             string
}

// PostgreSQLDatabaseConfig for PostgreSQL specific settings
type PostgreSQLDatabaseConfig struct {
	Host             string
	Port             int
	User             string
	Password         string
	Name             string
	LogLevel         string
	MaxIdleConns     int
	MaxOpenConns     int
	ConnMaxLifetime  int
	SSLMode          string
	TimeZone         string
	ConnectTimeout   int
	StatementTimeout int
}

// SQLiteDatabaseConfig for SQLite specific settings
type SQLiteDatabaseConfig struct {
	FilePath        string
	LogLevel        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	InMemory        bool
	ForeignKeys     bool
	Journal         string
	Synchronous     string
	CacheSize       int
	TempStore       string
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type JWTConfig struct {
	Secret                 string
	ExpirationHours        int
	RefreshExpirationHours int
	Algorithm              string
}

type LogConfig struct {
	Level  string
	Format string
}

type EmailConfig struct {
	Host               string
	Port               int
	Username           string
	Password           string
	From               string
	FromName           string
	TemplateDir        string
	MaxRetries         int
	RetryDelay         time.Duration
	InsecureSkipVerify bool
}

type SecureConfig struct {
	Key string
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Database: MultiDatabaseConfig{
			Type: database.DatabaseType(getEnv("DB_DRIVER", "mysql")),
			MySQL: MySQLDatabaseConfig{
				Host:            getEnv("DB_MYSQL_HOST", "localhost"),
				Port:            getEnvAsInt("DB_MYSQL_PORT", 3306),
				User:            getEnv("DB_MYSQL_USER", "root"),
				Password:        getEnv("DB_MYSQL_PASSWORD", "password"),
				Name:            getEnv("DB_MYSQL_NAME", "flex_service"),
				LogLevel:        getEnv("DB_MYSQL_LOG_LEVEL", "warn"),
				MaxIdleConns:    getEnvAsInt("DB_MYSQL_MAX_IDLE_CONNS", 10),
				MaxOpenConns:    getEnvAsInt("DB_MYSQL_MAX_OPEN_CONNS", 100),
				ConnMaxLifetime: getEnvAsInt("DB_MYSQL_CONN_MAX_LIFETIME", 60),
				ClientCert:      getEnv("DB_MYSQL_CLIENT_CERT", ""),
				ClientKey:       getEnv("DB_MYSQL_CLIENT_KEY", ""),
				CA:              getEnv("DB_MYSQL_CA", ""),
				Charset:         getEnv("DB_MYSQL_CHARSET", "utf8mb4"),
				ParseTime:       getEnvAsBool("DB_MYSQL_PARSE_TIME", true),
				Loc:             getEnv("DB_MYSQL_LOC", "Local"),
			},
			PostgreSQL: PostgreSQLDatabaseConfig{
				Host:             getEnv("DB_POSTGRES_HOST", "localhost"),
				Port:             getEnvAsInt("DB_POSTGRES_PORT", 5432),
				User:             getEnv("DB_POSTGRES_USER", "postgres"),
				Password:         getEnv("DB_POSTGRES_PASSWORD", "password"),
				Name:             getEnv("DB_POSTGRES_NAME", "flex_service"),
				LogLevel:         getEnv("DB_POSTGRES_LOG_LEVEL", "warn"),
				MaxIdleConns:     getEnvAsInt("DB_POSTGRES_MAX_IDLE_CONNS", 10),
				MaxOpenConns:     getEnvAsInt("DB_POSTGRES_MAX_OPEN_CONNS", 100),
				ConnMaxLifetime:  getEnvAsInt("DB_POSTGRES_CONN_MAX_LIFETIME", 60),
				SSLMode:          getEnv("DB_POSTGRES_SSL_MODE", "disable"),
				TimeZone:         getEnv("DB_POSTGRES_TIMEZONE", "UTC"),
				ConnectTimeout:   getEnvAsInt("DB_POSTGRES_CONNECT_TIMEOUT", 30),
				StatementTimeout: getEnvAsInt("DB_POSTGRES_STATEMENT_TIMEOUT", 0),
			},
			SQLite: SQLiteDatabaseConfig{
				FilePath:        getEnv("DB_SQLITE_FILE_PATH", "./database.db"),
				LogLevel:        getEnv("DB_SQLITE_LOG_LEVEL", "warn"),
				MaxIdleConns:    getEnvAsInt("DB_SQLITE_MAX_IDLE_CONNS", 1),
				MaxOpenConns:    getEnvAsInt("DB_SQLITE_MAX_OPEN_CONNS", 1),
				ConnMaxLifetime: getEnvAsInt("DB_SQLITE_CONN_MAX_LIFETIME", 60),
				InMemory:        getEnvAsBool("DB_SQLITE_IN_MEMORY", false),
				ForeignKeys:     getEnvAsBool("DB_SQLITE_FOREIGN_KEYS", true),
				Journal:         getEnv("DB_SQLITE_JOURNAL", "WAL"),
				Synchronous:     getEnv("DB_SQLITE_SYNCHRONOUS", "NORMAL"),
				CacheSize:       getEnvAsInt("DB_SQLITE_CACHE_SIZE", 10000),
				TempStore:       getEnv("DB_SQLITE_TEMP_STORE", "MEMORY"),
			},
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			ExpirationHours:        getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpirationHours: getEnvAsInt("JWT_REFRESH_EXPIRATION_HOURS", 720),
			Algorithm:              getEnv("JWT_ALGORITHM", "HS256"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Email: EmailConfig{
			Host:               getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:               getEnvAsInt("SMTP_PORT", 587),
			Username:           getEnv("SMTP_USERNAME", ""),
			Password:           getEnv("SMTP_PASSWORD", ""),
			From:               getEnv("SMTP_FROM", ""),
			FromName:           getEnv("SMTP_FROM_NAME", "Go Clean Gin"),
			TemplateDir:        getEnv("EMAIL_TEMPLATE_DIR", "./templates"),
			MaxRetries:         getEnvAsInt("EMAIL_MAX_RETRIES", 3),
			RetryDelay:         getEnvAsDuration("EMAIL_RETRY_DELAY", 1*time.Second),
			InsecureSkipVerify: getEnvAsBool("EMAIL_INSECURE_SKIP_VERIFY", false),
		},
		Secure: SecureConfig{
			Key: getEnv("ENCRYPTION_KEY", ""),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvAsInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Env:      getEnv("ENV", "development"),
		AppName:  getEnv("APP_NAME", "go-starter"),
		Timezone: getEnv("TIMEZONE", "Asia/Bangkok"),
	}
}

// GetDatabaseConfig returns the appropriate database configuration based on the selected type
func (c *Config) GetDatabaseConfig() database.DatabaseConfig {
	switch c.Database.Type {
	case database.DBTypeMySQL:
		return c.convertToMySQLConfig()
	case database.DBTypePostgreSQL:
		return c.convertToPostgreSQLConfig()
	case database.DBTypeSQLite:
		return c.convertToSQLiteConfig()
	default:
		// Default to MySQL if invalid type
		log.Printf("Warning: Unknown database type '%s', defaulting to MySQL", c.Database.Type)
		c.Database.Type = database.DBTypeMySQL
		return c.convertToMySQLConfig()
	}
}

// Convert config structs to database package configs
func (c *Config) convertToMySQLConfig() *database.MySQLConfig {
	mysql := c.Database.MySQL
	return &database.MySQLConfig{
		BaseConfig: database.BaseConfig{
			Host:     mysql.Host,
			Port:     mysql.Port,
			User:     mysql.User,
			Password: mysql.Password,
			Name:     mysql.Name,
			LogLevel: mysql.LogLevel,
			Pool: database.ConnectionPoolConfig{
				MaxIdleConns:    mysql.MaxIdleConns,
				MaxOpenConns:    mysql.MaxOpenConns,
				ConnMaxLifetime: mysql.ConnMaxLifetime,
			},
		},
		ClientCert: mysql.ClientCert,
		ClientKey:  mysql.ClientKey,
		CA:         mysql.CA,
		Charset:    mysql.Charset,
		ParseTime:  mysql.ParseTime,
		Loc:        mysql.Loc,
	}
}

func (c *Config) convertToPostgreSQLConfig() *database.PostgreSQLConfig {
	postgres := c.Database.PostgreSQL
	return &database.PostgreSQLConfig{
		BaseConfig: database.BaseConfig{
			Host:     postgres.Host,
			Port:     postgres.Port,
			User:     postgres.User,
			Password: postgres.Password,
			Name:     postgres.Name,
			LogLevel: postgres.LogLevel,
			Pool: database.ConnectionPoolConfig{
				MaxIdleConns:    postgres.MaxIdleConns,
				MaxOpenConns:    postgres.MaxOpenConns,
				ConnMaxLifetime: postgres.ConnMaxLifetime,
			},
		},
		SSLMode:          postgres.SSLMode,
		TimeZone:         postgres.TimeZone,
		ConnectTimeout:   postgres.ConnectTimeout,
		StatementTimeout: postgres.StatementTimeout,
	}
}

func (c *Config) convertToSQLiteConfig() *database.SQLiteConfig {
	sqlite := c.Database.SQLite
	return &database.SQLiteConfig{
		FilePath:    sqlite.FilePath,
		LogLevel:    sqlite.LogLevel,
		InMemory:    sqlite.InMemory,
		ForeignKeys: sqlite.ForeignKeys,
		Journal:     sqlite.Journal,
		Synchronous: sqlite.Synchronous,
		CacheSize:   sqlite.CacheSize,
		TempStore:   sqlite.TempStore,
		Pool: database.ConnectionPoolConfig{
			MaxIdleConns:    sqlite.MaxIdleConns,
			MaxOpenConns:    sqlite.MaxOpenConns,
			ConnMaxLifetime: sqlite.ConnMaxLifetime,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}
