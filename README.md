# 🚀 Go Starter - Multi-Database Laravel-style Development

A production-ready Go starter template with **multi-database support** (MySQL, PostgreSQL, SQLite), **Laravel-style CLI**, and **Clean Architecture** patterns.

[![Go Version](https://img.shields.io/badge/Go-1.23.4+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Database](https://img.shields.io/badge/Database-MySQL%20%7C%20PostgreSQL%20%7C%20SQLite-orange.svg)](#database-support)

## ✨ Features

### 🗄️ **Multi-Database Support**

- **MySQL** - Full SSL/TLS, connection pooling
- **PostgreSQL** - SSL modes, timeouts, timezone support
- **SQLite** - File/memory, foreign keys, WAL journal
- **Same Commands** - Switch databases without changing code

### 🎨 **Laravel-style CLI (Artisan)**

- **Dynamic Migration Discovery** - Auto-register migrations without rebuilding
- **Multi-Database Templates** - SQLite/MySQL/PostgreSQL specific migration generation
- **Primary Key Strategies** - Support for int, uuid, and dual (int+uuid) strategies
- **Seeders** - Database seeding with dependency resolution
- **Code Generation** - Auto-generate models, packages, migrations
- **Database Management** - Create, drop, reset databases

### 🏗️ **Clean Architecture**

- **Separation of Concerns** - Entity, Repository, Usecase, Handler layers
- **Modern Dependency Injection** - Factory Pattern with Service Registry
- **Interface-based Design** - Easy testing and mocking with ContainerInterface
- **Graceful Error Handling** - Non-fatal error recovery and resource cleanup
- **Domain-driven Design** - Business logic in the center

### ⚡ **Production Ready**

- **Hot Reload** - Air/CompileDaemon integration
- **Health Checks** - Built-in dependency health monitoring
- **Error Helper Functions** - Convenient error creation and handling
- **Logging** - Structured logging with Zap
- **Security** - Helmet, CORS, input validation
- **Email** - SMTP integration with templates
- **JWT Authentication** - Complete authentication system with refresh tokens

---

## 🚀 Quick Start

### 1. **Setup Project**

```bash
# Clone and setup
git clone https://github.com/pongsagon47/flex-service.git flex-service
cd flex-service

# Install dependencies and setup
make setup

# Copy environment configuration
cp env.example .env
# Edit .env with your database settings
```

### 2. **Choose Your Database**

```bash
# SQLite (default, no setup required)
export DB_DRIVER=sqlite

# MySQL
export DB_DRIVER=mysql
# Configure DB_MYSQL_* variables in .env

# PostgreSQL
export DB_DRIVER=postgresql
# Configure DB_POSTGRES_* variables in .env
```

### 3. **Run Migrations & Seeders**

```bash
# Run migrations
make migrate

# Seed database
make db-seed

# Start development server
make dev
```

### 4. **Access Your API**

```bash
# Health check
curl http://localhost:8080/health

# Database health
curl http://localhost:8080/health/db

# Demo endpoint with authentication info
curl http://localhost:8080/api/v1/demo

# Test authentication (register a user)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"MySecure123!","first_name":"Test","last_name":"User"}'

# Login and get access token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_username":"test@example.com","password":"MySecure123!"}'
```

---

## 📚 Database Support

### 🔄 **Switch Between Databases**

```bash
# Use SQLite for development
make db-sqlite migrate
make db-sqlite db-seed
make db-sqlite run

# Use MySQL for staging
make db-mysql migrate
make db-mysql db-seed
make db-mysql run

# Use PostgreSQL for production
make db-postgres migrate
make db-postgres db-seed
make db-postgres run
```

### ⚙️ **Database Configuration**

**SQLite** (Development)

```env
DB_DRIVER=sqlite
DB_SQLITE_FILE_PATH=./database.db
DB_SQLITE_FOREIGN_KEYS=true
DB_SQLITE_JOURNAL=WAL
```

**MySQL** (Staging/Production)

```env
DB_DRIVER=mysql
DB_MYSQL_HOST=localhost
DB_MYSQL_PORT=3306
DB_MYSQL_USER=root
DB_MYSQL_PASSWORD=password
DB_MYSQL_NAME=go_starter
DB_MYSQL_MAX_OPEN_CONNS=100
DB_MYSQL_MAX_IDLE_CONNS=10
```

**PostgreSQL** (Production)

```env
DB_DRIVER=postgresql
DB_POSTGRES_HOST=localhost
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=postgres
DB_POSTGRES_PASSWORD=password
DB_POSTGRES_NAME=go_starter
DB_POSTGRES_SSL_MODE=require
DB_POSTGRES_TIMEZONE=UTC
```

---

## 🎨 Laravel-style Development

### **Create Complete Model Stack**

```bash
# Create User model with migration and seeder
make make-model NAME=User TABLE=users FIELDS="name:string,email:string,age:int"

# This creates:
# ✅ internal/entity/user.go (Entity struct)
# ✅ internal/migrations/TIMESTAMP_create_users_table.go (Migration)
# ✅ internal/seeders/user_seeder.go (Seeder)
```

### **Create Package Structure**

```bash
# Create complete package (handler, usecase, repository, port)
make make-package NAME=Product

# This creates:
# ✅ internal/product/handler.go
# ✅ internal/product/usecase.go
# ✅ internal/product/repository.go
# ✅ internal/product/port.go
```

### **Database Migrations**

```bash
# Create migration
make make-migration NAME=create_posts_table CREATE=true TABLE=posts FIELDS="title:string,content:text"

# Run migrations
make migrate

# Check status
make migrate-status

# Rollback
make migrate-rollback COUNT=1

# Fresh migration (⚠️ DANGER - drops all data)
make migrate-fresh
```

### **Database Seeding with Dependencies**

```bash
# Create seeders with dependencies
make make-seeder NAME=UserSeeder TABLE=users
make make-seeder NAME=CategorySeeder TABLE=categories
make make-seeder NAME=ProductSeeder TABLE=products DEPS="UserSeeder,CategorySeeder"

# Run seeders (auto-resolves dependencies)
make db-seed

# List seeders with dependencies
make db-seed-list

# Run specific seeder (+ its dependencies)
make db-seed-specific NAME=ProductSeeder
```

---

## 🏗️ Project Structure

```
flex-service/
├── cmd/
│   ├── main.go                 # Application entry point
│   └── artisan/main.go         # Laravel-style CLI tool
├── config/
│   └── config.go               # Multi-database configuration
├── internal/
│   ├── entity/                 # Domain entities (with primary key strategies)
│   ├── auth/                   # Authentication system (repository, usecase, handler)
│   ├── migrations/             # Database migrations (auto-discovery)
│   ├── seeders/               # Database seeders
│   ├── container/             # Modern dependency injection (Factory + Registry)
│   ├── router/                # HTTP routes
│   └── middleware/            # HTTP middleware
├── pkg/
│   ├── database/              # 🆕 Multi-database drivers
│   │   ├── interface.go       # Database interfaces
│   │   ├── factory.go         # Database factory
│   │   ├── mysql.go           # MySQL implementation
│   │   ├── postgresql.go      # PostgreSQL implementation
│   │   └── sqlite.go          # SQLite implementation
│   ├── migration/             # 🆕 Reusable migration engine
│   ├── seeder/               # 🆕 Reusable seeder engine
│   ├── auth/                  # 🆕 JWT authentication & authorization
│   ├── errors/               # 🆕 Error helper functions
│   ├── logger/               # Structured logging
│   ├── mail/                 # Email functionality
│   ├── secure/               # Security utilities
│   └── validator/            # Input validation
├── env.example               # Environment configuration
├── Makefile                 # Laravel-style commands
└── README.md               # This file
```

---

## 🛠️ Development Commands

### **Setup & Development**

```bash
make setup              # Setup project (first time)
make dev                # Run with hot reload
make dev-force          # Kill port conflicts and run
make run                # Run without hot reload
make build              # Build application
make test               # Run tests
make test-coverage      # Run tests with coverage
```

### **Laravel-style Generators**

```bash
# Basic generators
make make-migration NAME=create_posts_table CREATE=true TABLE=posts
make make-seeder NAME=PostSeeder
make make-entity NAME=Post
make make-package NAME=Post

# Advanced model generation with strategies
make make-model NAME=User TABLE=users STRATEGY=dual     # int ID (primary) + UUID (public)
make make-model NAME=Product TABLE=products STRATEGY=uuid  # UUID primary key only
make make-model NAME=Category TABLE=categories           # int ID primary key (default)

# With custom fields
make make-model NAME=Post TABLE=posts FIELDS="title:string,content:text,user_id:uuid|fk:users"
```

### **🔑 Primary Key Strategies**

Choose the right primary key strategy for your use case:

| Strategy   | Primary Key      | Public ID        | Use Case                           | Example              |
| ---------- | ---------------- | ---------------- | ---------------------------------- | -------------------- |
| **`int`**  | `ID int`         | `ID int`         | Internal systems, simple APIs      | Categories, Settings |
| **`uuid`** | `UUID uuid.UUID` | `UUID uuid.UUID` | Distributed systems, external APIs | Products, Orders     |
| **`dual`** | `ID int`         | `UUID uuid.UUID` | Best of both worlds                | Users, Posts         |

#### **Strategy Details:**

**🔢 `int` Strategy (Default)**

```go
type Category struct {
    ID        int       `gorm:"primaryKey"`           // Primary key for DB relations
    Name      string    `gorm:"not null"`
    // ... other fields
}
```

**🆔 `uuid` Strategy**

```go
type Product struct {
    UUID      uuid.UUID `gorm:"type:varchar(36);primaryKey;not null"` // Primary key
    Name      string    `gorm:"not null"`
    // ... other fields
}
```

**⚡ `dual` Strategy (Recommended)**

```go
type User struct {
    ID        int       `gorm:"primaryKey"`           // Primary key for DB relations
    UUID      uuid.UUID `json:"id" gorm:"type:varchar(36);not null"` // Public ID in JSON
    Email     string    `gorm:"unique;not null"`
    // ... other fields
}
```

#### **When to Use Each Strategy:**

- **`int`**: Simple internal systems, admin panels, configuration tables
- **`uuid`**: Microservices, public APIs, distributed systems, external integrations
- **`dual`**: User-facing applications where you need both performance (int) and security (UUID)

### **Database Management**

```bash
make migrate            # Run pending migrations
make migrate-status     # Show migration status
make migrate-rollback   # Rollback migrations
make db-seed            # Run all seeders (auto-resolves dependencies)
make db-seed-list       # List all seeders with their dependencies
make db-info            # Show database information
```

### **Multi-Database Commands**

```bash
make db-mysql migrate   # Run migrations on MySQL
make db-postgres migrate # Run migrations on PostgreSQL
make db-sqlite migrate  # Run migrations on SQLite
make test-all-db        # Test all database types
```

### **Quick Actions**

```bash
make add-column TABLE=users COLUMN=phone TYPE=string
make drop-column TABLE=users COLUMN=old_field
make add-index TABLE=products COLUMNS="category,price"
```

---

## 🌟 Advanced Features

### **🔗 Seeder Dependencies**

Seeders automatically run in the correct order based on dependencies:

```go
// UserSeeder (no dependencies)
func (s *UserSeeder) Dependencies() []string {
    return []string{}
}

// ProductSeeder (depends on UserSeeder and CategorySeeder)
func (s *ProductSeeder) Dependencies() []string {
    return []string{"UserSeeder", "CategorySeeder"}
}
```

When you run `make db-seed`, seeders execute in dependency order:

1. UserSeeder
2. CategorySeeder
3. ProductSeeder

### **🏭 Database Factory Pattern**

Switch databases without changing application code:

```go
// Automatic database selection based on configuration
factory := database.NewDatabaseFactory()
db, err := factory.CreateDatabase(config.GetDatabaseConfig())

// Same interface, different implementations
db.RunMigrations()  // Works with MySQL, PostgreSQL, SQLite
db.SeedData("")     // Works with any database
```

### **⚡ Hot Reload Development**

Automatic code reloading during development:

```bash
# Uses Air (preferred) or CompileDaemon
make dev

# Manual reload
make dev-force  # Kills port conflicts first
```

### **📊 Health Monitoring**

Built-in health checks for monitoring:

```bash
# Application health
curl http://localhost:8080/health

# Database health
curl http://localhost:8080/health/db

# Check from command line
make health
make status
```

---

## 🧪 Testing

### **Unit Tests**

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Test specific package
go test ./internal/entity/...
```

### **Database Testing**

```bash
# Test all database types
make test-all-db

# Test specific database
DB_DRIVER=sqlite make migrate-status
DB_DRIVER=mysql make migrate-status
DB_DRIVER=postgresql make migrate-status
```

---

## 🐳 Docker Support

### **Docker Compose**

```bash
# Build and run containers
make docker-build
make docker-run

# View logs
make docker-logs

# Stop containers
make docker-stop
```

### **Dockerfile**

The included `Dockerfile` creates an optimized production image:

- Multi-stage build for smaller image size
- Non-root user for security
- Health checks included
- Environment-based configuration

---

## 📝 Environment Configuration

### **Core Settings**

```env
# Application
APP_NAME=flex-service
ENV=development
TIMEZONE=Asia/Bangkok

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Type Selection
DB_DRIVER=sqlite  # mysql, postgresql, sqlite
```

### **Database-Specific Settings**

Each database type has its own configuration section in `env.example`. Only the settings for your selected `DB_DRIVER` are used.

### **Production Considerations**

- Use PostgreSQL for production workloads
- Enable SSL/TLS for database connections
- Set appropriate connection pool sizes
- Configure proper logging levels
- Use environment-specific configurations

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 📚 Examples & Documentation

### **🎯 Practical Examples**

- **[Multi-Database Examples](./examples/multi-database/)** - Comprehensive usage examples
- **[E-commerce System](./examples/multi-database/ecommerce-example.md)** - Complete e-commerce implementation
- **[Production Deployment](./examples/multi-database/production-example.md)** - Production-ready deployment guide
- **[Testing Strategies](./examples/multi-database/testing-example.md)** - Multi-database testing approach

### **🛠️ Technical Documentation**

- **[Database Drivers](./pkg/database/)** - Multi-database implementation details
- **[Migration Engine](./pkg/migration/)** - Reusable migration system
- **[Seeder Engine](./pkg/seeder/)** - Dependency-aware seeding system
- **[Container System](./internal/container/)** - Modern dependency injection with Factory Pattern
- **[Error Helpers](./pkg/errors/)** - Convenient error handling functions
- **[Authentication](./pkg/auth/)** - JWT authentication and authorization
- **[Clean Architecture](./internal/)** - Project structure and patterns

---

## 🔒 Security & Authentication

Go Starter includes comprehensive security and authentication features:

### **🔐 JWT Authentication System**

- 🎫 **Access Tokens** - Short-lived tokens for API access
- 🔄 **Refresh Tokens** - Long-lived tokens for renewal
- 👤 **User Registration/Login** - Complete authentication flow
- 🚪 **Logout** - Secure token invalidation
- 🛡️ **Role-based Authorization** - Permission-based access control

### **🛡️ Security Features**

- 🛡️ **Rate Limiting** - DDoS protection (Redis-based)
- 🔐 **Session Management** - Secure Redis sessions
- 🌐 **CORS Protection** - Cross-origin request filtering
- 🛡️ **Security Headers** - Helmet middleware
- ✅ **Input Validation** - Request data validation
- 📊 **Audit Logging** - Security event tracking
- 🔒 **Password Hashing** - bcrypt with salt rounds

### **🚀 Quick Authentication Setup**

```bash
# Your authentication system is ready to use!
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","username":"user","password":"MySecure123!","first_name":"John","last_name":"Doe"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_username":"user@example.com","password":"MySecure123!"}'

# Use the access token for protected endpoints
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

For security issues, please see [SECURITY.md](SECURITY.md).

## 🙏 Acknowledgments

- **Laravel** - For the amazing Artisan CLI inspiration
- **Gin** - For the excellent HTTP framework
- **GORM** - For the powerful ORM
- **Zap** - For structured logging
- **Redis** - For high-performance caching and sessions
- **Go Community** - For the incredible ecosystem

---

## 📞 Support

- 📖 **Documentation**: Check the `make help` and `make examples` commands
- 🐛 **Issues**: Report bugs via GitHub Issues
- 💬 **Discussions**: Join GitHub Discussions for questions
- 📧 **Contact**: [Your contact information]

---

**Happy coding! 🚀**

> Built with ❤️ for the Go community. This starter template helps you build production-ready APIs with Laravel-style development experience and multi-database flexibility.
