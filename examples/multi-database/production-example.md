# 🏭 Production Multi-Database Deployment

This example shows how to deploy Go Starter in production with different database configurations.

## 🎯 Production Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Production Environment                    │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Application   │    Database     │       Monitoring        │
│    Servers      │    Cluster      │       & Logging         │
│                 │                 │                         │
│  ┌─────────┐    │  ┌─────────┐    │  ┌─────────────────┐    │
│  │Go Starter│    │  │PostgreSQL│   │  │   Prometheus    │    │
│  │ Instance │    │  │  Primary  │   │  │    Grafana      │    │
│  │    #1    │    │  │           │   │  │      ELK        │    │
│  └─────────┘    │  └─────────┘    │  └─────────────────┘    │
│                 │                 │                         │
│  ┌─────────┐    │  ┌─────────┐    │  ┌─────────────────┐    │
│  │Go Starter│    │  │PostgreSQL│   │  │   Health Checks │    │
│  │ Instance │    │  │ Replica   │   │  │   Alerting      │    │
│  │    #2    │    │  │           │   │  │   Backup        │    │
│  └─────────┘    │  └─────────┘    │  └─────────────────┘    │
└─────────────────┴─────────────────┴─────────────────────────┘
```

## 🚀 Deployment Strategies

### **Strategy 1: Single Database**

**Best for:** Small to medium applications

```bash
# Production with PostgreSQL
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=prod-db.company.com
export DB_POSTGRES_PORT=5432
export DB_POSTGRES_NAME=app_production
export DB_POSTGRES_USER=app_user
export DB_POSTGRES_PASSWORD=${DB_PASSWORD}
export DB_POSTGRES_SSL_MODE=require

# Deploy
make migrate
make build
./bin/go-starter
```

### **Strategy 2: Multi-Environment Pipeline**

**Best for:** Enterprise applications

```bash
# Development
DB_TYPE=sqlite make migrate && make test

# Staging
DB_TYPE=mysql make migrate && make test-integration

# Production
DB_TYPE=postgresql make migrate && make deploy
```

### **Strategy 3: Multi-Region Deployment**

**Best for:** Global applications

```bash
# US Region - PostgreSQL
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=us-east-1-postgres.company.com

# EU Region - PostgreSQL
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=eu-west-1-postgres.company.com

# Asia Region - PostgreSQL
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=ap-southeast-1-postgres.company.com
```

## 🐳 Docker Deployment

### **Dockerfile.production**

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/go-starter cmd/main.go
RUN go build -o bin/artisan cmd/artisan/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bin/go-starter .
COPY --from=builder /app/bin/artisan .
COPY --from=builder /app/env.example .env

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./go-starter"]
```

### **docker-compose.production.yml**

```yaml
version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.production
    ports:
      - "8080:8080"
    environment:
      - DB_TYPE=postgresql
      - DB_POSTGRES_HOST=postgres
      - DB_POSTGRES_NAME=app_production
      - DB_POSTGRES_USER=app_user
      - DB_POSTGRES_PASSWORD=${DB_PASSWORD}
      - ENV=production
      - LOG_LEVEL=info
    depends_on:
      - postgres
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=app_production
      - POSTGRES_USER=app_user
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app_user -d app_production"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  postgres_data:
```

## ☸️ Kubernetes Deployment

### **k8s-deployment.yaml**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-starter
  labels:
    app: go-starter
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-starter
  template:
    metadata:
      labels:
        app: go-starter
    spec:
      containers:
        - name: go-starter
          image: go-starter:latest
          ports:
            - containerPort: 8080
          env:
            - name: DB_TYPE
              value: "postgresql"
            - name: DB_POSTGRES_HOST
              value: "postgres-service"
            - name: DB_POSTGRES_NAME
              value: "app_production"
            - name: DB_POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: username
            - name: DB_POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: password
            - name: ENV
              value: "production"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: go-starter-service
spec:
  selector:
    app: go-starter
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
```

## 🔧 Configuration Management

### **Production Environment Variables**

```bash
# Application
export APP_NAME=go-starter-prod
export ENV=production
export SERVER_PORT=8080
export LOG_LEVEL=info

# Database
export DB_TYPE=postgresql
export DB_POSTGRES_HOST=${POSTGRES_HOST}
export DB_POSTGRES_PORT=5432
export DB_POSTGRES_NAME=app_production
export DB_POSTGRES_USER=${POSTGRES_USER}
export DB_POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
export DB_POSTGRES_SSL_MODE=require
export DB_POSTGRES_MAX_OPEN_CONNS=100
export DB_POSTGRES_MAX_IDLE_CONNS=10

# Security
export JWT_SECRET=${JWT_SECRET_KEY}
export ENCRYPTION_KEY=${ENCRYPTION_KEY}

