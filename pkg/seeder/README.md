# 🌱 Seeder Package

Laravel-style database seeding system with dependency resolution, batch processing, and environment-aware seeding.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Seeder Interface](#seeder-interface)
- [Creating Seeders](#creating-seeders)
- [Running Seeders](#running-seeders)
- [Dependency System](#dependency-system)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/seeder"
```

## ⚡ Quick Start

### Creating a Seeder

```bash
# Using Artisan CLI
make make-seeder NAME=UserSeeder TABLE=users

# This creates: internal/seeders/user_seeder.go
```

### Basic Seeder Structure

```go
package seeders

import (
    "go-starter/internal/entity"
    "go-starter/pkg/seeder"
    "gorm.io/gorm"
    "time"
)

type UserSeeder struct{}

func (s *UserSeeder) Run(db *gorm.DB) error {
    // Check if users already exist
    var count int64
    db.Model(&entity.User{}).Count(&count)
    if count > 0 {
        return nil // Skip if data exists
    }

    // Create sample users
    users := []entity.User{
        {
            Name:      "John Doe",
            Email:     "john@example.com",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        {
            Name:      "Jane Smith",
            Email:     "jane@example.com",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }

    return db.Create(&users).Error
}

func (s *UserSeeder) Name() string {
    return "UserSeeder"
}

func (s *UserSeeder) Dependencies() []string {
    return []string{} // No dependencies
}

func init() {
    seeder.Register(&UserSeeder{})
}
```

## 🔧 Seeder Interface

### Seeder Interface

```go
type Seeder interface {
    Run(db *gorm.DB) error    // Execute seeding logic
    Name() string             // Seeder name
    Dependencies() []string   // List of required seeders
}
```

### Seeder Manager

```go
type Manager struct {
    db      *gorm.DB
    config  *SeederConfig
    seeders []Seeder
}

// Key methods
func (m *Manager) RunSeeders(seederName string) error
func (m *Manager) RunSpecificSeeder(seederName string) error
func (m *Manager) ListSeeders() error
```

### Configuration

```go
type SeederConfig struct {
    BatchSize int  // Batch size for bulk operations
    Verbose   bool // Enable verbose output
}

func DefaultSeederConfig() *SeederConfig {
    return &SeederConfig{
        BatchSize: 1000,
        Verbose:   true,
    }
}
```

## 📝 Creating Seeders

### 1. **Using Artisan CLI (Recommended)**

```bash
# Basic seeder
make make-seeder NAME=ProductSeeder TABLE=products

# Seeder with dependencies
make make-seeder NAME=OrderSeeder TABLE=orders DEPS="UserSeeder,ProductSeeder"
```

### 2. **Manual Creation**

```go
package seeders

import (
    "go-starter/internal/entity"
    "go-starter/pkg/seeder"
    "gorm.io/gorm"
)

type ProductSeeder struct{}

func (s *ProductSeeder) Run(db *gorm.DB) error {
    // Check if products already exist
    var count int64
    db.Model(&entity.Product{}).Count(&count)
    if count > 0 {
        return nil
    }

    // Create sample products
    products := []entity.Product{
        {
            Name:        "Laptop",
            Description: "High-performance laptop",
            Price:       999.99,
            Stock:       50,
        },
        {
            Name:        "Mouse",
            Description: "Wireless mouse",
            Price:       29.99,
            Stock:       100,
        },
    }

    // Batch insert for better performance
    return db.CreateInBatches(products, 100).Error
}

func (s *ProductSeeder) Name() string {
    return "ProductSeeder"
}

func (s *ProductSeeder) Dependencies() []string {
    return []string{"CategorySeeder"} // Depends on categories
}

func init() {
    seeder.Register(&ProductSeeder{})
}
```

## 🚀 Running Seeders

### Command Line

```bash
# Run all seeders
make db-seed

# Run specific seeder
make db-seed-specific NAME=UserSeeder

# List available seeders
make db-seed-list
```

### Programmatic Usage

```go
// Create seeder manager
config := seeder.DefaultSeederConfig()
manager := seeder.NewManagerWithGlobalSeeders(db, config)

// Run all seeders
if err := manager.RunSeeders(""); err != nil {
    log.Fatal("Seeding failed:", err)
}

// Run specific seeder
if err := manager.RunSpecificSeeder("UserSeeder"); err != nil {
    log.Error("UserSeeder failed:", err)
}

// List seeders
if err := manager.ListSeeders(); err != nil {
    log.Error("Failed to list seeders:", err)
}
```

## 🔗 Dependency System

### Dependency Resolution

The seeder system automatically resolves dependencies and runs seeders in the correct order:

```go
type CategorySeeder struct{}
func (s *CategorySeeder) Dependencies() []string { return []string{} }

type ProductSeeder struct{}
func (s *ProductSeeder) Dependencies() []string { return []string{"CategorySeeder"} }

type OrderSeeder struct{}
func (s *OrderSeeder) Dependencies() []string { return []string{"UserSeeder", "ProductSeeder"} }

// Execution order: CategorySeeder → UserSeeder → ProductSeeder → OrderSeeder
```

### Complex Dependencies

```go
type UserRoleSeeder struct{}
func (s *UserRoleSeeder) Dependencies() []string { return []string{} }

type UserSeeder struct{}
func (s *UserSeeder) Dependencies() []string { return []string{"UserRoleSeeder"} }

type CategorySeeder struct{}
func (s *CategorySeeder) Dependencies() []string { return []string{} }

type ProductSeeder struct{}
func (s *ProductSeeder) Dependencies() []string { return []string{"CategorySeeder"} }

type OrderSeeder struct{}
func (s *OrderSeeder) Dependencies() []string { return []string{"UserSeeder", "ProductSeeder"} }

type ReviewSeeder struct{}
func (s *ReviewSeeder) Dependencies() []string { return []string{"OrderSeeder"} }
```

## 💡 Examples

### 1. **User Seeder with Roles**

```go
type UserSeeder struct{}

func (s *UserSeeder) Run(db *gorm.DB) error {
    var count int64
    db.Model(&entity.User{}).Count(&count)
    if count > 0 {
        return nil
    }

    // Get roles for assignment
    var adminRole, userRole entity.Role
    db.Where("name = ?", "admin").First(&adminRole)
    db.Where("name = ?", "user").First(&userRole)

    users := []entity.User{
        {
            Name:     "Admin User",
            Email:    "admin@example.com",
            RoleID:   adminRole.ID,
            Active:   true,
        },
        {
            Name:     "Regular User",
            Email:    "user@example.com",
            RoleID:   userRole.ID,
            Active:   true,
        },
    }

    return db.Create(&users).Error
}

func (s *UserSeeder) Dependencies() []string {
    return []string{"RoleSeeder"}
}
```

### 2. **Large Dataset Seeder**

```go
type LargeProductSeeder struct{}

func (s *LargeProductSeeder) Run(db *gorm.DB) error {
    var count int64
    db.Model(&entity.Product{}).Count(&count)
    if count > 1000 {
        return nil // Already seeded
    }

    // Generate large dataset
    const totalProducts = 10000
    const batchSize = 1000

    for i := 0; i < totalProducts; i += batchSize {
        var products []entity.Product

        for j := 0; j < batchSize && (i+j) < totalProducts; j++ {
            products = append(products, entity.Product{
                Name:        fmt.Sprintf("Product %d", i+j+1),
                Description: fmt.Sprintf("Description for product %d", i+j+1),
                Price:       float64(rand.Intn(1000) + 10),
                Stock:       rand.Intn(100) + 1,
                SKU:         fmt.Sprintf("SKU%06d", i+j+1),
            })
        }

        if err := db.CreateInBatches(products, batchSize).Error; err != nil {
            return fmt.Errorf("failed to create batch %d-%d: %w", i, i+batchSize, err)
        }

        // Progress logging
        fmt.Printf("Created products %d-%d/%d\n", i+1, i+len(products), totalProducts)
    }

    return nil
}
```

### 3. **Environment-Specific Seeder**

```go
type EnvironmentSeeder struct{}

func (s *EnvironmentSeeder) Run(db *gorm.DB) error {
    env := os.Getenv("ENV")

    switch env {
    case "development":
        return s.seedDevelopmentData(db)
    case "staging":
        return s.seedStagingData(db)
    case "production":
        return s.seedProductionData(db)
    default:
        return s.seedDevelopmentData(db)
    }
}

func (s *EnvironmentSeeder) seedDevelopmentData(db *gorm.DB) error {
    // Create test data with fake information
    users := []entity.User{
        {Name: "Test User 1", Email: "test1@example.com"},
        {Name: "Test User 2", Email: "test2@example.com"},
    }
    return db.Create(&users).Error
}

func (s *EnvironmentSeeder) seedStagingData(db *gorm.DB) error {
    // Create realistic but safe test data
    users := []entity.User{
        {Name: "Staging Admin", Email: "admin@staging.com"},
        {Name: "Staging User", Email: "user@staging.com"},
    }
    return db.Create(&users).Error
}

func (s *EnvironmentSeeder) seedProductionData(db *gorm.DB) error {
    // Only create essential data for production
    var adminCount int64
    db.Model(&entity.User{}).Where("email = ?", "admin@company.com").Count(&adminCount)

    if adminCount == 0 {
        admin := entity.User{
            Name:  "System Admin",
            Email: "admin@company.com",
            Role:  "admin",
        }
        return db.Create(&admin).Error
    }

    return nil
}
```

### 4. **File-Based Seeder**

```go
type FileBasedSeeder struct{}

func (s *FileBasedSeeder) Run(db *gorm.DB) error {
    // Read data from JSON file
    data, err := ioutil.ReadFile("seeders/data/countries.json")
    if err != nil {
        return fmt.Errorf("failed to read countries data: %w", err)
    }

    var countries []entity.Country
    if err := json.Unmarshal(data, &countries); err != nil {
        return fmt.Errorf("failed to parse countries data: %w", err)
    }

    // Check if data already exists
    var count int64
    db.Model(&entity.Country{}).Count(&count)
    if count > 0 {
        return nil
    }

    // Insert countries in batches
    return db.CreateInBatches(countries, 100).Error
}
```

### 5. **Relationship Seeder**

```go
type OrderSeeder struct{}

func (s *OrderSeeder) Run(db *gorm.DB) error {
    var count int64
    db.Model(&entity.Order{}).Count(&count)
    if count > 0 {
        return nil
    }

    // Get existing users and products
    var users []entity.User
    var products []entity.Product

    db.Find(&users)
    db.Find(&products)

    if len(users) == 0 || len(products) == 0 {
        return fmt.Errorf("users and products must exist before creating orders")
    }

    // Create orders with relationships
    for _, user := range users {
        for i := 0; i < 3; i++ { // 3 orders per user
            order := entity.Order{
                UserID:      user.ID,
                Status:      "completed",
                TotalAmount: 0,
                OrderDate:   time.Now().AddDate(0, -rand.Intn(6), -rand.Intn(30)),
            }

            if err := db.Create(&order).Error; err != nil {
                return err
            }

            // Add random products to order
            numProducts := rand.Intn(3) + 1 // 1-3 products per order
            for j := 0; j < numProducts; j++ {
                product := products[rand.Intn(len(products))]
                quantity := rand.Intn(3) + 1

                orderItem := entity.OrderItem{
                    OrderID:   order.ID,
                    ProductID: product.ID,
                    Quantity:  quantity,
                    Price:     product.Price,
                }

                if err := db.Create(&orderItem).Error; err != nil {
                    return err
                }

                order.TotalAmount += product.Price * float64(quantity)
            }

            // Update order total
            db.Save(&order)
        }
    }

    return nil
}

func (s *OrderSeeder) Dependencies() []string {
    return []string{"UserSeeder", "ProductSeeder"}
}
```

## 🎯 Best Practices

### 1. **Idempotent Seeders**

```go
// Always check if data exists before seeding
func (s *UserSeeder) Run(db *gorm.DB) error {
    var count int64
    db.Model(&entity.User{}).Count(&count)
    if count > 0 {
        return nil // Skip if data exists
    }

    // Seed data...
}
```

### 2. **Use Transactions for Complex Seeders**

```go
func (s *ComplexSeeder) Run(db *gorm.DB) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // All seeding operations in transaction
        if err := s.createUsers(tx); err != nil {
            return err
        }

        if err := s.createProducts(tx); err != nil {
            return err
        }

        return s.createOrders(tx)
    })
}
```

### 3. **Batch Processing for Large Datasets**

```go
func (s *LargeSeeder) Run(db *gorm.DB) error {
    const batchSize = 1000

    for i := 0; i < totalRecords; i += batchSize {
        batch := generateBatch(i, batchSize)
        if err := db.CreateInBatches(batch, batchSize).Error; err != nil {
            return err
        }
    }

    return nil
}
```

### 4. **Environment-Aware Seeding**

```go
func (s *SeederWithEnv) Run(db *gorm.DB) error {
    env := os.Getenv("ENV")

    if env == "production" {
        // Only essential data in production
        return s.seedEssentialData(db)
    }

    // Full test data for development/staging
    return s.seedTestData(db)
}
```

### 5. **Error Handling and Logging**

```go
func (s *SeederWithLogging) Run(db *gorm.DB) error {
    log.Printf("Starting %s...", s.Name())

    startTime := time.Now()
    defer func() {
        log.Printf("%s completed in %v", s.Name(), time.Since(startTime))
    }()

    if err := s.performSeeding(db); err != nil {
        log.Printf("Error in %s: %v", s.Name(), err)
        return fmt.Errorf("%s failed: %w", s.Name(), err)
    }

    return nil
}
```

### 6. **Testing Seeders**

```go
func TestUserSeeder(t *testing.T) {
    db := setupTestDatabase()
    defer cleanupTestDatabase(db)

    seeder := &UserSeeder{}

    // Test seeding
    err := seeder.Run(db)
    assert.NoError(t, err)

    // Verify data was created
    var count int64
    db.Model(&entity.User{}).Count(&count)
    assert.Greater(t, count, int64(0))

    // Test idempotency
    err = seeder.Run(db)
    assert.NoError(t, err)

    // Count should remain the same
    var newCount int64
    db.Model(&entity.User{}).Count(&newCount)
    assert.Equal(t, count, newCount)
}
```

## 🚨 Common Pitfalls

### 1. **Avoid These Mistakes**

```go
// BAD: Not checking if data exists
func (s *BadSeeder) Run(db *gorm.DB) error {
    users := []entity.User{...}
    return db.Create(&users).Error // Will fail on second run
}

// GOOD: Check before creating
func (s *GoodSeeder) Run(db *gorm.DB) error {
    var count int64
    db.Model(&entity.User{}).Count(&count)
    if count > 0 {
        return nil
    }

    users := []entity.User{...}
    return db.Create(&users).Error
}
```

### 2. **Handle Dependencies Properly**

```go
// BAD: Assuming dependencies exist
func (s *BadSeeder) Run(db *gorm.DB) error {
    var role entity.Role
    db.First(&role) // Might not exist!

    user := entity.User{RoleID: role.ID}
    return db.Create(&user).Error
}

// GOOD: Verify dependencies
func (s *GoodSeeder) Run(db *gorm.DB) error {
    var role entity.Role
    if err := db.First(&role).Error; err != nil {
        return fmt.Errorf("required role not found: %w", err)
    }

    user := entity.User{RoleID: role.ID}
    return db.Create(&user).Error
}
```

## 🔗 Related Packages

- [`pkg/database`](../database/) - Multi-database support
- [`pkg/migration`](../migration/) - Database migration system
- [`internal/seeders`](../../internal/seeders/) - Application seeders
- [`cmd/artisan`](../../cmd/artisan/) - Seeder CLI commands

## 📚 Additional Resources

- [GORM Create Documentation](https://gorm.io/docs/create.html)
- [Database Seeding Best Practices](https://laravel.com/docs/seeding)
- [Test Data Generation Strategies](https://martinfowler.com/articles/mocksArentStubs.html)
