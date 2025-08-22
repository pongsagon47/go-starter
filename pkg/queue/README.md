# 🔄 Queue Package

Redis-based background job queue system with priority support, delayed jobs, retry mechanisms, and concurrent workers for processing asynchronous tasks.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Queue Configuration](#queue-configuration)
- [Job Management](#job-management)
- [Workers](#workers)
- [Job Handlers](#job-handlers)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in flex-service
import "flex-service/pkg/queue"
```

## ⚡ Quick Start

### Basic Job Queue Setup

```go
package main

import (
    "context"
    "time"
    "flex-service/pkg/queue"
)

func main() {
    // Configure Redis queue
    config := &queue.RedisQueueConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        PoolSize: 10,
        Prefix:   "myapp",
    }

    // Create queue
    q, err := queue.NewRedisQueue("default", config)
    if err != nil {
        panic(err)
    }
    defer q.Close()

    // Create worker
    worker := queue.NewRedisWorker(q, &queue.WorkerConfig{
        NumWorkers: 4,
        PollTime:   time.Second,
    })

    // Register job handlers
    worker.RegisterHandler("email", queue.EmailJobHandler(nil))
    worker.RegisterHandler("webhook", queue.WebhookJobHandler())

    // Start worker
    ctx := context.Background()
    go worker.Start(ctx)

    // Dispatch jobs
    dispatcher := queue.NewJobDispatcher(q)

    // Send email job
    dispatcher.Dispatch("email", map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Welcome!",
        "body":    "Welcome to our service!",
    })

    // Send delayed webhook job
    dispatcher.DispatchDelayed("webhook", map[string]interface{}{
        "url":    "https://api.example.com/webhook",
        "method": "POST",
        "data":   map[string]string{"event": "user_registered"},
    }, 5*time.Minute)
}
```

## ⚙️ Queue Configuration

### **Redis Queue Configuration**

```go
type RedisQueueConfig struct {
    Addr         string        // Redis address (localhost:6379)
    Password     string        // Redis password
    DB           int           // Redis database number
    PoolSize     int           // Connection pool size
    MinIdleConns int           // Minimum idle connections
    MaxRetries   int           // Maximum retry attempts
    RetryDelays  []time.Duration // Retry delay intervals
    Prefix       string        // Key prefix for Redis keys
}

// Example configuration
config := &queue.RedisQueueConfig{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     20,
    MinIdleConns: 5,
    MaxRetries:   3,
    RetryDelays: []time.Duration{
        1 * time.Second,
        5 * time.Second,
        30 * time.Second,
        5 * time.Minute,
    },
    Prefix: "myapp_queue",
}
```

### **Worker Configuration**

```go
type WorkerConfig struct {
    NumWorkers int           // Number of concurrent workers
    PollTime   time.Duration // How often to poll for jobs
    Logger     *zap.Logger   // Logger instance
}

// Example configuration
workerConfig := &queue.WorkerConfig{
    NumWorkers: 8,           // 8 concurrent workers
    PollTime:   time.Second, // Poll every second
    Logger:     logger.Logger,
}
```

## 📋 Job Management

### **Job Structure**

```go
type Job struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Payload     map[string]interface{} `json:"payload"`
    Attempts    int                    `json:"attempts"`
    MaxAttempts int                    `json:"max_attempts"`
    Priority    int                    `json:"priority"`      // Higher = higher priority
    Delay       time.Duration          `json:"delay"`         // Delay before processing
    CreatedAt   time.Time              `json:"created_at"`
    ProcessedAt *time.Time             `json:"processed_at,omitempty"`
    FailedAt    *time.Time             `json:"failed_at,omitempty"`
    Error       string                 `json:"error,omitempty"`
}
```

### **Job Statuses**

```go
const (
    StatusPending    = "pending"    // Waiting to be processed
    StatusProcessing = "processing" // Currently being processed
    StatusCompleted  = "completed"  // Successfully completed
    StatusFailed     = "failed"     // Permanently failed
    StatusRetrying   = "retrying"   // Waiting for retry
    StatusCancelled  = "cancelled"  // Cancelled by user
)
```

### **Creating Jobs**

```go
// Basic job creation
job := &queue.Job{
    Type: "email",
    Payload: map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Welcome",
        "body":    "Welcome to our service!",
    },
    MaxAttempts: 3,
    Priority:    5,
}

