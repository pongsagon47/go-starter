# 🧪 Multi-Database Testing Strategy

This example demonstrates comprehensive testing strategies for multi-database applications.

## 🎯 Testing Pyramid

```
┌─────────────────────────────────────────────┐
│                E2E Tests                    │ ← PostgreSQL (Production-like)
│              (Slow, Expensive)              │
├─────────────────────────────────────────────┤
│           Integration Tests                 │ ← MySQL (Compatibility)
│         (Medium Speed, Medium Cost)         │
├─────────────────────────────────────────────┤
│              Unit Tests                     │ ← SQLite (Fast, Isolated)
│           (Fast, Inexpensive)               │
└─────────────────────────────────────────────┘
```

## 🚀 Testing Strategies by Database

### **SQLite - Unit Tests**

**Use Case:** Fast, isolated unit tests

```bash
# Configuration for unit tests
export DB_TYPE=sqlite
export DB_SQLITE_IN_MEMORY=true
export DB_SQLITE_FILE_PATH=":memory:"

# Run unit tests
make test
```

**Benefits:**

- ⚡ **Ultra Fast** - In-memory database
- 🔒 **Isolated** - Each test gets fresh database
- 📦 **No Dependencies** - No external database required
- 💰 **Cost Effective** - No infrastructure costs

### **MySQL - Integration Tests**

**Use Case:** Cross-database compatibility testing

```bash
# Configuration for integration tests
export DB_TYPE=mysql
export DB_MYSQL_HOST=localhost
export DB_MYSQL_NAME=test_integration
export DB_MYSQL_USER=test_user
export DB_MYSQL_PASSWORD=test_pass

# Run integration tests
make test-integration
```

**Benefits:**

- 🔍 **Compatibility Testing** - Catch MySQL-specific issues
- 📊 **Realistic Data Types** - Test with actual MySQL types
- 🔄 **Transaction Testing** - Test complex transactions
- 🌐 **Network Testing** - Test with network database

### **PostgreSQL - E2E Tests**

**Use Case:** Production-like end-to-end testing

```bash
# Configuration for E2E tests
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=test-postgres
export DB_POSTGRES_NAME=test_e2e
export DB_POSTGRES_USER=test_user
export DB_POSTGRES_PASSWORD=test_pass

# Run E2E tests
make test-e2e
```

**Benefits:**

- 🏭 **Production-like** - Same database as production
- ⚡ **Performance Testing** - Real performance characteristics
- 🔧 **Feature Testing** - Test PostgreSQL-specific features
- 📈 **Scale Testing** - Test with large datasets

## 📋 Test Implementation Examples

### **1. Unit Test with SQLite**

```go
// user_test.go
func TestUserRepository_Create(t *testing.T) {
    // Setup in-memory SQLite
    db := setupTestDB(t)
    defer db.Close()

    repo := NewUserRepository(db)

    // Test data
    user := &User{
        Name:  "John Doe",
        Email: "john@example.com",
    }

    // Execute
    err := repo.Create(user)

    // Assert
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)

    // Verify in database
    found, err := repo.GetByID(user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
}

func setupTestDB(t *testing.T) *gorm.DB {
    // Use in-memory SQLite for tests
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    err = db.AutoMigrate(&User{})
    require.NoError(t, err)

    return db
}
```

### **2. Integration Test with MySQL**

```go
// integration_test.go
func TestUserService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup MySQL test database
    db := setupMySQLTestDB(t)
    defer cleanupTestDB(t, db)

    // Setup services
    userRepo := NewUserRepository(db)
    userService := NewUserService(userRepo)

    // Test complex scenario
    t.Run("CreateUserWithProfile", func(t *testing.T) {
        user := &User{
            Name:  "Jane Doe",
            Email: "jane@example.com",
            Profile: &Profile{
                Bio:     "Software Developer",
                Website: "https://jane.dev",
            },
        }

        err := userService.CreateWithProfile(user)
        assert.NoError(t, err)

        // Verify relationships work correctly
        found, err := userService.GetWithProfile(user.ID)
        assert.NoError(t, err)
        assert.Equal(t, user.Profile.Bio, found.Profile.Bio)
    })
}

func setupMySQLTestDB(t *testing.T) *gorm.DB {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        os.Getenv("DB_MYSQL_USER"),
        os.Getenv("DB_MYSQL_PASSWORD"),
        os.Getenv("DB_MYSQL_HOST"),
        os.Getenv("DB_MYSQL_PORT"),
        os.Getenv("DB_MYSQL_NAME"))

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    err = db.AutoMigrate(&User{}, &Profile{})
    require.NoError(t, err)

    return db
}
```

### **3. E2E Test with PostgreSQL**

