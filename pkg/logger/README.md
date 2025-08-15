# 📊 Logger Package

Structured logging system built on Uber's Zap library with configurable levels, formats, and output destinations for production-ready applications.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Logging Levels](#logging-levels)
- [Structured Logging](#structured-logging)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in go-starter
import "go-starter/pkg/logger"
```

## ⚡ Quick Start

### Basic Logging

```go
package main

import (
    "go-starter/pkg/logger"
    "go.uber.org/zap"
)

func main() {
    // Initialize logger
    err := logger.Init("info", "json")
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Basic logging
    logger.Info("Application started")
    logger.Debug("Debug information")
    logger.Warn("Warning message")
    logger.Error("Error occurred")

    // Structured logging with fields
    logger.Info("User created",
        zap.String("user_id", "123"),
        zap.String("email", "user@example.com"),
        zap.Int("age", 25),
    )
}
```

## ⚙️ Configuration

### **Initialize Logger**

```go
func Init(level, format string) error
```

**Parameters:**

- `level`: Log level (`debug`, `info`, `warn`, `error`)
- `format`: Output format (`json` for production, `console` for development)

### **Environment-Based Configuration**

```go
func initLogger() {
    var level, format string

    switch os.Getenv("APP_ENV") {
    case "production":
        level = "info"
        format = "json"
    case "staging":
        level = "debug"
        format = "json"
    default: // development
        level = "debug"
        format = "console"
    }

    if err := logger.Init(level, format); err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }
}
```

### **Configuration Examples**

```env
# Environment variables
LOG_LEVEL=info
LOG_FORMAT=json
```

```go
// In config/config.go
type LogConfig struct {
    Level  string `env:"LOG_LEVEL" envDefault:"info"`
    Format string `env:"LOG_FORMAT" envDefault:"console"`
}

func (cfg *Config) InitLogger() error {
    return logger.Init(cfg.Log.Level, cfg.Log.Format)
}
```

## 📊 Logging Levels

### **Level Hierarchy**

| Level   | Use Case              | Example                              |
| ------- | --------------------- | ------------------------------------ |
| `debug` | Development debugging | Variable values, function calls      |
| `info`  | General information   | Application events, user actions     |
| `warn`  | Warning conditions    | Deprecated usage, recoverable errors |
| `error` | Error conditions      | Failed operations, exceptions        |

### **Level Examples**

```go
// DEBUG: Detailed information for diagnosing problems
logger.Debug("Processing user request",
    zap.String("method", "POST"),
    zap.String("path", "/api/users"),
    zap.Any("payload", requestData),
)

// INFO: General operational information
logger.Info("User logged in successfully",
    zap.String("user_id", userID),
    zap.String("ip_address", clientIP),
    zap.Duration("duration", time.Since(start)),
)

// WARN: Something unexpected happened, but application can continue
logger.Warn("Using fallback email service",
    zap.Error(primaryServiceError),
    zap.String("fallback_service", "sendgrid"),
)

// ERROR: Serious problem that needs attention
logger.Error("Failed to process payment",
    zap.String("transaction_id", txnID),
    zap.Error(err),
    zap.Float64("amount", amount),
)
```

## 🏗️ Structured Logging

### **Common Field Types**

```go
import "go.uber.org/zap"

// String fields
logger.Info("Event occurred",
    zap.String("user_id", "12345"),
    zap.String("action", "login"),
    zap.String("source", "mobile_app"),
)

// Numeric fields
logger.Info("API request processed",
    zap.Int("status_code", 200),
    zap.Float64("response_time_ms", 123.45),
    zap.Int64("bytes_sent", 2048),
)

// Time and Duration
logger.Info("Database query completed",
    zap.Time("started_at", startTime),
    zap.Duration("duration", time.Since(startTime)),
)

// Boolean fields
logger.Info("User preferences updated",
    zap.Bool("email_notifications", true),
    zap.Bool("sms_notifications", false),
)