// Push to queue
err := q.Push(job)

// Push with delay
err := q.PushDelayed(job, 10*time.Minute)
```

### **Job Options**

```go
type JobOptions struct {
    MaxAttempts int           `json:"max_attempts"`
    Priority    int           `json:"priority"`
    Delay       time.Duration `json:"delay"`
    Queue       string        `json:"queue"`
}

// Using job options
options := &queue.JobOptions{
    MaxAttempts: 5,
    Priority:    10,
    Delay:       time.Hour,
}

dispatcher.Dispatch("important_task", payload, options)
```

### **Queue Operations**

```go
// Get next job
job, err := q.Pop()

// Get job by ID
job, err := q.GetJob("job_abc123")

// Get job status
status, err := q.GetJobStatus("job_abc123")

// Cancel job
err := q.CancelJob("job_abc123")

// Get queue size
size, err := q.GetQueueSize()

// Get queue statistics
stats, err := q.GetStats()
```

## 👥 Workers

### **Worker Lifecycle**

```go
// Create worker
worker := queue.NewRedisWorker(q, &queue.WorkerConfig{
    NumWorkers: 4,
    PollTime:   time.Second,
})

// Register handlers
worker.RegisterHandler("email", emailHandler)
worker.RegisterHandler("webhook", webhookHandler)

// Start worker
ctx := context.Background()
go func() {
    if err := worker.Start(ctx); err != nil {
        log.Fatal("Failed to start worker:", err)
    }
}()

// Check if running
if worker.IsRunning() {
    fmt.Println("Worker is running")
}

// Graceful shutdown
if err := worker.Stop(); err != nil {
    log.Printf("Error stopping worker: %v", err)
}
```

### **Multiple Workers**

```go
// Create multiple workers for different queues
emailQueue, _ := queue.NewRedisQueue("email", config)
webhookQueue, _ := queue.NewRedisQueue("webhook", config)

emailWorker := queue.NewRedisWorker(emailQueue, &queue.WorkerConfig{
    NumWorkers: 2,
})

webhookWorker := queue.NewRedisWorker(webhookQueue, &queue.WorkerConfig{
    NumWorkers: 4,
})

// Start both workers
go emailWorker.Start(ctx)
go webhookWorker.Start(ctx)
```

## 🔧 Job Handlers

### **Handler Interface**

```go
type Handler interface {
    Handle(ctx context.Context, job *Job) *JobResult
}

type JobResult struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### **Function Handlers**

```go
// Using HandlerFunc
emailHandler := queue.HandlerFunc(func(ctx context.Context, job *queue.Job) *queue.JobResult {
    to := job.Payload["to"].(string)
    subject := job.Payload["subject"].(string)
    body := job.Payload["body"].(string)

    // Send email logic here
    err := sendEmail(to, subject, body)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   err.Error(),
        }
    }

    return &queue.JobResult{
        Success: true,
        Data: map[string]interface{}{
            "sent_to": to,
            "sent_at": time.Now(),
        },
    }
})
```

### **Struct Handlers**

```go
type EmailService struct {
    mailer *mail.Mailer
}

func (es *EmailService) Handle(ctx context.Context, job *queue.Job) *queue.JobResult {
    to := job.Payload["to"].(string)
    subject := job.Payload["subject"].(string)
    body := job.Payload["body"].(string)

    err := es.mailer.SendEmail([]string{to}, subject, body, nil)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   err.Error(),
        }
    }

    return &queue.JobResult{
        Success: true,
        Data: map[string]interface{}{
            "email_sent": true,
            "timestamp":  time.Now(),
        },
    }
}

// Register struct handler
emailService := &EmailService{mailer: mailer}
worker.RegisterHandler("email", emailService)
```

## 💡 Examples

### **1. Email Notification System**