```go
// e2e_test.go
func TestAPI_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test")
    }

    // Setup PostgreSQL test database
    db := setupPostgreSQLTestDB(t)
    defer cleanupTestDB(t, db)

    // Setup test server
    server := setupTestServer(t, db)
    defer server.Close()

    client := &http.Client{Timeout: 10 * time.Second}

    t.Run("CompleteUserFlow", func(t *testing.T) {
        // Create user
        user := createUserViaAPI(t, client, server.URL)
        assert.NotEmpty(t, user.ID)

        // Get user
        foundUser := getUserViaAPI(t, client, server.URL, user.ID)
        assert.Equal(t, user.Name, foundUser.Name)

        // Update user
        updatedUser := updateUserViaAPI(t, client, server.URL, user.ID)
        assert.NotEqual(t, user.Name, updatedUser.Name)

        // Delete user
        deleteUserViaAPI(t, client, server.URL, user.ID)

        // Verify deletion
        _, err := getUserViaAPI(t, client, server.URL, user.ID)
        assert.Error(t, err)
    })
}

func setupPostgreSQLTestDB(t *testing.T) *gorm.DB {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        os.Getenv("DB_POSTGRES_HOST"),
        os.Getenv("DB_POSTGRES_USER"),
        os.Getenv("DB_POSTGRES_PASSWORD"),
        os.Getenv("DB_POSTGRES_NAME"),
        os.Getenv("DB_POSTGRES_PORT"))

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    err = db.AutoMigrate(&User{}, &Profile{})
    require.NoError(t, err)

    return db
}
```

## 🔧 Test Configuration

### **Makefile Test Targets**

```makefile
# Add to Makefile

## Run unit tests with SQLite
test:
	@echo "🧪 Running unit tests with SQLite..."
	@DB_TYPE=sqlite DB_SQLITE_IN_MEMORY=true go test -v -short ./...

## Run integration tests with MySQL
test-integration:
	@echo "🔧 Running integration tests with MySQL..."
	@DB_TYPE=mysql go test -v -tags=integration ./...

## Run E2E tests with PostgreSQL
test-e2e:
	@echo "🌐 Running E2E tests with PostgreSQL..."
	@DB_TYPE=postgresql go test -v -tags=e2e ./...

## Run all tests
test-all:
	@echo "🚀 Running all test suites..."
	@$(MAKE) test
	@$(MAKE) test-integration
	@$(MAKE) test-e2e

## Run performance tests
test-performance:
	@echo "📊 Running performance tests..."
	@DB_TYPE=postgresql go test -v -tags=performance -bench=. ./...

## Run tests with coverage
test-coverage:
	@echo "📈 Running tests with coverage..."
	@DB_TYPE=sqlite DB_SQLITE_IN_MEMORY=true go test -v -short -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
```

### **Test Environment Files**

**test.env**

```env
# Unit test configuration
DB_TYPE=sqlite
DB_SQLITE_IN_MEMORY=true
LOG_LEVEL=error
```

**integration.env**

```env
# Integration test configuration
DB_TYPE=mysql
DB_MYSQL_HOST=localhost
DB_MYSQL_PORT=3306
DB_MYSQL_USER=test_user
DB_MYSQL_PASSWORD=test_pass
DB_MYSQL_NAME=test_integration
LOG_LEVEL=warn
```

**e2e.env**

```env
# E2E test configuration
DB_TYPE=postgresql
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=test_user
DB_POSTGRES_PASSWORD=test_pass
DB_POSTGRES_NAME=test_e2e
LOG_LEVEL=info
```

## 🐳 Docker Test Environment

### **docker-compose.test.yml**

```yaml
version: "3.8"

services:
  # MySQL for integration tests
  mysql-test:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root_pass
      MYSQL_DATABASE: test_integration
      MYSQL_USER: test_user
      MYSQL_PASSWORD: test_pass
    ports:
      - "3307:3306"
    tmpfs:
      - /var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password

  # PostgreSQL for E2E tests
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: test_e2e
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_pass
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data

  # Test runner
  test-runner:
    build:
      context: .
      dockerfile: Dockerfile.test
    depends_on:
      - mysql-test
      - postgres-test
    environment:
      # Integration test env
      - DB_MYSQL_HOST=mysql-test
      - DB_MYSQL_PORT=3306
      - DB_MYSQL_USER=test_user
      - DB_MYSQL_PASSWORD=test_pass
      - DB_MYSQL_NAME=test_integration
      # E2E test env
      - DB_POSTGRES_HOST=postgres-test
      - DB_POSTGRES_PORT=5432
      - DB_POSTGRES_USER=test_user
      - DB_POSTGRES_PASSWORD=test_pass
      - DB_POSTGRES_NAME=test_e2e
    volumes:
      - .:/app
    working_dir: /app
    command: make test-all
```

