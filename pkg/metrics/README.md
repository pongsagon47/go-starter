# 📊 Metrics Package

Comprehensive monitoring and metrics collection system with counters, gauges, histograms, health checks, and middleware for tracking application performance and system health.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Metrics Types](#metrics-types)
- [Health Checks](#health-checks)
- [Middleware](#middleware)
- [System Monitoring](#system-monitoring)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/metrics"
```

## ⚡ Quick Start

### Basic Metrics Collection

```go
package main

import (
    "context"
    "time"
    "go-starter/pkg/metrics"
)

func main() {
    // Create different types of metrics
    requestCounter := metrics.NewCounter("requests_total", "Total requests", metrics.Labels{
        "service": "api",
    })

    responseTimeHistogram := metrics.NewHistogram("response_time_seconds", "Response times", nil, metrics.Labels{
        "endpoint": "/api/users",
    })

    activeUsersGauge := metrics.NewGauge("active_users", "Currently active users", metrics.Labels{})

    // Use metrics
    requestCounter.Inc()
    responseTimeHistogram.Observe(0.250) // 250ms
    activeUsersGauge.Set(142)

    // Start system metrics collection
    metrics.StartSystemMetrics(30 * time.Second)

    // Register health checks
    metrics.RegisterHealthCheck(
        metrics.NewMemoryHealthChecker("memory", 80.0, 95.0),
    )

    // Get all metrics
    allMetrics := metrics.GetAllMetrics()
    fmt.Printf("Collected %d metrics\n", len(allMetrics))

    // Check system health
    health, _ := metrics.GetOverallHealth(context.Background())
    fmt.Printf("System health: %s\n", health)
}
```

## 📏 Metrics Types

### **Counter - Monotonically Increasing Values**

```go
// Create counter
requestsCounter := metrics.NewCounter("http_requests_total", "Total HTTP requests", metrics.Labels{
    "service": "api",
})

// Use counter
requestsCounter.Inc()                    // Increment by 1
requestsCounter.Add(5)                   // Add specific value
currentValue := requestsCounter.Get()    // Get current value

// Counter with labels
authCounter := requestsCounter.With(metrics.Labels{
    "endpoint": "/auth/login",
    "method":   "POST",
})
authCounter.Inc()
```

### **Gauge - Values That Can Go Up and Down**

```go
// Create gauge
activeUsersGauge := metrics.NewGauge("active_users", "Number of active users", metrics.Labels{})

// Use gauge
activeUsersGauge.Set(100)                // Set to specific value
activeUsersGauge.Inc()                   // Increment by 1
activeUsersGauge.Dec()                   // Decrement by 1
activeUsersGauge.Add(10)                 // Add value
activeUsersGauge.Sub(5)                  // Subtract value
currentValue := activeUsersGauge.Get()   // Get current value

// Memory usage gauge
memoryGauge := metrics.NewGauge("memory_usage_bytes", "Memory usage in bytes", metrics.Labels{
    "type": "heap",
})
memoryGauge.Set(float64(memStats.Alloc))
```

### **Histogram - Distribution of Values**

```go
// Create histogram with custom buckets
responseTimeHistogram := metrics.NewHistogram(
    "response_time_seconds",
    "HTTP response time in seconds",
    []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0, 5.0}, // Custom buckets
    metrics.Labels{},
)

// Use histogram
responseTimeHistogram.Observe(0.250)                    // Observe 250ms
responseTimeHistogram.ObserveDuration(time.Now())       // Observe duration since now

// Request size histogram
requestSizeHistogram := metrics.NewHistogram(
    "request_size_bytes",
    "HTTP request size in bytes",
    []float64{100, 1000, 10000, 100000}, // Size buckets
    metrics.Labels{},
)
requestSizeHistogram.Observe(1500) // 1.5KB request
```

### **Timer - Measure Durations**

```go
// Create timer from histogram
responseTimeHistogram := metrics.NewHistogram("response_time", "Response times", nil, metrics.Labels{})
timer := metrics.NewTimer(responseTimeHistogram)

// Method 1: Start/stop pattern
stop := timer.Start()
// ... do work ...
stop() // Records duration

// Method 2: Observe duration
start := time.Now()
// ... do work ...
timer.ObserveDuration(start)

// Method 3: Measure function execution
func MeasuredFunction() {
    stop := timer.Start()
    defer stop()

    // Function logic here
    time.Sleep(100 * time.Millisecond)
}
```

## 🏥 Health Checks

### **Built-in Health Checkers**

```go
// Database health check
db := getDatabase() // Your database instance
dbChecker := metrics.NewDatabaseHealthChecker("primary_db", db)
metrics.RegisterHealthCheck(dbChecker)

// Redis health check
redisClient := getRedisClient() // Your Redis client
redisChecker := metrics.NewRedisHealthChecker("cache", redisClient)
metrics.RegisterHealthCheck(redisChecker)

// HTTP service health check
serviceChecker := metrics.NewHTTPServiceHealthChecker(
    "external_api",
    "https://api.external.com/health",
)
metrics.RegisterHealthCheck(serviceChecker)

// Memory health check
memoryChecker := metrics.NewMemoryHealthChecker(
    "memory",
    80.0,  // Warning threshold (80%)
    95.0,  // Critical threshold (95%)
)
metrics.RegisterHealthCheck(memoryChecker)
```

### **Custom Health Checkers**

```go
// Custom health check function
customChecker := metrics.NewCustomHealthChecker(
    "disk_space",
    func(ctx context.Context) (metrics.HealthStatus, string, map[string]string) {
        // Check disk space logic
        usedPercent := getDiskUsagePercent()

        metadata := map[string]string{
            "used_percent": fmt.Sprintf("%.2f", usedPercent),
            "mount_point":  "/",
        }

        if usedPercent > 90 {
            return metrics.HealthStatusUnhealthy,
                   fmt.Sprintf("Disk usage critical: %.2f%%", usedPercent),
                   metadata
        } else if usedPercent > 80 {
            return metrics.HealthStatusDegraded,
                   fmt.Sprintf("Disk usage high: %.2f%%", usedPercent),
                   metadata
        }

        return metrics.HealthStatusHealthy, "Disk space OK", metadata
    },
)
metrics.RegisterHealthCheck(customChecker)
```

### **Health Check Results**

```go
// Get all health checks
ctx := context.Background()
healthChecks, err := metrics.GetHealthChecks(ctx)
if err != nil {
    log.Printf("Error getting health checks: %v", err)
}

for _, check := range healthChecks {
    fmt.Printf("%s: %s (%v)\n", check.Name, check.Status, check.Duration)
    if check.Message != "" {
        fmt.Printf("  Message: %s\n", check.Message)
    }
    for key, value := range check.Metadata {
        fmt.Printf("  %s: %s\n", key, value)
    }
}

// Get overall health status
overallHealth, err := metrics.GetOverallHealth(ctx)
fmt.Printf("Overall system health: %s\n", overallHealth)
```

## 🛡️ Middleware

### **HTTP Metrics Middleware**

```go
// Setup router with metrics middleware
router := gin.Default()

// Add HTTP metrics middleware
router.Use(metrics.HTTPMetricsMiddleware())

// Your routes
router.GET("/api/users", getUsersHandler)
router.POST("/api/users", createUserHandler)

// This will automatically collect:
// - http_requests_total (counter)
// - http_request_duration_seconds (histogram)
// - http_requests_in_flight (gauge)
// - http_request_size_bytes (histogram)
// - http_response_size_bytes (histogram)
```

### **Database Metrics Middleware**

```go
router.Use(metrics.DatabaseMetricsMiddleware())

// In your database layer
func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    start := time.Now()

    // Database query
    user := &User{}
    err := r.db.Where("id = ?", id).First(user).Error

    // Record metrics (extracted from gin context if available)
    if ginCtx, ok := ctx.(*gin.Context); ok {
        metrics.RecordDatabaseQuery(ginCtx, "select", time.Since(start))
    }

    return user, err
}
```

### **Complete Middleware Setup**

```go
func SetupMetricsMiddleware(router *gin.Engine) {
    // Core metrics
    router.Use(metrics.HTTPMetricsMiddleware())
    router.Use(metrics.ErrorMetricsMiddleware())

    // Application-specific metrics
    router.Use(metrics.DatabaseMetricsMiddleware())
    router.Use(metrics.CacheMetricsMiddleware())
    router.Use(metrics.QueueMetricsMiddleware())
    router.Use(metrics.AuthMetricsMiddleware())
    router.Use(metrics.BusinessMetricsMiddleware())
}
```

## 🖥️ System Monitoring

### **System Metrics Collection**

```go
// Start collecting system metrics every 30 seconds
err := metrics.StartSystemMetrics(30 * time.Second)
if err != nil {
    log.Printf("Failed to start system metrics: %v", err)
}

// Get current system metrics
systemMetrics, err := metrics.GetSystemMetrics()
if err != nil {
    log.Printf("Error getting system metrics: %v", err)
} else {
    fmt.Printf("Memory usage: %.2f%%\n", systemMetrics.MemoryUsage)
    fmt.Printf("Goroutines: %d\n", systemMetrics.Goroutines)
    fmt.Printf("Memory total: %d bytes\n", systemMetrics.MemoryTotal)
}

// Get application metrics
appMetrics, err := metrics.GetApplicationMetrics()
if err != nil {
    log.Printf("Error getting app metrics: %v", err)
} else {
    fmt.Printf("HTTP requests: %d\n", appMetrics.HTTPRequestsTotal)
    fmt.Printf("Database connections: %d\n", appMetrics.DatabaseConnections)
    fmt.Printf("Cache hits: %d\n", appMetrics.CacheHits)
}
```

### **Monitor Configuration**

```go
// Configure metrics registry
config := &metrics.Config{
    Enabled:    true,
    Namespace:  "myapp",
    Subsystem:  "api",
    Labels: metrics.Labels{
        "version":     "1.0.0",
        "environment": "production",
    },
    Buckets: []float64{0.001, 0.01, 0.1, 1, 10}, // Custom histogram buckets
    FlushInterval: 15 * time.Second,
}
metrics.SetDefaultConfig(config)
```

## 💡 Examples

### **1. Web API Metrics**

```go
func SetupAPIMetrics(router *gin.Engine) {
    // Add metrics middleware
    router.Use(metrics.HTTPMetricsMiddleware())

    // Custom business metrics
    userRegistrations := metrics.NewCounter("user_registrations_total", "Total user registrations", metrics.Labels{})
    activeUsers := metrics.NewGauge("active_users_current", "Currently active users", metrics.Labels{})

    // Registration endpoint
    router.POST("/register", func(c *gin.Context) {
        var req RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": "Invalid request"})
            return
        }

        // Create user logic...
        user, err := createUser(req)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to create user"})
            return
        }

        // Record business metric
        userRegistrations.With(metrics.Labels{
            "user_type": user.Type,
            "source":    req.Source,
        }).Inc()

        c.JSON(201, gin.H{"user_id": user.ID})
    })

    // Update active users count periodically
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for range ticker.C {
            count := getActiveUserCount()
            activeUsers.Set(float64(count))
        }
    }()
}
```

### **2. Database Performance Monitoring**

```go
type MetricsUserRepository struct {
    db       *gorm.DB
    queries  metrics.Counter
    duration metrics.Histogram
}