```go
type EmailNotificationHandler struct {
    mailer *mail.Mailer
    userService *UserService
}

func (enh *EmailNotificationHandler) Handle(ctx context.Context, job *queue.Job) *queue.JobResult {
    userID := job.Payload["user_id"].(string)
    template := job.Payload["template"].(string)
    data := job.Payload["data"].(map[string]interface{})

    // Get user
    user, err := enh.userService.GetByID(userID)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("User not found: %v", err),
        }
    }

    // Prepare email data
    emailData := map[string]interface{}{
        "User": user,
        "Data": data,
    }

    // Send templated email
    err = enh.mailer.SendTemplate(
        []string{user.Email},
        getEmailSubject(template),
        template,
        emailData,
        nil,
    )

    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   err.Error(),
        }
    }

    return &queue.JobResult{
        Success: true,
        Data: map[string]interface{}{
            "user_id":   userID,
            "template":  template,
            "sent_to":   user.Email,
            "sent_at":   time.Now(),
        },
    }
}

// Usage
func SendWelcomeEmail(dispatcher *queue.JobDispatcher, userID string) error {
    return dispatcher.Dispatch("email_notification", map[string]interface{}{
        "user_id":  userID,
        "template": "welcome",
        "data": map[string]interface{}{
            "welcome_bonus": 100,
        },
    })
}
```

### **2. Image Processing Pipeline**

```go
type ImageProcessingHandler struct {
    storage *FileStorage
}

func (iph *ImageProcessingHandler) Handle(ctx context.Context, job *queue.Job) *queue.JobResult {
    imagePath := job.Payload["image_path"].(string)
    operations := job.Payload["operations"].([]interface{})

    logger.Info("Processing image",
        zap.String("job_id", job.ID),
        zap.String("image_path", imagePath),
        zap.Any("operations", operations),
    )

    // Load image
    img, err := iph.storage.LoadImage(imagePath)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("Failed to load image: %v", err),
        }
    }

    // Process operations
    for _, op := range operations {
        operation := op.(map[string]interface{})
        opType := operation["type"].(string)

        switch opType {
        case "resize":
            width := int(operation["width"].(float64))
            height := int(operation["height"].(float64))
            img = resizeImage(img, width, height)

        case "crop":
            x := int(operation["x"].(float64))
            y := int(operation["y"].(float64))
            width := int(operation["width"].(float64))
            height := int(operation["height"].(float64))
            img = cropImage(img, x, y, width, height)

        case "filter":
            filterType := operation["filter"].(string)
            img = applyFilter(img, filterType)
        }
    }

    // Save processed image
    outputPath := generateOutputPath(imagePath)
    err = iph.storage.SaveImage(img, outputPath)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("Failed to save image: %v", err),
        }
    }

    return &queue.JobResult{
        Success: true,
        Data: map[string]interface{}{
            "original_path": imagePath,
            "output_path":   outputPath,
            "operations":    operations,
            "processed_at":  time.Now(),
        },
    }
}

// Usage
func ProcessUserAvatar(dispatcher *queue.JobDispatcher, imagePath string) error {
    return dispatcher.Dispatch("image_processing", map[string]interface{}{
        "image_path": imagePath,
        "operations": []map[string]interface{}{
            {
                "type":   "resize",
                "width":  400,
                "height": 400,
            },
            {
                "type":   "crop",
                "x":      50,
                "y":      50,
                "width":  300,
                "height": 300,
            },
        },
    }, &queue.JobOptions{
        Priority:    5,
        MaxAttempts: 2,
    })
}
```

### **3. Report Generation**