# Email
export SMTP_HOST=${SMTP_HOST}
export SMTP_USER=${SMTP_USER}
export SMTP_PASSWORD=${SMTP_PASSWORD}
```

### **Database Configuration Tuning**

```sql
-- PostgreSQL production tuning
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- Reload configuration
SELECT pg_reload_conf();
```

## 📊 Monitoring & Observability

### **Health Check Endpoints**

```go
// Add to router
router.GET("/health", healthCheck)
router.GET("/health/db", databaseHealthCheck)
router.GET("/metrics", prometheusMetrics)
```

### **Prometheus Metrics**

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "go-starter"
    static_configs:
      - targets: ["app:8080"]
    metrics_path: /metrics
    scrape_interval: 10s
```

### **Grafana Dashboard**

```json
{
  "dashboard": {
    "title": "Go Starter Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "go_sql_open_connections"
          }
        ]
      }
    ]
  }
}
```

## 🔒 Security Configuration

### **SSL/TLS Setup**

```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name api.company.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://go-starter:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### **Database Security**

```sql
-- Create dedicated user
CREATE USER app_user WITH PASSWORD 'strong_password';

-- Grant minimal permissions
GRANT CONNECT ON DATABASE app_production TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;

-- Enable SSL
ALTER SYSTEM SET ssl = on;
ALTER SYSTEM SET ssl_cert_file = '/etc/ssl/certs/server.crt';
ALTER SYSTEM SET ssl_key_file = '/etc/ssl/private/server.key';
```

## 🚀 CI/CD Pipeline

### **GitHub Actions Workflow**

```yaml
name: Production Deployment

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        db: [sqlite, mysql, postgresql]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: "1.23"

      - name: Test with ${{ matrix.db }}
        run: |
          export DB_TYPE=${{ matrix.db }}
          make test

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: |
          docker build -f Dockerfile.production -t go-starter:${{ github.sha }} .
          docker tag go-starter:${{ github.sha }} go-starter:latest

      - name: Deploy to production
        run: |
          # Deploy to Kubernetes
          kubectl set image deployment/go-starter go-starter=go-starter:${{ github.sha }}
          kubectl rollout status deployment/go-starter
```

## 📈 Performance Optimization

### **Connection Pooling**

```env
# Optimize for production load
DB_POSTGRES_MAX_OPEN_CONNS=100
DB_POSTGRES_MAX_IDLE_CONNS=25
DB_POSTGRES_CONN_MAX_LIFETIME=300

# Monitor connections
DB_POSTGRES_LOG_LEVEL=warn
```

### **Caching Strategy**

```go
// Add Redis caching
type CacheService struct {
    redis *redis.Client
}

// Cache database queries
func (c *CacheService) GetUser(id string) (*User, error) {
    // Check cache first
    cached, err := c.redis.Get(ctx, "user:"+id).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // Fallback to database
    user, err := c.db.GetUser(id)
    if err != nil {
        return nil, err
    }

    // Cache result
    data, _ := json.Marshal(user)
    c.redis.Set(ctx, "user:"+id, data, time.Hour)

    return user, nil
}
```

## 🔄 Backup & Recovery

### **Database Backup**

```bash
#!/bin/bash
# backup.sh

# PostgreSQL backup
pg_dump -h $DB_POSTGRES_HOST -U $DB_POSTGRES_USER -d $DB_POSTGRES_NAME \
    --no-password --verbose --clean --no-owner --no-privileges \
    --format=custom > backup_$(date +%Y%m%d_%H%M%S).dump

# Upload to S3
aws s3 cp backup_*.dump s3://company-backups/go-starter/
```

### **Disaster Recovery**

```bash
#!/bin/bash
# restore.sh

# Download latest backup
aws s3 cp s3://company-backups/go-starter/latest.dump ./

# Restore database
pg_restore -h $DB_POSTGRES_HOST -U $DB_POSTGRES_USER -d $DB_POSTGRES_NAME \
    --verbose --clean --no-owner --no-privileges ./latest.dump

# Verify restoration
make migrate-status
```

## 📋 Production Checklist

### **Pre-Deployment**

- [ ] Environment variables configured
- [ ] SSL certificates installed
- [ ] Database tuned for production
- [ ] Monitoring setup complete
- [ ] Backup strategy implemented
- [ ] Security audit passed
- [ ] Load testing completed

### **Post-Deployment**

- [ ] Health checks passing
- [ ] Metrics collecting
- [ ] Logs aggregating
- [ ] Alerts configured
- [ ] Performance baseline established
- [ ] Documentation updated

## 🎯 Key Benefits

1. **Database Flexibility** - Switch databases without code changes
2. **Horizontal Scaling** - Multiple app instances
3. **High Availability** - Database replication
4. **Monitoring** - Comprehensive observability
5. **Security** - SSL/TLS, secrets management
6. **Automation** - CI/CD pipeline
7. **Disaster Recovery** - Backup and restore procedures

## 🔗 Related Examples

- [E-commerce Example](./ecommerce-example.md)
- [Testing Strategies](./testing-example.md)
- [Blog System Example](./blog-example.md)
