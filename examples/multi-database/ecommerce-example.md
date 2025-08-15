# 🛒 E-commerce Multi-Database Example

This example demonstrates building an e-commerce system that can run on different databases.

## 🏗️ System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Development   │    │     Staging     │    │   Production    │
│     SQLite      │    │      MySQL      │    │  PostgreSQL     │
│   (Fast & Easy) │    │  (Compatibility)│    │ (Performance)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📋 Step-by-Step Implementation

### **Step 1: Create Models**

```bash
# Create User model
make make-model NAME=User TABLE=users FIELDS="name:string,email:string,phone:string"

# Create Category model
make make-model NAME=Category TABLE=categories FIELDS="name:string,description:text,slug:string"

# Create Product model
make make-model NAME=Product TABLE=products FIELDS="name:string,description:text,price:decimal,category_id:uuid,sku:string,stock:int"

# Create Order model
make make-model NAME=Order TABLE=orders FIELDS="user_id:uuid,total:decimal,status:string,notes:text"

# Create OrderItem model
make make-model NAME=OrderItem TABLE=order_items FIELDS="order_id:uuid,product_id:uuid,quantity:int,price:decimal"
```

### **Step 2: Create Seeders with Dependencies**

```bash
# Create seeders with proper dependency chain
make make-seeder NAME=UserSeeder TABLE=users

make make-seeder NAME=CategorySeeder TABLE=categories

make make-seeder NAME=ProductSeeder TABLE=products DEPS="CategorySeeder"

make make-seeder NAME=OrderSeeder TABLE=orders DEPS="UserSeeder,ProductSeeder"

make make-seeder NAME=OrderItemSeeder TABLE=order_items DEPS="OrderSeeder"
```

### **Step 3: Development with SQLite**

```bash
# Setup development environment
export DB_DRIVER=sqlite
export DB_SQLITE_FILE_PATH=./ecommerce_dev.db

# Run migrations and seeders
make migrate
make db-seed

# Start development server
make dev
```

**Development Benefits:**

- ⚡ Fast startup (no external DB)
- 🔧 Easy reset (`rm ecommerce_dev.db`)
- 📱 Portable (single file)

### **Step 4: Staging with MySQL**

```bash
# Setup staging environment
export DB_DRIVER=mysql
export DB_MYSQL_HOST=staging-mysql.company.com
export DB_MYSQL_NAME=ecommerce_staging
export DB_MYSQL_USER=staging_user
export DB_MYSQL_PASSWORD=staging_pass

# Deploy to staging
make migrate
make db-seed

# Run staging tests
make test
```

**Staging Benefits:**

- 🔍 MySQL-specific testing
- 📊 Performance validation
- 🔒 Security testing

### **Step 5: Production with PostgreSQL**

```bash
# Setup production environment
export DB_DRIVER=postgresql
export DB_POSTGRES_HOST=prod-postgres.company.com
export DB_POSTGRES_NAME=ecommerce_prod
export DB_POSTGRES_USER=prod_user
export DB_POSTGRES_PASSWORD=prod_pass
export DB_POSTGRES_SSL_MODE=require

# Deploy to production
make migrate

# Seed only essential data (no test data)
make db-seed-specific NAME=CategorySeeder
```

**Production Benefits:**

- 🚀 High performance
- 📈 Excellent concurrency
- 🔧 Advanced features (JSONB, arrays)

## 🗄️ Database-Specific Optimizations

### **SQLite Optimizations**

```sql
-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;

-- Enable foreign keys
PRAGMA foreign_keys=ON;

-- Optimize for speed
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=10000;
```

### **MySQL Optimizations**

```sql
-- Use InnoDB engine
ENGINE=InnoDB;

-- Optimize for e-commerce
SET innodb_buffer_pool_size = 1G;
SET query_cache_size = 256M;

-- Enable full-text search
ALTER TABLE products ADD FULLTEXT(name, description);
```

### **PostgreSQL Optimizations**

```sql
-- Use JSONB for flexible product attributes
ALTER TABLE products ADD COLUMN attributes JSONB;

-- Create GIN index for JSONB
CREATE INDEX idx_products_attributes ON products USING GIN (attributes);

-- Use arrays for tags
ALTER TABLE products ADD COLUMN tags TEXT[];
```

## 📊 Sample Data Structure

### **Categories**

```json
[
  { "name": "Electronics", "slug": "electronics" },
  { "name": "Clothing", "slug": "clothing" },
  { "name": "Books", "slug": "books" }
]
```