func NewMetricsUserRepository(db *gorm.DB) *MetricsUserRepository {
    return &MetricsUserRepository{
        db: db,
        queries: metrics.NewCounter("db_queries_total", "Database queries", metrics.Labels{
            "table": "users",
        }),
        duration: metrics.NewHistogram("db_query_duration_seconds", "Query duration", nil, metrics.Labels{
            "table": "users",
        }),
    }
}

func (r *MetricsUserRepository) GetByID(id string) (*User, error) {
    start := time.Now()
    defer func() {
        r.duration.With(metrics.Labels{"operation": "select"}).ObserveDuration(start)
        r.queries.With(metrics.Labels{"operation": "select"}).Inc()
    }()

    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    return &user, err
}

func (r *MetricsUserRepository) Create(user *User) error {
    start := time.Now()
    defer func() {
        r.duration.With(metrics.Labels{"operation": "insert"}).ObserveDuration(start)
        r.queries.With(metrics.Labels{"operation": "insert"}).Inc()
    }()

    return r.db.Create(user).Error
}
```

### **3. Queue Monitoring**

```go
type MetricsJobQueue struct {
    queue       queue.Queue
    jobsQueued  metrics.Counter
    jobsProcessed metrics.Counter
    queueSize   metrics.Gauge
    processingTime metrics.Histogram
}

