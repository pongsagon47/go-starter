# ⚡ Quick Start Guide

Get up and running with Go Starter in 5 minutes!

## 🚀 1-Minute Setup

```bash
# Clone and setup
git clone https://github.com/pongsagon47/go-starter.git my-api
cd my-api
make setup

# Start with SQLite (no external dependencies)
make run
```

**That's it! Your API is running at http://localhost:8080** 🎉

## 📋 2-Minute Demo

### **Test Your API**

```bash
# Health check
curl http://localhost:8080/health

# Database health
curl http://localhost:8080/health/db

# Demo endpoint
curl http://localhost:8080/api/v1/demo
```

### **Try the Database**

```bash
# Check migration status
make migrate-status

# List available seeders
make db-seed-list

# Run seeders
make db-seed
```

## 🗄️ 3-Minute Multi-Database

### **Switch to MySQL**

```bash
# Stop current server (Ctrl+C)

# Setup MySQL in .env
export DB_TYPE=mysql
export DB_MYSQL_HOST=localhost
export DB_MYSQL_NAME=my_api

# Run migrations and start
make migrate
make db-seed
make run
```

### **Switch to PostgreSQL**

```bash
# Stop current server (Ctrl+C)

# Setup PostgreSQL in .env
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=localhost
export DB_POSTGRES_NAME=my_api

# Run migrations and start
make migrate
make db-seed
make run
```

## 🎨 5-Minute Laravel-style Development

### **Create Your First Model**

```bash
# Create complete User model stack
make make-model NAME=User TABLE=users FIELDS="name:string,email:string,age:int"

# This creates:
# ✅ internal/entity/user.go (Model)
# ✅ internal/migrations/TIMESTAMP_create_users_table.go (Migration)
# ✅ internal/seeders/user_seeder.go (Seeder)
```

### **Create a Package**

```bash
# Create complete package structure
make make-package NAME=User

# This creates:
# ✅ internal/user/handler.go (HTTP handlers)
# ✅ internal/user/usecase.go (Business logic)
# ✅ internal/user/repository.go (Data access)
# ✅ internal/user/port.go (Interfaces)
```

### **Run Everything**

```bash
# Run migrations
make migrate

# Seed data
make db-seed

# Start with hot reload
make dev
```

## 🧪 Test Different Databases

```bash
# Test all databases at once
make test-all-db

# Or test individually
make db-sqlite migrate-status
make db-mysql migrate-status
make db-postgres migrate-status
```

## 📚 What's Next?

### **Learn More**

- **[Multi-Database Examples](./multi-database/)** - Detailed examples
- **[E-commerce Tutorial](./multi-database/ecommerce-example.md)** - Build a complete system
- **[Production Guide](./multi-database/production-example.md)** - Deploy to production

### **Explore Features**

- **Laravel-style CLI** - `make help`
- **Multi-database support** - Switch databases without code changes
- **Clean Architecture** - Well-organized, testable code
- **Hot Reload** - Fast development cycle

### **Get Help**

- **Commands:** `make help`
- **Examples:** `make examples`
- **Documentation:** [README.md](../README.md)

## 🎯 Common Commands

```bash
# Development
make dev                    # Start with hot reload
make build                  # Build application
make test                   # Run tests

# Database
make migrate                # Run migrations
make migrate-status         # Check status
make db-seed               # Seed data
make db-info               # Show database info

# Code Generation
make make-model NAME=Post TABLE=posts FIELDS="title:string,content:text"
make make-seeder NAME=PostSeeder TABLE=posts
make make-package NAME=Post

# Multi-Database
make db-sqlite migrate      # Use SQLite
make db-mysql migrate       # Use MySQL
make db-postgres migrate    # Use PostgreSQL
```

---

**🎉 You're ready to build amazing APIs with Go Starter!**

**Happy coding! 🚀**