// Complex objects
logger.Info("Order created",
    zap.Any("order", order),
    zap.Strings("product_ids", productIDs),
)

// Error fields
logger.Error("Database connection failed",
    zap.Error(err),
    zap.String("database", "users"),
    zap.String("operation", "select"),
)
```

## 💡 Examples

### **1. HTTP Request Logging**

```go
func LoggingMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        logger.Info("HTTP request",
            zap.String("method", param.Method),
            zap.String("path", param.Path),
            zap.Int("status", param.StatusCode),
            zap.Duration("latency", param.Latency),
            zap.String("client_ip", param.ClientIP),
            zap.String("user_agent", param.Request.UserAgent()),
            zap.Int("body_size", param.BodySize),
        )
        return ""
    })
}

// Usage in router
func SetupRouter() *gin.Engine {
    router := gin.New()
    router.Use(LoggingMiddleware())
    router.Use(gin.Recovery())

    return router
}
```

### **2. Database Operation Logging**

```go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) Create(user *User) error {
    start := time.Now()

    logger.Debug("Creating user",
        zap.String("email", user.Email),
        zap.String("name", user.Name),
    )

    err := r.db.Create(user).Error
    if err != nil {
        logger.Error("Failed to create user",
            zap.Error(err),
            zap.String("email", user.Email),
            zap.Duration("duration", time.Since(start)),
        )
        return err
    }

    logger.Info("User created successfully",
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
        zap.Duration("duration", time.Since(start)),
    )

    return nil
}

func (r *UserRepository) GetByID(id string) (*User, error) {
    start := time.Now()

    var user User
    err := r.db.Where("id = ?", id).First(&user).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            logger.Warn("User not found",
                zap.String("user_id", id),
                zap.Duration("query_time", time.Since(start)),
            )
        } else {
            logger.Error("Database query failed",
                zap.Error(err),
                zap.String("user_id", id),
                zap.String("operation", "select"),
                zap.Duration("query_time", time.Since(start)),
            )
        }
        return nil, err
    }

    logger.Debug("User retrieved",
        zap.String("user_id", id),
        zap.Duration("query_time", time.Since(start)),
    )

    return &user, nil
}
```

### **3. Service Layer Logging**

```go
type EmailService struct {
    mailer *mail.Mailer
}