func NewMetricsJobQueue(q queue.Queue) *MetricsJobQueue {
    return &MetricsJobQueue{
        queue: q,
        jobsQueued: metrics.NewCounter("jobs_queued_total", "Jobs queued", metrics.Labels{}),
        jobsProcessed: metrics.NewCounter("jobs_processed_total", "Jobs processed", metrics.Labels{}),
        queueSize: metrics.NewGauge("queue_size_current", "Current queue size", metrics.Labels{}),
        processingTime: metrics.NewHistogram("job_processing_duration_seconds", "Job processing time", nil, metrics.Labels{}),
    }
}

func (mq *MetricsJobQueue) Push(job *queue.Job) error {
    err := mq.queue.Push(job)
    if err == nil {
        mq.jobsQueued.With(metrics.Labels{"type": job.Type}).Inc()
        mq.updateQueueSize()
    }
    return err
}

func (mq *MetricsJobQueue) ProcessJob(job *queue.Job) error {
    start := time.Now()

    err := mq.processJobLogic(job)

    // Record metrics
    duration := time.Since(start)
    mq.processingTime.With(metrics.Labels{
        "type":   job.Type,
        "status": getJobStatus(err),
    }).Observe(duration.Seconds())

    mq.jobsProcessed.With(metrics.Labels{
        "type":   job.Type,
        "status": getJobStatus(err),
    }).Inc()

    mq.updateQueueSize()
    return err
}