### **Dockerfile.test**

```dockerfile
FROM golang:1.23-alpine

RUN apk add --no-cache make curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["make", "test-all"]
```

## 🚀 CI/CD Test Pipeline

### **GitHub Actions Workflow**

```yaml
name: Multi-Database Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: "1.23"

      - name: Run unit tests
        run: make test

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root_pass
          MYSQL_DATABASE: test_integration
          MYSQL_USER: test_user
          MYSQL_PASSWORD: test_pass
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: "1.23"

      - name: Wait for MySQL
        run: |
          until mysqladmin ping -h127.0.0.1 -P3306 -utest_user -ptest_pass --silent; do
            echo 'waiting for mysql'
            sleep 1
          done

      - name: Run integration tests
        env:
          DB_MYSQL_HOST: localhost
          DB_MYSQL_PORT: 3306
          DB_MYSQL_USER: test_user
          DB_MYSQL_PASSWORD: test_pass
          DB_MYSQL_NAME: test_integration
        run: make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: test_e2e
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_pass
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: "1.23"

      - name: Run E2E tests
        env:
          DB_POSTGRES_HOST: localhost
          DB_POSTGRES_PORT: 5432
          DB_POSTGRES_USER: test_user
          DB_POSTGRES_PASSWORD: test_pass
          DB_POSTGRES_NAME: test_e2e
        run: make test-e2e

  performance-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: test_performance
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_pass
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: "1.23"

      - name: Run performance tests
        env:
          DB_POSTGRES_HOST: localhost
          DB_POSTGRES_PORT: 5432
          DB_POSTGRES_USER: test_user
          DB_POSTGRES_PASSWORD: test_pass
          DB_POSTGRES_NAME: test_performance
        run: make test-performance
```

## 📊 Test Metrics & Reporting

### **Test Coverage Analysis**

```bash
# Generate coverage report
make test-coverage

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### **Performance Benchmarks**

```go
// benchmark_test.go
func BenchmarkUserRepository_Create(b *testing.B) {
    databases := []struct {
        name string
        db   *gorm.DB
    }{
        {"SQLite", setupSQLiteDB(b)},
        {"MySQL", setupMySQLDB(b)},
        {"PostgreSQL", setupPostgreSQLDB(b)},
    }

    for _, database := range databases {
        b.Run(database.name, func(b *testing.B) {
            repo := NewUserRepository(database.db)

            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                user := &User{
                    Name:  fmt.Sprintf("User%d", i),
                    Email: fmt.Sprintf("user%d@example.com", i),
                }
                repo.Create(user)
            }
        })
    }
}
```

## 🎯 Testing Best Practices

### **1. Test Isolation**

- Use transactions that rollback
- Use separate test databases
- Clean up after each test

### **2. Test Data Management**

```go
// Use test fixtures
func setupTestData(db *gorm.DB) {
    users := []User{
        {Name: "Alice", Email: "alice@test.com"},
        {Name: "Bob", Email: "bob@test.com"},
    }
    db.Create(&users)
}

// Use factories for complex data
func CreateTestUser(overrides ...func(*User)) *User {
    user := &User{
        Name:  "Test User",
        Email: "test@example.com",
    }

    for _, override := range overrides {
        override(user)
    }

    return user
}
```

### **3. Database-Specific Tests**

```go
// Test database-specific features
func TestPostgreSQLJSONB(t *testing.T) {
    if os.Getenv("DB_TYPE") != "postgresql" {
        t.Skip("PostgreSQL-specific test")
    }

    // Test JSONB functionality
}

func TestMySQLFullText(t *testing.T) {
    if os.Getenv("DB_TYPE") != "mysql" {
        t.Skip("MySQL-specific test")
    }

    // Test full-text search
}
```

## 📈 Test Execution Strategy

### **Local Development**

```bash
# Quick feedback loop
make test                    # ~5 seconds
make test-integration       # ~30 seconds
make test-e2e              # ~2 minutes
```

### **CI/CD Pipeline**

```bash
# Parallel execution
make test &                 # SQLite unit tests
make test-integration &     # MySQL integration tests
make test-e2e &            # PostgreSQL E2E tests
wait                       # Wait for all to complete
```

### **Production Deployment**

```bash
# Full test suite before deployment
make test-all              # All databases, all test types
make test-performance      # Performance regression tests
```

## 🔗 Related Examples

- [E-commerce Example](./ecommerce-example.md)
- [Production Deployment](./production-example.md)
- [Blog System Example](./blog-example.md)