func (s *EmailService) SendWelcomeEmail(user *User) error {
    start := time.Now()

    logger.Info("Sending welcome email",
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
        zap.String("template", "welcome"),
    )

    err := s.mailer.SendTemplate(
        []string{user.Email},
        "Welcome to Our Platform!",
        "welcome",
        map[string]interface{}{
            "Name": user.Name,
            "Email": user.Email,
        },
        nil,
    )

    if err != nil {
        logger.Error("Failed to send welcome email",
            zap.Error(err),
            zap.String("user_id", user.ID),
            zap.String("email", user.Email),
            zap.Duration("duration", time.Since(start)),
        )
        return err
    }

    logger.Info("Welcome email sent successfully",
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
        zap.Duration("duration", time.Since(start)),
    )

    return nil
}
```

### **4. Background Job Logging**

```go
func ProcessPaymentJob(paymentID string) {
    logger.Info("Starting payment processing job",
        zap.String("payment_id", paymentID),
        zap.String("job_type", "payment_processing"),
    )

    start := time.Now()

    defer func() {
        if r := recover(); r != nil {
            logger.Error("Payment processing job panicked",
                zap.String("payment_id", paymentID),
                zap.Any("panic", r),
                zap.Duration("duration", time.Since(start)),
            )
        }
    }()

    // Process payment
    payment, err := getPayment(paymentID)
    if err != nil {
        logger.Error("Failed to retrieve payment",
            zap.Error(err),
            zap.String("payment_id", paymentID),
        )
        return
    }

    logger.Debug("Payment retrieved for processing",
        zap.String("payment_id", paymentID),
        zap.Float64("amount", payment.Amount),
        zap.String("currency", payment.Currency),
        zap.String("status", payment.Status),
    )

    // Process with external service
    result, err := processWithGateway(payment)
    if err != nil {
        logger.Error("Payment gateway processing failed",
            zap.Error(err),
            zap.String("payment_id", paymentID),
            zap.String("gateway", "stripe"),
            zap.Duration("duration", time.Since(start)),
        )
        return
    }

    logger.Info("Payment processed successfully",
        zap.String("payment_id", paymentID),
        zap.String("transaction_id", result.TransactionID),
        zap.String("status", result.Status),
        zap.Duration("duration", time.Since(start)),
    )
}
```

### **5. Error Tracking and Monitoring**

```go
func HandleError(err error, ctx context.Context) {
    if err == nil {
        return
    }

    // Extract context information
    userID := getUserIDFromContext(ctx)
    requestID := getRequestIDFromContext(ctx)

    if appErr, ok := err.(*errors.AppError); ok {
        // Log application errors with full context
        logger.Error("Application error occurred",
            zap.String("error_code", appErr.Code),
            zap.String("error_message", appErr.Message),
            zap.Int("status_code", appErr.StatusCode),
            zap.Any("error_details", appErr.Details),
            zap.String("user_id", userID),
            zap.String("request_id", requestID),
            zap.Error(appErr.Cause),
        )
    } else {
        // Log unexpected errors
        logger.Error("Unexpected error occurred",
            zap.Error(err),
            zap.String("error_type", fmt.Sprintf("%T", err)),
            zap.String("user_id", userID),
            zap.String("request_id", requestID),
        )
    }
}
```

## 🎯 Best Practices

### **1. Consistent Field Names**

```go
// Use consistent field names across your application
const (
    FieldUserID      = "user_id"
    FieldRequestID   = "request_id"
    FieldTraceID     = "trace_id"
    FieldOperationID = "operation_id"
    FieldDuration    = "duration"
    FieldStatusCode  = "status_code"
    FieldMethod      = "method"
    FieldPath        = "path"
    FieldErrorCode   = "error_code"
)

// Usage
logger.Info("User action completed",
    zap.String(FieldUserID, userID),
    zap.String(FieldRequestID, requestID),
    zap.String("action", "profile_update"),
    zap.Duration(FieldDuration, time.Since(start)),
)
```

### **2. Context-Aware Logging**

```go
// Create logger helpers that extract context
func LogWithContext(ctx context.Context) *zap.Logger {
    fields := []zap.Field{}

    if userID := getUserIDFromContext(ctx); userID != "" {
        fields = append(fields, zap.String(FieldUserID, userID))
    }

    if requestID := getRequestIDFromContext(ctx); requestID != "" {
        fields = append(fields, zap.String(FieldRequestID, requestID))
    }

    if traceID := getTraceIDFromContext(ctx); traceID != "" {
        fields = append(fields, zap.String(FieldTraceID, traceID))
    }

    return logger.Logger.With(fields...)
}

// Usage
func CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    log := LogWithContext(ctx)

    log.Info("Creating user",
        zap.String("email", req.Email),
        zap.String("name", req.Name),
    )

    // ... create user logic

    log.Info("User created successfully",
        zap.String("user_id", user.ID),
    )

    return user, nil
}
```

### **3. Performance Monitoring**

```go
func LogPerformance(operation string, fn func() error) error {
    start := time.Now()

    logger.Debug("Starting operation",
        zap.String("operation", operation),
        zap.Time("started_at", start),
    )

    err := fn()
    duration := time.Since(start)

    if err != nil {
        logger.Error("Operation failed",
            zap.String("operation", operation),
            zap.Duration("duration", duration),
            zap.Error(err),
        )
    } else {
        logger.Info("Operation completed",
            zap.String("operation", operation),
            zap.Duration("duration", duration),
        )
    }

    // Log performance metrics
    if duration > time.Second {
        logger.Warn("Slow operation detected",
            zap.String("operation", operation),
            zap.Duration("duration", duration),
            zap.Duration("threshold", time.Second),
        )
    }

    return err
}