func (mq *MetricsJobQueue) updateQueueSize() {
    size, _ := mq.queue.GetQueueSize()
    mq.queueSize.Set(float64(size))
}
```

### **4. Cache Performance Tracking**

```go
type MetricsCache struct {
    cache     cache.Cache
    hits      metrics.Counter
    misses    metrics.Counter
    operations metrics.Histogram
}

func NewMetricsCache(c cache.Cache) *MetricsCache {
    return &MetricsCache{
        cache: c,
        hits: metrics.NewCounter("cache_hits_total", "Cache hits", metrics.Labels{}),
        misses: metrics.NewCounter("cache_misses_total", "Cache misses", metrics.Labels{}),
        operations: metrics.NewHistogram("cache_operation_duration_seconds", "Cache operation duration", nil, metrics.Labels{}),
    }
}

func (mc *MetricsCache) Get(key string) (string, error) {
    start := time.Now()
    defer func() {
        mc.operations.With(metrics.Labels{"operation": "get"}).ObserveDuration(start)
    }()

    value, err := mc.cache.Get(key)

    if err == nil && value != "" {
        mc.hits.With(metrics.Labels{"operation": "get"}).Inc()
    } else {
        mc.misses.With(metrics.Labels{"operation": "get"}).Inc()
    }

    return value, err
}

func (mc *MetricsCache) Set(key, value string, ttl time.Duration) error {
    start := time.Now()
    defer func() {
        mc.operations.With(metrics.Labels{"operation": "set"}).ObserveDuration(start)
    }()

    return mc.cache.Set(key, value, ttl)
}
```

### **5. Business Metrics Dashboard**

```go
type BusinessMetricsDashboard struct {
    // Revenue metrics
    dailyRevenue   metrics.Gauge
    monthlyRevenue metrics.Gauge

    // User metrics
    totalUsers     metrics.Gauge
    activeUsers    metrics.Gauge
    newUsers       metrics.Counter

    // Order metrics
    totalOrders    metrics.Counter
    orderValue     metrics.Histogram

    // Product metrics
    topProducts    metrics.Counter
}

func NewBusinessMetricsDashboard() *BusinessMetricsDashboard {
    return &BusinessMetricsDashboard{
        dailyRevenue: metrics.NewGauge("revenue_daily_total", "Daily revenue", metrics.Labels{}),
        monthlyRevenue: metrics.NewGauge("revenue_monthly_total", "Monthly revenue", metrics.Labels{}),
        totalUsers: metrics.NewGauge("users_total", "Total users", metrics.Labels{}),
        activeUsers: metrics.NewGauge("users_active", "Active users", metrics.Labels{}),
        newUsers: metrics.NewCounter("users_new_total", "New users", metrics.Labels{}),
        totalOrders: metrics.NewCounter("orders_total", "Total orders", metrics.Labels{}),
        orderValue: metrics.NewHistogram("order_value", "Order value distribution",
            []float64{10, 50, 100, 500, 1000, 5000}, metrics.Labels{}),
        topProducts: metrics.NewCounter("products_sold_total", "Products sold", metrics.Labels{}),
    }
}

