# 🗄️ Multi-Database Examples

This directory contains practical examples of using Go Starter with different database types.

## 🚀 Quick Examples

### **1. Switch Database Types**

```bash
# Development with SQLite
export DB_DRIVER=sqlite
make migrate
make db-seed
make run

# Staging with MySQL
export DB_DRIVER=mysql
make migrate
make db-seed
make run

# Production with PostgreSQL
export DB_DRIVER=postgresql
make migrate
make db-seed
make run
```

### **2. Using Makefile Commands**

```bash
# SQLite Development
make db-sqlite migrate-status
make db-sqlite migrate
make db-sqlite db-seed

# MySQL Staging
make db-mysql migrate-status
make db-mysql migrate
make db-mysql db-seed

# PostgreSQL Production
make db-postgres migrate-status
make db-postgres migrate
make db-postgres db-seed
```

### **3. Test All Databases**

```bash
# Test migration status on all databases
make test-all-db

# Manual testing
DB_DRIVER=sqlite make migrate-status
DB_DRIVER=mysql make migrate-status
DB_DRIVER=postgresql make migrate-status
```

## 📋 Detailed Examples

### **Example 1: E-commerce Setup**

See: [ecommerce-example.md](./ecommerce-example.md)

### **Example 2: Blog System**

See: [blog-example.md](./blog-example.md)

### **Example 3: Production Deployment**

See: [production-example.md](./production-example.md)

### **Example 4: Testing Strategy**

See: [testing-example.md](./testing-example.md)

## 🔧 Configuration Examples

### **SQLite (Development)**

```env
DB_DRIVER=sqlite
DB_SQLITE_FILE_PATH=./dev.db
DB_SQLITE_FOREIGN_KEYS=true
DB_SQLITE_JOURNAL=WAL
```

### **MySQL (Staging)**

```env
DB_DRIVER=mysql
DB_MYSQL_HOST=staging-mysql.company.com
DB_MYSQL_PORT=3306
DB_MYSQL_USER=staging_user
DB_MYSQL_PASSWORD=staging_pass
DB_MYSQL_NAME=staging_db
DB_MYSQL_SSL_MODE=require
```

### **PostgreSQL (Production)**

```env
DB_DRIVER=postgresql
DB_POSTGRES_HOST=prod-postgres.company.com
DB_POSTGRES_PORT=5432
DB_POSTGRES_USER=prod_user
DB_POSTGRES_PASSWORD=prod_pass
DB_POSTGRES_NAME=prod_db
DB_POSTGRES_SSL_MODE=require
DB_POSTGRES_TIMEZONE=UTC
```

## 🧪 Testing Scenarios

### **Scenario 1: Local Development**

1. Use SQLite for fast local development
2. No external dependencies
3. Easy database reset

### **Scenario 2: CI/CD Pipeline**

1. Use SQLite for unit tests (fast)
2. Use MySQL for integration tests
3. Use PostgreSQL for production-like tests

### **Scenario 3: Multi-Environment**

1. Development: SQLite
2. Staging: MySQL
3. Production: PostgreSQL

## 🔄 Migration Strategies

### **Cross-Database Migrations**

```bash
# Create migration that works on all databases
make make-migration NAME=create_universal_table CREATE=true TABLE=products

# Test on all databases
make db-sqlite migrate
make db-mysql migrate
make db-postgres migrate
```

### **Database-Specific Features**

- **MySQL**: Use `JSON` columns, full-text search
- **PostgreSQL**: Use `JSONB`, arrays, custom types
- **SQLite**: Use simple types, file-based storage

## 📊 Performance Comparison

| Feature     | SQLite     | MySQL    | PostgreSQL |
| ----------- | ---------- | -------- | ---------- |
| Setup       | ⭐⭐⭐⭐⭐ | ⭐⭐⭐   | ⭐⭐       |
| Performance | ⭐⭐⭐     | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Concurrency | ⭐⭐       | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Features    | ⭐⭐⭐     | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Scalability | ⭐⭐       | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

## 🚀 Best Practices

### **Development**

- Use SQLite for rapid prototyping
- Use in-memory SQLite for tests
- Keep database schema simple initially

### **Staging**

- Use MySQL to catch MySQL-specific issues
- Test with realistic data volumes
- Validate performance characteristics

### **Production**

- Use PostgreSQL for complex applications
- Enable SSL/TLS connections
- Configure connection pooling
- Set up monitoring and backups

## 🔗 Related Documentation

- [Main README](../../README.md)
- [Database Configuration](../../config/README.md)
- [Migration Guide](../migrations/README.md)
- [Seeder Guide](../seeders/README.md)