// Usage
err := LogPerformance("user_creation", func() error {
    return userService.CreateUser(req)
})
```

### **4. Security and PII Handling**

```go
// Create safe logging functions for sensitive data
func SafeLogUser(user *User) []zap.Field {
    return []zap.Field{
        zap.String("user_id", user.ID),
        zap.String("email_domain", getEmailDomain(user.Email)),
        zap.Bool("email_verified", user.EmailVerified),
        zap.Time("created_at", user.CreatedAt),
        // Don't log: email, password, phone, etc.
    }
}

func SafeLogPayment(payment *Payment) []zap.Field {
    return []zap.Field{
        zap.String("payment_id", payment.ID),
        zap.Float64("amount", payment.Amount),
        zap.String("currency", payment.Currency),
        zap.String("card_last_four", payment.CardLastFour),
        zap.String("status", payment.Status),
        // Don't log: card number, CVV, etc.
    }
}

// Usage
logger.Info("Payment processed", SafeLogPayment(payment)...)
```

### **5. Testing with Logs**

```go
func TestUserService_CreateUser(t *testing.T) {
    // Initialize test logger
    err := logger.Init("debug", "console")
    require.NoError(t, err)
    defer logger.Sync()

    userService := NewUserService(mockRepo)

    req := CreateUserRequest{
        Name:     "Test User",
        Email:    "test@example.com",
        Password: "password123",
    }

    user, err := userService.CreateUser(req)

    assert.NoError(t, err)
    assert.NotNil(t, user)

    // Logs will show the test execution flow
}

// For production tests, use a test logger that captures logs
func setupTestLogger() (*zap.Logger, *observer.ObservedLogs) {
    core, recorded := observer.New(zapcore.InfoLevel)
    testLogger := zap.New(core)

    // Replace global logger for testing
    originalLogger := logger.Logger
    logger.Logger = testLogger

    return testLogger, recorded
}
```

### **6. Log Rotation and Management**

```go
// In production, configure log rotation
func InitProductionLogger() error {
    config := zap.NewProductionConfig()

    // Configure output paths
    config.OutputPaths = []string{
        "stdout",
        "/var/log/app/app.log",
    }

    config.ErrorOutputPaths = []string{
        "stderr",
        "/var/log/app/error.log",
    }

    // Build logger
    var err error
    logger.Logger, err = config.Build()
    return err
}
```

## 🔄 Integration Examples

### **With HTTP Middleware**

```go
func RequestLoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        requestID := uuid.New().String()

        // Add request ID to context
        c.Set("request_id", requestID)

        logger.Info("Request started",
            zap.String("request_id", requestID),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.String("query", c.Request.URL.RawQuery),
            zap.String("user_agent", c.Request.UserAgent()),
            zap.String("client_ip", c.ClientIP()),
        )

        c.Next()

        logger.Info("Request completed",
            zap.String("request_id", requestID),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", time.Since(start)),
            zap.Int("response_size", c.Writer.Size()),
        )
    }
}
```

### **With Error Handlers**

```go
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            requestID := c.GetString("request_id")

            logger.Error("Request error",
                zap.String("request_id", requestID),
                zap.String("method", c.Request.Method),
                zap.String("path", c.Request.URL.Path),
                zap.Error(err),
            )
        }
    }
}
```

## 🔗 Related Packages

- [`pkg/errors`](../errors/) - Error handling
- [`config`](../../config/) - Logger configuration
- [`internal/middleware`](../../internal/middleware/) - HTTP request logging

## 📚 Additional Resources

- [Uber Zap Documentation](https://pkg.go.dev/go.uber.org/zap)
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team/)
- [Go Logging Guidelines](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)