func (bm *BusinessMetricsDashboard) RecordOrder(order *Order) {
    // Record order
    bm.totalOrders.With(metrics.Labels{
        "payment_method": order.PaymentMethod,
        "currency":       order.Currency,
    }).Inc()

    // Record order value
    bm.orderValue.With(metrics.Labels{
        "currency": order.Currency,
    }).Observe(order.TotalAmount)

    // Record products sold
    for _, item := range order.Items {
        bm.topProducts.With(metrics.Labels{
            "product_id":   item.ProductID,
            "category":     item.Category,
        }).Add(float64(item.Quantity))
    }

    // Update revenue
    bm.updateRevenue(order.TotalAmount)
}

func (bm *BusinessMetricsDashboard) RecordUserRegistration(user *User) {
    bm.newUsers.With(metrics.Labels{
        "source":    user.RegistrationSource,
        "user_type": user.Type,
    }).Inc()

    bm.updateUserCounts()
}

func (bm *BusinessMetricsDashboard) updateRevenue(amount float64) {
    // Update daily revenue
    today := time.Now().Format("2006-01-02")
    dailyRevenue := getDailyRevenue(today)
    bm.dailyRevenue.Set(dailyRevenue)

    // Update monthly revenue
    month := time.Now().Format("2006-01")
    monthlyRevenue := getMonthlyRevenue(month)
    bm.monthlyRevenue.Set(monthlyRevenue)
}

func (bm *BusinessMetricsDashboard) updateUserCounts() {
    total := getTotalUserCount()
    active := getActiveUserCount()

    bm.totalUsers.Set(float64(total))
    bm.activeUsers.Set(float64(active))
}
```

## 🎯 Best Practices

### **1. Metric Naming Conventions**

```go
// ✅ DO: Use descriptive, consistent names
httpRequestsTotal := metrics.NewCounter("http_requests_total", "Total HTTP requests", labels)
httpRequestDuration := metrics.NewHistogram("http_request_duration_seconds", "Request duration", nil, labels)

// ✅ DO: Include units in metric names
memorySizeBytes := metrics.NewGauge("memory_size_bytes", "Memory size in bytes", labels)
requestDurationSeconds := metrics.NewHistogram("request_duration_seconds", "Duration in seconds", nil, labels)

// ✅ DO: Use consistent label names
labels := metrics.Labels{
    "method":   "GET",
    "endpoint": "/api/users",
    "status":   "200",
}

// ❌ DON'T: Use inconsistent or unclear names
counter := metrics.NewCounter("cnt", "counter", labels) // Too short
requests := metrics.NewCounter("requests", "requests", labels) // No _total suffix
```

### **2. Label Management**

```go
// ✅ DO: Use meaningful labels
userActions := metrics.NewCounter("user_actions_total", "User actions", metrics.Labels{
    "action": "login",
    "source": "web",
    "result": "success",
})

// ✅ DO: Limit label cardinality
// Good: method, endpoint, status (limited values)
httpMetrics := metrics.NewCounter("http_requests_total", "HTTP requests", metrics.Labels{
    "method": c.Request.Method,     // GET, POST, PUT, DELETE (low cardinality)
    "status": strconv.Itoa(status), // 200, 404, 500, etc. (low cardinality)
})