### **Products**

```json
[
  {
    "name": "iPhone 15",
    "price": 999.0,
    "category": "Electronics",
    "sku": "IPH15-128",
    "stock": 50
  },
  {
    "name": "MacBook Pro",
    "price": 1999.0,
    "category": "Electronics",
    "sku": "MBP-M3-14",
    "stock": 25
  }
]
```

## 🧪 Testing Strategy

### **Unit Tests (SQLite)**

```bash
# Fast unit tests with in-memory SQLite
export DB_DRIVER=sqlite
export DB_SQLITE_IN_MEMORY=true
make test
```

### **Integration Tests (MySQL)**

```bash
# Integration tests with MySQL
export DB_DRIVER=mysql
export DB_MYSQL_NAME=ecommerce_test
make test-integration
```

### **Performance Tests (PostgreSQL)**

```bash
# Performance tests with PostgreSQL
export DB_DRIVER=postgresql
export DB_POSTGRES_NAME=ecommerce_perf
make test-performance
```

## 🚀 Deployment Pipeline

```yaml
# .github/workflows/deploy.yml
name: E-commerce Deployment

on: [push]

jobs:
  test-sqlite:
    runs-on: ubuntu-latest
    env:
      DB_DRIVER: sqlite
      DB_SQLITE_IN_MEMORY: true
    steps:
      - uses: actions/checkout@v2
      - name: Run SQLite tests
        run: make test

  test-mysql:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: test
          MYSQL_DATABASE: ecommerce_test
    env:
      DB_DRIVER: mysql
      DB_MYSQL_HOST: localhost
      DB_MYSQL_NAME: ecommerce_test
      DB_MYSQL_USER: root
      DB_MYSQL_PASSWORD: test
    steps:
      - uses: actions/checkout@v2
      - name: Run MySQL tests
        run: make test

  deploy-production:
    needs: [test-sqlite, test-mysql]
    runs-on: ubuntu-latest
    env:
      DB_DRIVER: postgresql
    steps:
      - uses: actions/checkout@v2
      - name: Deploy to production
        run: |
          make migrate
          make db-seed-specific NAME=CategorySeeder
```

## 📈 Performance Metrics

### **Database Comparison for E-commerce**

| Metric            | SQLite | MySQL     | PostgreSQL |
| ----------------- | ------ | --------- | ---------- |
| Product Search    | 50ms   | 20ms      | 15ms       |
| Order Creation    | 30ms   | 25ms      | 20ms       |
| Report Generation | 200ms  | 100ms     | 80ms       |
| Concurrent Users  | 10     | 100       | 1000       |
| Data Size Limit   | 281TB  | Unlimited | Unlimited  |

### **Scaling Recommendations**

- **< 1000 products**: SQLite is sufficient
- **< 100K products**: MySQL is recommended
- **> 100K products**: PostgreSQL is optimal

## 🔧 Configuration Examples

### **Development (.env.development)**

```env
DB_DRIVER=sqlite
DB_SQLITE_FILE_PATH=./ecommerce_dev.db
DB_SQLITE_FOREIGN_KEYS=true
LOG_LEVEL=debug
```

### **Staging (.env.staging)**

```env
DB_DRIVER=mysql
DB_MYSQL_HOST=staging-db.company.com
DB_MYSQL_NAME=ecommerce_staging
DB_MYSQL_USER=staging_user
DB_MYSQL_PASSWORD=${MYSQL_STAGING_PASSWORD}
LOG_LEVEL=info
```

### **Production (.env.production)**

```env
DB_DRIVER=postgresql
DB_POSTGRES_HOST=prod-db.company.com
DB_POSTGRES_NAME=ecommerce_prod
DB_POSTGRES_USER=prod_user
DB_POSTGRES_PASSWORD=${POSTGRES_PROD_PASSWORD}
DB_POSTGRES_SSL_MODE=require
LOG_LEVEL=warn
```

## 🎯 Key Benefits

1. **Flexible Development** - Start fast with SQLite
2. **Gradual Scaling** - Move to MySQL when needed
3. **Production Ready** - Scale to PostgreSQL
4. **Same Codebase** - No database-specific code
5. **Easy Testing** - Test on multiple databases
6. **Cost Effective** - Pay for what you need

## 🔗 Next Steps

- [Blog System Example](./blog-example.md)
- [Production Deployment](./production-example.md)
- [Testing Strategies](./testing-example.md)
