# 🔄 Migration Package

Laravel-style database migration system with version control, rollback support, and multi-database compatibility.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Migration Interface](#migration-interface)
- [Creating Migrations](#creating-migrations)
- [Running Migrations](#running-migrations)
- [Rollback System](#rollback-system)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in flex-service
import "flex-service/pkg/migration"
```

## ⚡ Quick Start

### Creating a Migration

```bash
# Using Artisan CLI
make make-migration NAME=create_users_table CREATE=true TABLE=users

# This creates: internal/migrations/YYYY_MM_DD_HHMMSS_create_users_table.go
```

### Basic Migration Structure

```go
package migrations

import (
    "flex-service/pkg/migration"
    "gorm.io/gorm"
)

type CreateUsersTable20250101000001 struct{}

func (m *CreateUsersTable20250101000001) Up(db *gorm.DB) error {
    type User struct {
        ID        uint      `gorm:"primaryKey"`
        Name      string    `gorm:"size:255;not null"`
        Email     string    `gorm:"size:255;uniqueIndex;not null"`
        CreatedAt time.Time
        UpdatedAt time.Time
    }

    return db.AutoMigrate(&User{})
}

func (m *CreateUsersTable20250101000001) Down(db *gorm.DB) error {
    return db.Migrator().DropTable("users")
}

func (m *CreateUsersTable20250101000001) Description() string {
    return "Create users table with basic fields"
}

func (m *CreateUsersTable20250101000001) Version() string {
    return "2025_01_01_000001"
}

func init() {
    migration.Register(&CreateUsersTable20250101000001{})
}
```

## 🔧 Migration Interface

### Migration Interface

```go
type Migration interface {
    Up(db *gorm.DB) error      // Apply migration
    Down(db *gorm.DB) error    // Rollback migration
    Description() string       // Human-readable description
    Version() string          // Unique version identifier
}
```

### Migration Manager

```go
type Manager struct {
    db           *gorm.DB
    config       *MigrationConfig
    migrations   []Migration
    tableName    string
}

// Key methods
func (m *Manager) RunMigrations() error
func (m *Manager) RollbackMigrations(count string) error
func (m *Manager) GetMigrationStatus() error
```

### Configuration

```go
type MigrationConfig struct {
    TableName string // Migration tracking table name
    BatchSize int    // Batch size for migrations
}

func DefaultMigrationConfig() *MigrationConfig {
    return &MigrationConfig{
        TableName: "migrations",
        BatchSize: 100,
    }
}
```

## 📝 Creating Migrations

### 1. **Using Artisan CLI (Recommended)**

```bash
# Create table migration
make make-migration NAME=create_products_table CREATE=true TABLE=products

# Add column migration
make make-migration NAME=add_price_to_products_table TABLE=products

# General migration
make make-migration NAME=update_user_indexes
```

### 2. **Manual Creation**

```go
package migrations

import (
    "flex-service/pkg/migration"
    "gorm.io/gorm"
    "time"
)

type AddPriceToProductsTable20250101000002 struct{}

func (m *AddPriceToProductsTable20250101000002) Up(db *gorm.DB) error {
    // Add new column
    return db.Exec("ALTER TABLE products ADD COLUMN price DECIMAL(10,2) DEFAULT 0").Error
}

func (m *AddPriceToProductsTable20250101000002) Down(db *gorm.DB) error {
    // Remove column
    return db.Migrator().DropColumn("products", "price")
}

func (m *AddPriceToProductsTable20250101000002) Description() string {
    return "Add price column to products table"
}

func (m *AddPriceToProductsTable20250101000002) Version() string {
    return "2025_01_01_000002"
}

func init() {
    migration.Register(&AddPriceToProductsTable20250101000002{})
}
```

## 🚀 Running Migrations

### Command Line

```bash
# Run all pending migrations
make migrate

# Check migration status
make migrate-status

# Rollback last migration
make migrate-rollback

# Rollback specific number of migrations
make migrate-rollback COUNT=3
```

### Programmatic Usage

```go
// Create migration manager
config := migration.DefaultMigrationConfig()
manager := migration.NewManagerWithGlobalMigrations(db, config)

// Run all pending migrations
if err := manager.RunMigrations(); err != nil {
    log.Fatal("Migration failed:", err)
}

// Check status
if err := manager.GetMigrationStatus(); err != nil {
    log.Error("Failed to get status:", err)
}

// Rollback migrations
if err := manager.RollbackMigrations("2"); err != nil {
    log.Error("Rollback failed:", err)
}
```

## ↩️ Rollback System

### Rollback Strategies

```go
// Rollback last migration
manager.RollbackMigrations("1")

// Rollback last 3 migrations
manager.RollbackMigrations("3")

// Rollback all migrations
manager.RollbackMigrations("all")
```

### Safe Rollback Patterns

```go
func (m *AddColumnMigration) Up(db *gorm.DB) error {
    return db.Exec("ALTER TABLE users ADD COLUMN phone VARCHAR(20)").Error
}

func (m *AddColumnMigration) Down(db *gorm.DB) error {
    // Safe rollback - check if column exists first
    if db.Migrator().HasColumn("users", "phone") {
        return db.Migrator().DropColumn("users", "phone")
    }
    return nil
}
```

## 💡 Examples

### 1. **Create Table Migration**

```go
type CreateOrdersTable20250101000003 struct{}

func (m *CreateOrdersTable20250101000003) Up(db *gorm.DB) error {
    type Order struct {
        ID          uint      `gorm:"primaryKey"`
        UserID      uint      `gorm:"not null;index"`
        TotalAmount float64   `gorm:"type:decimal(10,2);not null"`
        Status      string    `gorm:"size:50;not null;default:'pending'"`
        OrderDate   time.Time `gorm:"not null"`
        CreatedAt   time.Time
        UpdatedAt   time.Time

        // Foreign key constraint
        User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
    }

    return db.AutoMigrate(&Order{})
}

func (m *CreateOrdersTable20250101000003) Down(db *gorm.DB) error {
    return db.Migrator().DropTable("orders")
}
```

### 2. **Add Index Migration**

```go
type AddIndexesToUsersTable20250101000004 struct{}

func (m *AddIndexesToUsersTable20250101000004) Up(db *gorm.DB) error {
    // Add composite index
    if err := db.Exec("CREATE INDEX idx_users_name_email ON users(name, email)").Error; err != nil {
        return err
    }

    // Add partial index (PostgreSQL)
    if db.Dialector.Name() == "postgres" {
        return db.Exec("CREATE INDEX idx_users_active ON users(id) WHERE active = true").Error
    }

    return nil
}

func (m *AddIndexesToUsersTable20250101000004) Down(db *gorm.DB) error {
    // Drop indexes
    db.Exec("DROP INDEX IF EXISTS idx_users_name_email")
    if db.Dialector.Name() == "postgres" {
        db.Exec("DROP INDEX IF EXISTS idx_users_active")
    }
    return nil
}
```

### 3. **Data Migration**

```go
type UpdateUserRoles20250101000005 struct{}

func (m *UpdateUserRoles20250101000005) Up(db *gorm.DB) error {
    // Update existing data
    return db.Exec("UPDATE users SET role = 'user' WHERE role IS NULL OR role = ''").Error
}

func (m *UpdateUserRoles20250101000005) Down(db *gorm.DB) error {
    // Revert data changes (if possible)
    return db.Exec("UPDATE users SET role = NULL WHERE role = 'user'").Error
}
```

### 4. **Database-Specific Migration**

```go
type AddFullTextSearchToProducts20250101000006 struct{}

func (m *AddFullTextSearchToProducts20250101000006) Up(db *gorm.DB) error {
    switch db.Dialector.Name() {
    case "mysql":
        // MySQL full-text index
        return db.Exec("ALTER TABLE products ADD FULLTEXT(name, description)").Error
    case "postgres":
        // PostgreSQL GIN index for full-text search
        return db.Exec("CREATE INDEX idx_products_fulltext ON products USING GIN(to_tsvector('english', name || ' ' || description))").Error
    case "sqlite":
        // SQLite FTS virtual table
        return db.Exec("CREATE VIRTUAL TABLE products_fts USING fts5(name, description, content=products)").Error
    }
    return nil
}

func (m *AddFullTextSearchToProducts20250101000006) Down(db *gorm.DB) error {
    switch db.Dialector.Name() {
    case "mysql":
        return db.Exec("ALTER TABLE products DROP INDEX name").Error // MySQL auto-names fulltext indexes
    case "postgres":
        return db.Exec("DROP INDEX IF EXISTS idx_products_fulltext").Error
    case "sqlite":
        return db.Exec("DROP TABLE IF EXISTS products_fts").Error
    }
    return nil
}
```

### 5. **Complex Schema Changes**

```go
type RefactorUserProfileTable20250101000007 struct{}

func (m *RefactorUserProfileTable20250101000007) Up(db *gorm.DB) error {
    // Step 1: Create new table structure
    type UserProfile struct {
        ID          uint   `gorm:"primaryKey"`
        UserID      uint   `gorm:"uniqueIndex;not null"`
        FirstName   string `gorm:"size:100"`
        LastName    string `gorm:"size:100"`
        Bio         text   `gorm:"type:text"`
        Avatar      string `gorm:"size:255"`
        CreatedAt   time.Time
        UpdatedAt   time.Time
    }

    if err := db.AutoMigrate(&UserProfile{}); err != nil {
        return err
    }

    // Step 2: Migrate data from old structure
    return db.Exec(`
        INSERT INTO user_profiles (user_id, first_name, last_name, bio, created_at, updated_at)
        SELECT id, SUBSTRING_INDEX(name, ' ', 1), SUBSTRING_INDEX(name, ' ', -1), bio, created_at, updated_at
        FROM users
        WHERE name IS NOT NULL
    `).Error
}

func (m *RefactorUserProfileTable20250101000007) Down(db *gorm.DB) error {
    // Revert changes
    return db.Migrator().DropTable("user_profiles")
}
```

## 🎯 Best Practices

### 1. **Migration Naming**

```bash
# Good naming conventions
2025_01_01_000001_create_users_table.go
2025_01_01_000002_add_email_index_to_users.go
2025_01_01_000003_update_user_roles_data.go

# Bad naming
migration1.go
user_stuff.go
fix.go
```

### 2. **Version Control**

```go
// Always use timestamp-based versions
func (m *Migration) Version() string {
    return "2025_01_01_000001" // YYYY_MM_DD_HHMMSS format
}

// Include meaningful descriptions
func (m *Migration) Description() string {
    return "Create users table with email uniqueness constraint"
}
```

### 3. **Safe Schema Changes**

```go
// Good: Check before making changes
func (m *Migration) Up(db *gorm.DB) error {
    if !db.Migrator().HasTable("users") {
        return db.AutoMigrate(&User{})
    }
    return nil
}

// Good: Handle rollback safely
func (m *Migration) Down(db *gorm.DB) error {
    if db.Migrator().HasTable("users") {
        return db.Migrator().DropTable("users")
    }
    return nil
}
```

### 4. **Data Migration Safety**

```go
func (m *DataMigration) Up(db *gorm.DB) error {
    // Use transactions for data migrations
    return db.Transaction(func(tx *gorm.DB) error {
        // Batch process large datasets
        var users []User
        return tx.FindInBatches(&users, 1000, func(tx *gorm.DB, batch int) error {
            for _, user := range users {
                // Process each user
                if err := updateUserData(tx, &user); err != nil {
                    return err
                }
            }
            return nil
        }).Error
    })
}
```

### 5. **Cross-Database Compatibility**

```go
func (m *Migration) Up(db *gorm.DB) error {
    dialect := db.Dialector.Name()

    switch dialect {
    case "mysql":
        return db.Exec("ALTER TABLE users ADD COLUMN settings JSON").Error
    case "postgres":
        return db.Exec("ALTER TABLE users ADD COLUMN settings JSONB").Error
    case "sqlite":
        return db.Exec("ALTER TABLE users ADD COLUMN settings TEXT").Error
    default:
        return fmt.Errorf("unsupported database: %s", dialect)
    }
}
```

### 6. **Testing Migrations**

```go
func TestMigration(t *testing.T) {
    // Test with different databases
    databases := []string{"sqlite", "mysql", "postgres"}

    for _, dbType := range databases {
        t.Run(dbType, func(t *testing.T) {
            db := setupTestDatabase(dbType)
            defer cleanupTestDatabase(db)

            migration := &CreateUsersTable20250101000001{}

            // Test Up migration
            err := migration.Up(db)
            assert.NoError(t, err)
            assert.True(t, db.Migrator().HasTable("users"))

            // Test Down migration
            err = migration.Down(db)
            assert.NoError(t, err)
            assert.False(t, db.Migrator().HasTable("users"))
        })
    }
}
```

## 🚨 Common Pitfalls

### 1. **Avoid These Mistakes**

```go
// BAD: Don't use application models in migrations
func (m *Migration) Up(db *gorm.DB) error {
    // This will break if User model changes
    return db.AutoMigrate(&models.User{})
}

// GOOD: Define models in migration
func (m *Migration) Up(db *gorm.DB) error {
    type User struct {
        ID    uint   `gorm:"primaryKey"`
        Name  string `gorm:"size:255"`
        Email string `gorm:"size:255;uniqueIndex"`
    }
    return db.AutoMigrate(&User{})
}
```

### 2. **Handle Rollback Properly**

```go
// BAD: Irreversible migration
func (m *Migration) Down(db *gorm.DB) error {
    return errors.New("cannot rollback this migration")
}

// GOOD: Provide rollback logic
func (m *Migration) Down(db *gorm.DB) error {
    return db.Migrator().DropColumn("users", "new_column")
}
```

## 🔗 Related Packages

- [`pkg/database`](../database/) - Multi-database support
- [`pkg/seeder`](../seeder/) - Database seeding system
- [`internal/migrations`](../../internal/migrations/) - Application migrations
- [`cmd/artisan`](../../cmd/artisan/) - Migration CLI commands

## 📚 Additional Resources

- [GORM Migration Guide](https://gorm.io/docs/migration.html)
- [Database Migration Best Practices](https://www.prisma.io/dataguide/types/relational/what-are-database-migrations)
- [Laravel Migration Documentation](https://laravel.com/docs/migrations)
