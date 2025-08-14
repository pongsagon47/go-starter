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

- **Migrations** - Version control for database schema
- **Seeders** - Database seeding with dependency resolution
- **Code Generation** - Auto-generate models, packages, migrations
- **Database Management** - Create, drop, reset databases

### 🏗️ **Clean Architecture**

- **Separation of Concerns** - Entity, Repository, Usecase, Handler layers
- **Dependency Injection** - Container-based DI system
- **Interface-based Design** - Easy testing and mocking
- **Domain-driven Design** - Business logic in the center

### ⚡ **Production Ready**

- **Hot Reload** - Air/CompileDaemon integration
- **Health Checks** - Database and application monitoring
- **Error Handling** - Comprehensive error management
- **Logging** - Structured logging with Zap
- **Security** - Helmet, CORS, input validation
- **Email** - SMTP integration with templates

---

## 🚀 Quick Start

### 1. **Setup Project**

```bash
# Clone and setup
git clone https://github.com/pongsagon47/go-starter.git go-starter
cd go-starter

# Install dependencies and setup
make setup

# Copy environment configuration
cp env.example .env
# Edit .env with your database settings
```

### 2. **Choose Your Database**

```bash
# SQLite (default, no setup required)
export DB_TYPE=sqlite

# MySQL
export DB_TYPE=mysql
# Configure DB_MYSQL_* variables in .env

# PostgreSQL
export DB_TYPE=postgresql
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

# Demo endpoint
curl http://localhost:8080/api/v1/demo
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
DB_TYPE=sqlite
DB_SQLITE_FILE_PATH=./database.db
DB_SQLITE_FOREIGN_KEYS=true
DB_SQLITE_JOURNAL=WAL
```

**MySQL** (Staging/Production)

```env
DB_TYPE=mysql
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
DB_TYPE=postgresql
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
go-starter/
├── cmd/
│   ├── main.go                 # Application entry point
│   └── artisan/main.go         # Laravel-style CLI tool
├── config/
│   └── config.go               # Multi-database configuration
├── internal/
│   ├── entity/                 # Domain entities
│   ├── migrations/             # Database migrations
│   ├── seeders/               # Database seeders
│   ├── container/             # Dependency injection
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
make make-migration     # Create new migration file
make make-seeder        # Create seeder with dependency support
make make-entity        # Create new entity/model file
make make-package       # Create new package (handler, usecase, repository, port)
make make-model         # Create complete model stack (entity + migration + seeder)
```

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
DB_TYPE=sqlite make migrate-status
DB_TYPE=mysql make migrate-status
DB_TYPE=postgresql make migrate-status
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
APP_NAME=go-starter
ENV=development
TIMEZONE=Asia/Bangkok

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Type Selection
DB_TYPE=sqlite  # mysql, postgresql, sqlite
```

### **Database-Specific Settings**

Each database type has its own configuration section in `env.example`. Only the settings for your selected `DB_TYPE` are used.

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
- **[Clean Architecture](./internal/)** - Project structure and patterns

---

## 🔒 Security

Go Starter includes built-in security features:

- 🛡️ **Rate Limiting** - DDoS protection (Redis-based)
- 🔐 **Session Management** - Secure Redis sessions
- 🌐 **CORS Protection** - Cross-origin request filtering
- 🛡️ **Security Headers** - Helmet middleware
- ✅ **Input Validation** - Request data validation
- 📊 **Audit Logging** - Security event tracking

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