```go
type ReportGenerationHandler struct {
    reportService *ReportService
    storage       *FileStorage
    mailer        *mail.Mailer
}

func (rgh *ReportGenerationHandler) Handle(ctx context.Context, job *queue.Job) *queue.JobResult {
    reportType := job.Payload["report_type"].(string)
    userID := job.Payload["user_id"].(string)
    params := job.Payload["parameters"].(map[string]interface{})

    logger.Info("Generating report",
        zap.String("job_id", job.ID),
        zap.String("report_type", reportType),
        zap.String("user_id", userID),
    )

    // Generate report
    report, err := rgh.reportService.GenerateReport(reportType, params)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("Failed to generate report: %v", err),
        }
    }

    // Save to storage
    filename := fmt.Sprintf("report_%s_%s.pdf", reportType, time.Now().Format("20060102_150405"))
    filePath, err := rgh.storage.SaveFile(filename, report.Data)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("Failed to save report: %v", err),
        }
    }

    // Send email notification
    user, _ := getUserByID(userID)
    if user != nil {
        rgh.mailer.SendTemplate(
            []string{user.Email},
            "Your Report is Ready",
            "report_ready",
            map[string]interface{}{
                "User":       user,
                "ReportType": reportType,
                "DownloadURL": generateDownloadURL(filePath),
            },
            []string{filePath}, // Attach the report
        )
    }

    return &queue.JobResult{
        Success: true,
        Data: map[string]interface{}{
            "report_type":   reportType,
            "file_path":     filePath,
            "generated_at":  time.Now(),
            "user_notified": user != nil,
        },
    }
}

// Usage
func GenerateMonthlyReport(dispatcher *queue.JobDispatcher, userID string, month time.Month, year int) error {
    return dispatcher.DispatchDelayed("report_generation", map[string]interface{}{
        "report_type": "monthly_sales",
        "user_id":     userID,
        "parameters": map[string]interface{}{
            "month": month,
            "year":  year,
        },
    }, 5*time.Minute, &queue.JobOptions{ // Delay to ensure all data is ready
        Priority:    3,
        MaxAttempts: 2,
    })
}
```

### **4. Webhook Delivery System**

```go
type WebhookDeliveryHandler struct {
    httpClient *http.Client
}

func (wdh *WebhookDeliveryHandler) Handle(ctx context.Context, job *queue.Job) *queue.JobResult {
    url := job.Payload["url"].(string)
    method := job.Payload["method"].(string)
    headers := job.Payload["headers"].(map[string]interface{})
    payload := job.Payload["payload"]

    logger.Info("Delivering webhook",
        zap.String("job_id", job.ID),
        zap.String("url", url),
        zap.String("method", method),
    )

    // Prepare request
    var reqBody io.Reader
    if payload != nil {
        jsonData, err := json.Marshal(payload)
        if err != nil {
            return &queue.JobResult{
                Success: false,
                Error:   fmt.Sprintf("Failed to marshal payload: %v", err),
            }
        }
        reqBody = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("Failed to create request: %v", err),
        }
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    for key, value := range headers {
        req.Header.Set(key, fmt.Sprintf("%v", value))
    }

    // Make request
    resp, err := wdh.httpClient.Do(req)
    if err != nil {
        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("Failed to deliver webhook: %v", err),
        }
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return &queue.JobResult{
            Success: true,
            Data: map[string]interface{}{
                "url":           url,
                "status_code":   resp.StatusCode,
                "delivered_at":  time.Now(),
            },
        }
    }

    // Read error response
    body, _ := io.ReadAll(resp.Body)
    return &queue.JobResult{
        Success: false,
        Error:   fmt.Sprintf("Webhook delivery failed with status %d: %s", resp.StatusCode, string(body)),
    }
}

// Usage
func SendUserRegisteredWebhook(dispatcher *queue.JobDispatcher, user *User) error {
    return dispatcher.Dispatch("webhook_delivery", map[string]interface{}{
        "url":    "https://api.partner.com/webhooks/user_registered",
        "method": "POST",
        "headers": map[string]interface{}{
            "Authorization": "Bearer " + getWebhookToken(),
            "X-Event-Type":  "user.registered",
        },
        "payload": map[string]interface{}{
            "event":      "user.registered",
            "user_id":    user.ID,
            "user_email": user.Email,
            "timestamp":  time.Now(),
        },
    }, &queue.JobOptions{
        MaxAttempts: 5,
        Priority:    2,
    })
}
```

## 🎯 Best Practices

### **1. Job Design**