// ❌ DON'T: Use high cardinality labels
badMetrics := metrics.NewCounter("requests_total", "Requests", metrics.Labels{
    "user_id":    userID,    // High cardinality!
    "request_id": requestID, // Very high cardinality!
    "timestamp":  timestamp, // Infinite cardinality!
})
```

### **3. Health Check Design**

```go
// ✅ DO: Implement timeout-aware health checks
func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthCheck {
    // Use context with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    start := time.Now()
    err := d.db.PingContext(timeoutCtx)
    duration := time.Since(start)

    // Provide meaningful status and metadata
    if err != nil {
        return HealthCheck{
            Name:      d.name,
            Status:    HealthStatusUnhealthy,
            Message:   fmt.Sprintf("Database ping failed: %v", err),
            Duration:  duration,
            Timestamp: time.Now(),
            Metadata: map[string]string{
                "error":         err.Error(),
                "response_time": duration.String(),
            },
        }
    }

    // Consider degraded status for slow responses
    status := HealthStatusHealthy
    message := "Database is healthy"

    if duration > 2*time.Second {
        status = HealthStatusDegraded
        message = fmt.Sprintf("Database response is slow: %v", duration)
    }

    return HealthCheck{
        Name:      d.name,
        Status:    status,
        Message:   message,
        Duration:  duration,
        Timestamp: time.Now(),
        Metadata: map[string]string{
            "response_time": duration.String(),
        },
    }
}
```

### **4. Performance Considerations**

```go
// ✅ DO: Use appropriate metric types
requestCounter := metrics.NewCounter("requests_total", "Total requests", labels)    // For counting
activeConnections := metrics.NewGauge("connections_active", "Active connections", labels) // For current values
responseTime := metrics.NewHistogram("response_time", "Response time", nil, labels) // For distributions

// ✅ DO: Batch metric updates when possible
func RecordBatchMetrics(orders []Order) {
    orderCounter := metrics.NewCounter("orders_total", "Total orders", metrics.Labels{})
    revenueGauge := metrics.NewGauge("revenue_total", "Total revenue", metrics.Labels{})

    var totalRevenue float64
    for _, order := range orders {
        orderCounter.With(metrics.Labels{
            "status": order.Status,
        }).Inc()
        totalRevenue += order.Amount
    }

    revenueGauge.Set(totalRevenue)
}

// ✅ DO: Use timers efficiently
func MeasureFunction(histogram metrics.Histogram) {
    timer := metrics.NewTimer(histogram)
    stop := timer.Start()
    defer stop() // Automatically records when function returns

    // Function logic here
}
```

### **5. Error Handling and Monitoring**

```go
// ✅ DO: Monitor metrics collection itself
func SafeRecordMetric(counter metrics.Counter, labels metrics.Labels) {
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Panic while recording metric",
                zap.Any("panic", r),
                zap.Any("labels", labels),
            )
        }
    }()

    counter.With(labels).Inc()
}

// ✅ DO: Provide metrics about metrics
metricsErrors := metrics.NewCounter("metrics_errors_total", "Metrics collection errors", metrics.Labels{})
metricsLatency := metrics.NewHistogram("metrics_collection_duration_seconds", "Time to collect metrics", nil, metrics.Labels{})

func CollectMetricsWithMonitoring() {
    start := time.Now()
    defer func() {
        metricsLatency.Observe(time.Since(start).Seconds())
    }()

    if err := collectSomeMetrics(); err != nil {
        metricsErrors.With(metrics.Labels{"type": "collection"}).Inc()
        logger.Error("Failed to collect metrics", zap.Error(err))
    }
}
```

### **6. Testing Metrics**

```go
func TestUserRegistrationMetrics(t *testing.T) {
    // Clear metrics before test
    metrics.Clear()

    // Create test metrics
    registrationCounter := metrics.NewCounter("test_registrations_total", "Test registrations", metrics.Labels{})

    // Record test data
    registrationCounter.With(metrics.Labels{"source": "web"}).Inc()
    registrationCounter.With(metrics.Labels{"source": "mobile"}).Inc()
    registrationCounter.With(metrics.Labels{"source": "web"}).Inc()

    // Verify metrics
    allMetrics := metrics.GetAllMetrics()

    webCount := 0
    mobileCount := 0

    for _, metric := range allMetrics {
        if metric.Name == "test_registrations_total" {
            if metric.Labels["source"] == "web" {
                webCount = int(metric.Value)
            } else if metric.Labels["source"] == "mobile" {
                mobileCount = int(metric.Value)
            }
        }
    }

    assert.Equal(t, 2, webCount)
    assert.Equal(t, 1, mobileCount)
}
```

## 🔗 Related Packages

- [`pkg/logger`](../logger/) - Logging integration
- [`pkg/cache`](../cache/) - Cache metrics
- [`pkg/queue`](../queue/) - Queue metrics
- [`config`](../../config/) - Metrics configuration

## 📚 Additional Resources

- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Monitoring Microservices](https://microservices.io/patterns/observability/application-metrics.html)
- [RED Method](https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/)
