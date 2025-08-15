# 🗄️ Database Package

Multi-database support with unified interface for MySQL, PostgreSQL, and SQLite using GORM.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Database Interface](#database-interface)
- [Supported Databases](#supported-databases)
- [Configuration](#configuration)
- [Factory Pattern](#factory-pattern)
- [Examples](#examples)
- [Migration Integration](#migration-integration)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/database"
```

## ⚡ Quick Start

### Basic Usage

```go
package main

import (
    "go-starter/pkg/database"
    "go-starter/config"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Create database using factory
    factory := database.NewDatabaseFactory()
    dbConfig := cfg.GetDatabaseConfig()

    db, err := factory.CreateDatabase(dbConfig)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Use database
    gormDB := db.GetDB()

    // Your database operations...
    var users []User
    gormDB.Find(&users)
}
```

## 🔧 Database Interface

### Unified Database Interface

```go
type Database interface {
    // Core operations
    GetDB() *gorm.DB
    Close() error
    HealthCheck() error

    // Database info
    GetDatabaseType() DatabaseType
    GetConnectionString() string

    // Migration operations
    RunMigrations() error
    RollbackMigrations(count string) error
    GetMigrationStatus() error

    // Seeder operations
    SeedData(seederName string) error
    RunSpecificSeeder(seederName string) error
    ListSeeders() error
}
```

### Database Types

```go
type DatabaseType string

const (
    DBTypeMySQL      DatabaseType = "mysql"
    DBTypePostgreSQL DatabaseType = "postgresql"
    DBTypeSQLite     DatabaseType = "sqlite"
)
```

## 🗄️ Supported Databases

### 1. **MySQL**

```go
// Configuration
type MySQLConfig struct {
    Host            string // Database host
    Port            int    // Database port
    User            string // Database user
    Password        string // Database password
    Name            string // Database name
    LogLevel        string // Log level
    MaxIdleConns    int    // Max idle connections
    MaxOpenConns    int    // Max open connections
    ConnMaxLifetime int    // Connection max lifetime (minutes)
    ClientCert      string // SSL client certificate
    ClientKey       string // SSL client key
    CA              string // SSL CA certificate
    Charset         string // Character set
    ParseTime       bool   // Parse time
    Loc             string // Time zone location
}
```

### 2. **PostgreSQL**

```go
// Configuration
type PostgreSQLConfig struct {
    Host               string // Database host
    Port               int    // Database port
    User               string // Database user
    Password           string // Database password
    Name               string // Database name
    LogLevel           string // Log level
    MaxIdleConns       int    // Max idle connections
    MaxOpenConns       int    // Max open connections
    ConnMaxLifetime    int    // Connection max lifetime (minutes)
    SSLMode            string // SSL mode
    Timezone           string // Time zone
    ConnectTimeout     int    // Connection timeout (seconds)
    StatementTimeout   int    // Statement timeout (seconds)
}
```

### 3. **SQLite**

```go
// Configuration
type SQLiteConfig struct {
    FilePath        string // Database file path
    LogLevel        string // Log level
    MaxIdleConns    int    // Max idle connections
    MaxOpenConns    int    // Max open connections
    ConnMaxLifetime int    // Connection max lifetime (minutes)
    InMemory        bool   // Use in-memory database
    ForeignKeys     bool   // Enable foreign keys
    JournalMode     string // Journal mode (WAL, DELETE, etc.)
    Synchronous     string // Synchronous mode
    CacheSize       int    // Cache size in KB
    TempStore       string // Temporary storage mode
}
```

## ⚙️ Configuration

### Environment Variables

```env
# Database type selection
DB_DRIVER=mysql  # mysql, postgresql, sqlite

# MySQL configuration
DB_MYSQL_HOST=localhost
DB_MYSQL_PORT=3306
DB_MYSQL_USER=root
DB_MYSQL_PASSWORD=password
DB_MYSQL_NAME=my_app
DB_MYSQL_SSL_MODE=disable

# PostgreSQL configuration
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=password
DB_POSTGRES_NAME=my_app
DB_POSTGRES_SSL_MODE=disable

# SQLite configuration
DB_SQLITE_FILE_PATH=./database.db
DB_SQLITE_FOREIGN_KEYS=true
DB_SQLITE_JOURNAL=WAL
```

### Database Configuration Interface

```go
type DatabaseConfig interface {
    GetDatabaseType() DatabaseType
    GetConnectionString() string
    GetLogLevel() string
    GetMaxIdleConns() int
    GetMaxOpenConns() int
    GetConnMaxLifetime() time.Duration
}
```

## 🏭 Factory Pattern

### Database Factory

```go
type DatabaseFactory struct{}

func NewDatabaseFactory() *DatabaseFactory {
    return &DatabaseFactory{}
}

func (f *DatabaseFactory) CreateDatabase(config DatabaseConfig) (Database, error) {
    switch config.GetDatabaseType() {
    case DBTypeMySQL:
        return NewMySQL(config.(*MySQLDatabaseConfig))
    case DBTypePostgreSQL:
        return NewPostgreSQL(config.(*PostgreSQLDatabaseConfig))
    case DBTypeSQLite:
        return NewSQLite(config.(*SQLiteDatabaseConfig))
    default:
        return nil, ErrUnsupportedDatabaseType
    }
}
```

### Usage with Factory

```go
// Create factory
factory := database.NewDatabaseFactory()

// Get configuration
dbConfig := config.GetDatabaseConfig()

// Create database instance
db, err := factory.CreateDatabase(dbConfig)
if err != nil {
    return err
}

// Use database
gormDB := db.GetDB()
```

## 💡 Examples

### Switching Between Databases

```go
// Development with SQLite
os.Setenv("DB_DRIVER", "sqlite")
cfg := config.Load()
db, _ := factory.CreateDatabase(cfg.GetDatabaseConfig())

// Staging with MySQL
os.Setenv("DB_DRIVER", "mysql")
cfg = config.Load()
db, _ = factory.CreateDatabase(cfg.GetDatabaseConfig())

// Production with PostgreSQL
os.Setenv("DB_DRIVER", "postgresql")
cfg = config.Load()
db, _ = factory.CreateDatabase(cfg.GetDatabaseConfig())
```

### Repository Pattern

```go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *User) error {
    return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id string) (*User, error) {
    var user User
    err := r.db.First(&user, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Update(user *User) error {
    return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id string) error {
    return r.db.Delete(&User{}, "id = ?", id).Error
}

// Usage with any database
func main() {
    // Database is created based on configuration
    db, _ := factory.CreateDatabase(config.GetDatabaseConfig())

    // Repository works with any database
    userRepo := NewUserRepository(db.GetDB())

    user := &User{Name: "John", Email: "john@example.com"}
    userRepo.Create(user)
}
```

### Transaction Example

```go
func TransferMoney(db *gorm.DB, fromUserID, toUserID string, amount float64) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // Deduct from sender
        if err := tx.Model(&User{}).Where("id = ?", fromUserID).
            Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }

        // Add to receiver
        if err := tx.Model(&User{}).Where("id = ?", toUserID).
            Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
            return err
        }

        // Create transaction record
        transaction := &Transaction{
            FromUserID: fromUserID,
            ToUserID:   toUserID,
            Amount:     amount,
            Status:     "completed",
        }

        return tx.Create(transaction).Error
    })
}
```

### Health Check Example

```go
func DatabaseHealthCheck(db database.Database) error {
    // Check database connection
    if err := db.HealthCheck(); err != nil {
        return fmt.Errorf("database health check failed: %w", err)
    }

    // Check if we can perform a simple query
    var count int64
    if err := db.GetDB().Raw("SELECT 1").Count(&count).Error; err != nil {
        return fmt.Errorf("database query test failed: %w", err)
    }

    return nil
}
```

## 🔄 Migration Integration

### Running Migrations

```go
// Automatic migration running
if err := db.RunMigrations(); err != nil {
    log.Fatal("Migration failed:", err)
}

// Check migration status
if err := db.GetMigrationStatus(); err != nil {
    log.Error("Failed to get migration status:", err)
}

// Rollback migrations
if err := db.RollbackMigrations("1"); err != nil {
    log.Error("Rollback failed:", err)
}
```

### Seeder Integration

```go
// Run all seeders
if err := db.SeedData(""); err != nil {
    log.Error("Seeding failed:", err)
}

// Run specific seeder
if err := db.RunSpecificSeeder("UserSeeder"); err != nil {
    log.Error("UserSeeder failed:", err)
}

// List available seeders
if err := db.ListSeeders(); err != nil {
    log.Error("Failed to list seeders:", err)
}
```

## 🎯 Best Practices

### 1. **Environment-based Configuration**

```go
// Use different databases for different environments
switch os.Getenv("ENV") {
case "development":
    os.Setenv("DB_DRIVER", "sqlite")
case "staging":
    os.Setenv("DB_DRIVER", "mysql")
case "production":
    os.Setenv("DB_DRIVER", "postgresql")
}
```

### 2. **Connection Pool Configuration**

```go
// Optimize connection pools for each database
mysqlConfig := &MySQLConfig{
    MaxIdleConns:    10,  // Keep 10 idle connections
    MaxOpenConns:    100, // Max 100 open connections
    ConnMaxLifetime: 60,  // 60 minutes max lifetime
}

postgresConfig := &PostgreSQLConfig{
    MaxIdleConns:    25,  // PostgreSQL handles more idle connections
    MaxOpenConns:    200, // Higher concurrency
    ConnMaxLifetime: 120, // Longer lifetime
}

sqliteConfig := &SQLiteConfig{
    MaxIdleConns: 1,  // SQLite doesn't benefit from multiple connections
    MaxOpenConns: 1,  // Single connection for SQLite
}
```

### 3. **Error Handling**

```go
func HandleDatabaseError(err error) {
    switch {
    case errors.Is(err, gorm.ErrRecordNotFound):
        // Handle record not found
        log.Info("Record not found")
    case errors.Is(err, gorm.ErrInvalidTransaction):
        // Handle transaction error
        log.Error("Transaction error:", err)
    case strings.Contains(err.Error(), "connection refused"):
        // Handle connection error
        log.Error("Database connection failed:", err)
    default:
        // Handle other errors
        log.Error("Database error:", err)
    }
}
```

### 4. **Database-Specific Optimizations**

#### MySQL Optimizations

```go
// Use InnoDB engine
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"index"` // Add indexes for frequently queried fields
    // Use GORM tags for MySQL-specific features
}

// Full-text search
db.Raw("SELECT * FROM users WHERE MATCH(name, email) AGAINST(? IN NATURAL LANGUAGE MODE)", query).Scan(&users)
```

#### PostgreSQL Optimizations

```go
// Use PostgreSQL-specific features
type Product struct {
    ID         uint           `gorm:"primaryKey"`
    Attributes postgres.Jsonb `gorm:"type:jsonb"` // JSONB for flexible data
    Tags       pq.StringArray `gorm:"type:text[]"` // Arrays
}

// JSONB queries
db.Where("attributes @> ?", `{"color": "red"}`).Find(&products)
```

#### SQLite Optimizations

```go
// SQLite-specific settings
func optimizeSQLite(db *gorm.DB) {
    db.Exec("PRAGMA journal_mode=WAL")
    db.Exec("PRAGMA synchronous=NORMAL")
    db.Exec("PRAGMA cache_size=10000")
    db.Exec("PRAGMA foreign_keys=ON")
}
```

### 5. **Testing with Multiple Databases**

```go
func TestUserRepository(t *testing.T) {
    databases := []struct {
        name string
        config DatabaseConfig
    }{
        {"SQLite", &SQLiteConfig{InMemory: true}},
        {"MySQL", &MySQLConfig{Host: "localhost", Name: "test_db"}},
        {"PostgreSQL", &PostgreSQLConfig{Host: "localhost", Name: "test_db"}},
    }

    for _, dbConfig := range databases {
        t.Run(dbConfig.name, func(t *testing.T) {
            db, err := factory.CreateDatabase(dbConfig.config)
            require.NoError(t, err)
            defer db.Close()

            // Run tests with this database
            testUserOperations(t, db.GetDB())
        })
    }
}
```

## 🚨 Error Handling

### Common Errors

```go
var (
    ErrUnsupportedDatabaseType = errors.New("unsupported database type")
    ErrConnectionFailed        = errors.New("database connection failed")
    ErrMigrationFailed         = errors.New("migration failed")
    ErrSeederFailed           = errors.New("seeder failed")
)
```

### Database-Specific Errors

```go
func HandleDatabaseSpecificErrors(dbType DatabaseType, err error) {
    switch dbType {
    case DBTypeMySQL:
        if mysqlErr, ok := err.(*mysql.MySQLError); ok {
            switch mysqlErr.Number {
            case 1062: // Duplicate entry
                log.Error("Duplicate entry error")
            case 1146: // Table doesn't exist
                log.Error("Table not found")
            }
        }
    case DBTypePostgreSQL:
        if pqErr, ok := err.(*pq.Error); ok {
            switch pqErr.Code {
            case "23505": // Unique violation
                log.Error("Unique constraint violation")
            case "42P01": // Undefined table
                log.Error("Table not found")
            }
        }
    case DBTypeSQLite:
        if strings.Contains(err.Error(), "UNIQUE constraint failed") {
            log.Error("Unique constraint violation")
        }
    }
}
```

## 🔗 Related Packages

- [`pkg/migration`](../migration/) - Database migration system
- [`pkg/seeder`](../seeder/) - Database seeder system
- [`internal/entity`](../../internal/entity/) - Database models
- [`config`](../../config/) - Database configuration

## 📚 Additional Resources

- [GORM Documentation](https://gorm.io/docs/)
- [MySQL Documentation](https://dev.mysql.com/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