```go
// ✅ DO: Keep jobs idempotent
func ProcessPaymentHandler(ctx context.Context, job *queue.Job) *queue.JobResult {
    paymentID := job.Payload["payment_id"].(string)

    // Check if already processed
    payment, err := getPayment(paymentID)
    if err != nil {
        return &queue.JobResult{Success: false, Error: err.Error()}
    }

    if payment.Status == "completed" {
        return &queue.JobResult{Success: true, Data: payment} // Already processed
    }

    // Process payment...
}

// ✅ DO: Include necessary context in payload
dispatcher.Dispatch("process_order", map[string]interface{}{
    "order_id":    order.ID,
    "user_id":     order.UserID,
    "amount":      order.Amount,
    "currency":    order.Currency,
    "created_at":  order.CreatedAt,
})

// ❌ DON'T: Store large objects in payload
// Instead, store IDs and fetch the data in the handler
```

### **2. Error Handling**

```go
// ✅ DO: Handle different types of errors appropriately
func EmailHandler(ctx context.Context, job *queue.Job) *queue.JobResult {
    to := job.Payload["to"].(string)

    // Validate input
    if to == "" {
        return &queue.JobResult{
            Success: false,
            Error:   "recipient email is required", // Don't retry
        }
    }

    // Attempt to send
    err := sendEmail(to, subject, body)
    if err != nil {
        if isTemporaryError(err) {
            return &queue.JobResult{
                Success: false,
                Error:   fmt.Sprintf("temporary error: %v", err), // Will retry
            }
        }

        return &queue.JobResult{
            Success: false,
            Error:   fmt.Sprintf("permanent error: %v", err), // Won't retry if max attempts reached
        }
    }

    return &queue.JobResult{Success: true}
}
```

### **3. Monitoring and Logging**

```go
// ✅ DO: Log important events
func MonitoredHandler(ctx context.Context, job *queue.Job) *queue.JobResult {
    logger.Info("Job started",
        zap.String("job_id", job.ID),
        zap.String("job_type", job.Type),
        zap.Int("attempt", job.Attempts),
    )

    start := time.Now()
    result := actualHandler(ctx, job)
    duration := time.Since(start)

    if result.Success {
        logger.Info("Job completed successfully",
            zap.String("job_id", job.ID),
            zap.Duration("duration", duration),
        )
    } else {
        logger.Error("Job failed",
            zap.String("job_id", job.ID),
            zap.String("error", result.Error),
            zap.Duration("duration", duration),
        )
    }

    return result
}
```

### **4. Worker Management**

```go
// ✅ DO: Implement graceful shutdown
func main() {
    worker := setupWorker()

    // Handle shutdown signals
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start worker
    go worker.Start(ctx)

    // Wait for shutdown signal
    <-c
    logger.Info("Shutting down worker...")

    // Stop worker gracefully
    if err := worker.Stop(); err != nil {
        logger.Error("Error stopping worker", zap.Error(err))
    }
}
```

### **5. Queue Monitoring**

```go
// ✅ DO: Monitor queue health
func MonitorQueue(q queue.Queue) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        stats, err := q.GetStats()
        if err != nil {
            logger.Error("Failed to get queue stats", zap.Error(err))
            continue
        }

        logger.Info("Queue stats",
            zap.Int64("pending", stats.PendingJobs),
            zap.Int64("processing", stats.ProcessingJobs),
            zap.Int64("failed", stats.FailedJobs),
        )

        // Alert if too many failed jobs
        if stats.FailedJobs > 100 {
            sendAlert("High number of failed jobs: %d", stats.FailedJobs)
        }

        // Alert if queue is backing up
        if stats.PendingJobs > 1000 {
            sendAlert("Queue backing up: %d pending jobs", stats.PendingJobs)
        }
    }
}
```

## 🔗 Related Packages

- [`pkg/cache`](../cache/) - Redis connection sharing
- [`pkg/logger`](../logger/) - Job logging
- [`pkg/mail`](../mail/) - Email job handlers
- [`config`](../../config/) - Queue configuration

## 📚 Additional Resources

- [Redis Documentation](https://redis.io/documentation)
- [Background Job Best Practices](https://github.com/collectiveidea/delayed_job#best-practices)
- [Queue Design Patterns](https://www.enterpriseintegrationpatterns.com/patterns/messaging/)
