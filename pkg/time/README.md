# ⏰ Time Package

Timezone-aware time utilities for consistent time handling across different environments and locales with configurable timezone support.

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Core Functions](#core-functions)
- [Timezone Handling](#timezone-handling)
- [Examples](#examples)
- [Best Practices](#best-practices)

## 🚀 Installation

```bash
# Already included in flex-service
import "flex-service/pkg/time"
```

## ⚡ Quick Start

### Basic Time Operations

```go
package main

import (
    "fmt"
    apptime "flex-service/pkg/time"
    "time"
)

func main() {
    // Initialize with timezone
    err := apptime.InitTimezone("America/New_York")
    if err != nil {
        panic(err)
    }

    // Get current time in configured timezone
    now := apptime.Now()
    fmt.Printf("Current time: %s\n", now.Format("2006-01-02 15:04:05 MST"))

    // Parse time in configured timezone
    parsed, err := apptime.Parse("2006-01-02", "2024-01-15")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Parsed time: %s\n", parsed.Format("2006-01-02 15:04:05 MST"))

    // Format current time
    formatted := apptime.FormatNow("2006-01-02 15:04:05")
    fmt.Printf("Formatted now: %s\n", formatted)
}
```

## ⚙️ Configuration

### **Initialize Timezone**

```go
func InitTimezone(timezone string) error
```

**Common Timezones:**

- `UTC` - Coordinated Universal Time
- `America/New_York` - Eastern Time
- `America/Los_Angeles` - Pacific Time
- `Europe/London` - Greenwich Mean Time
- `Asia/Tokyo` - Japan Standard Time
- `Asia/Shanghai` - China Standard Time
- `Asia/Bangkok` - Indochina Time

### **Environment-Based Configuration**

```go
func initializeTime() {
    timezone := os.Getenv("TIMEZONE")
    if timezone == "" {
        timezone = "UTC" // Default to UTC
    }

    if err := apptime.InitTimezone(timezone); err != nil {
        log.Printf("Failed to set timezone %s, using UTC: %v", timezone, err)
        apptime.InitTimezone("UTC")
    }
}
```

### **Configuration in Application**

```go
// In config/config.go
type TimeConfig struct {
    Timezone string `env:"TIMEZONE" envDefault:"UTC"`
}

func (cfg *Config) InitializeTime() error {
    return apptime.InitTimezone(cfg.Time.Timezone)
}
```

## 🕒 Core Functions

### **Current Time Functions**

```go
// Get current time in configured timezone
func Now() time.Time

// Format current time with layout
func FormatNow(layout string) string

// Get configured timezone location
func GetLocation() *time.Location
```

### **Parsing Functions**

```go
// Parse time string in configured timezone
func Parse(layout, value string) (time.Time, error)

// Parse time in specific location then convert to configured timezone
func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error)
```

### **Conversion Functions**

```go
// Create time from Unix timestamp in configured timezone
func Unix(sec int64, nsec int64) time.Time

// Create time with date components in configured timezone
func Date(year int, month time.Month, day, hour, min, sec, nsec int) time.Time

// Convert any time to configured timezone
func ToLocal(t time.Time) time.Time

// Format time with timezone conversion
func Format(t time.Time, layout string) string
```

### **Date Range Functions**

```go
// Get start of day (00:00:00) in configured timezone
func StartOfDay(t time.Time) time.Time

// Get end of day (23:59:59.999999999) in configured timezone
func EndOfDay(t time.Time) time.Time
```

## 🌍 Timezone Handling

### **Multi-Environment Setup**

```go
func setupTimezone(env string) error {
    var timezone string

    switch env {
    case "production":
        timezone = "UTC" // Always use UTC in production
    case "staging":
        timezone = "UTC" // Consistent with production
    case "development":
        timezone = "America/New_York" // Local development timezone
    case "test":
        timezone = "UTC" // Consistent testing
    default:
        timezone = "UTC"
    }

    return apptime.InitTimezone(timezone)
}
```

### **User-Specific Timezones**

```go
type User struct {
    ID       string
    Name     string
    Timezone string // User's preferred timezone
}

func formatTimeForUser(t time.Time, user *User) string {
    userLocation, err := time.LoadLocation(user.Timezone)
    if err != nil {
        // Fallback to app default
        userLocation = apptime.GetLocation()
    }

    return t.In(userLocation).Format("2006-01-02 15:04:05 MST")
}
```

## 💡 Examples

### **1. API Response Time Formatting**

```go
type APIResponse struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (r *APIResponse) MarshalJSON() ([]byte, error) {
    type Alias APIResponse
    return json.Marshal(&struct {
        CreatedAt string `json:"created_at"`
        UpdatedAt string `json:"updated_at"`
        *Alias
    }{
        CreatedAt: apptime.Format(r.CreatedAt, time.RFC3339),
        UpdatedAt: apptime.Format(r.UpdatedAt, time.RFC3339),
        Alias:     (*Alias)(r),
    })
}

// Usage in handler
func GetUserHandler(c *gin.Context) {
    user := getUserFromDB(c.Param("id"))

    response := APIResponse{
        ID:        user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }

    c.JSON(200, response)
}
```

### **2. Date Range Queries**

```go
func GetOrdersByDateRange(startDate, endDate string) ([]Order, error) {
    // Parse input dates
    start, err := apptime.Parse(apptime.DateLayout, startDate)
    if err != nil {
        return nil, fmt.Errorf("invalid start date: %w", err)
    }

    end, err := apptime.Parse(apptime.DateLayout, endDate)
    if err != nil {
        return nil, fmt.Errorf("invalid end date: %w", err)
    }

    // Get full day ranges
    startOfDay := apptime.StartOfDay(start)
    endOfDay := apptime.EndOfDay(end)

    var orders []Order
    err = db.Where("created_at BETWEEN ? AND ?", startOfDay, endOfDay).Find(&orders).Error

    return orders, err
}

// Usage
orders, err := GetOrdersByDateRange("2024-01-01", "2024-01-31")
```

### **3. Scheduled Jobs and Cron**

```go
type SchedulerService struct {
    timezone *time.Location
}

func NewSchedulerService() *SchedulerService {
    return &SchedulerService{
        timezone: apptime.GetLocation(),
    }
}

func (s *SchedulerService) ScheduleDailyReport() {
    // Schedule daily report at 9 AM in configured timezone
    c := cron.New(cron.WithLocation(s.timezone))

    c.AddFunc("0 9 * * *", func() {
        s.generateDailyReport()
    })

    c.Start()
}

func (s *SchedulerService) generateDailyReport() {
    now := apptime.Now()
    yesterday := now.AddDate(0, 0, -1)

    startOfDay := apptime.StartOfDay(yesterday)
    endOfDay := apptime.EndOfDay(yesterday)

    report := generateReport(startOfDay, endOfDay)
    sendReport(report)

    logger.Info("Daily report generated",
        zap.String("report_date", apptime.Format(yesterday, apptime.DateLayout)),
        zap.Time("generated_at", now),
    )
}
```

### **4. Event Logging with Timestamps**

```go
type EventLogger struct {
    events []Event
}

type Event struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`
    UserID    string    `json:"user_id"`
    Data      any       `json:"data"`
    Timestamp time.Time `json:"timestamp"`
}

func (el *EventLogger) LogEvent(eventType, userID string, data any) {
    event := Event{
        ID:        uuid.New().String(),
        Type:      eventType,
        UserID:    userID,
        Data:      data,
        Timestamp: apptime.Now(),
    }

    el.events = append(el.events, event)

    logger.Info("Event logged",
        zap.String("event_id", event.ID),
        zap.String("event_type", eventType),
        zap.String("user_id", userID),
        zap.Time("timestamp", event.Timestamp),
    )
}

func (el *EventLogger) GetEventsByTimeRange(start, end time.Time) []Event {
    var filteredEvents []Event

    startTime := apptime.ToLocal(start)
    endTime := apptime.ToLocal(end)

    for _, event := range el.events {
        eventTime := apptime.ToLocal(event.Timestamp)
        if eventTime.After(startTime) && eventTime.Before(endTime) {
            filteredEvents = append(filteredEvents, event)
        }
    }

    return filteredEvents
}
```

### **5. Time-Based Business Logic**

```go
type BusinessHours struct {
    Timezone  string
    StartHour int
    EndHour   int
    WeekDays  []time.Weekday
}

func (bh *BusinessHours) IsBusinessHours(t time.Time) bool {
    location, err := time.LoadLocation(bh.Timezone)
    if err != nil {
        return false
    }

    localTime := t.In(location)

    // Check if it's a business day
    isBusinessDay := false
    for _, weekday := range bh.WeekDays {
        if localTime.Weekday() == weekday {
            isBusinessDay = true
            break
        }
    }

    if !isBusinessDay {
        return false
    }

    // Check if it's within business hours
    hour := localTime.Hour()
    return hour >= bh.StartHour && hour < bh.EndHour
}

// Usage
businessHours := &BusinessHours{
    Timezone:  "America/New_York",
    StartHour: 9,  // 9 AM
    EndHour:   17, // 5 PM
    WeekDays: []time.Weekday{
        time.Monday, time.Tuesday, time.Wednesday,
        time.Thursday, time.Friday,
    },
}

func ProcessOrder(order *Order) error {
    now := apptime.Now()

    if businessHours.IsBusinessHours(now) {
        // Process immediately during business hours
        return processOrderImmediate(order)
    } else {
        // Queue for next business day
        return queueOrderForBusinessHours(order)
    }
}
```

### **6. Time Zone Conversion for API**

```go
type TimezonedResponse struct {
    Data      any       `json:"data"`
    Timestamp time.Time `json:"timestamp"`
    Timezone  string    `json:"timezone"`
}

func RespondWithTimezone(c *gin.Context, data any) {
    // Get user's preferred timezone from header or default
    userTimezone := c.GetHeader("X-Timezone")
    if userTimezone == "" {
        userTimezone = apptime.GetLocation().String()
    }

    location, err := time.LoadLocation(userTimezone)
    if err != nil {
        location = apptime.GetLocation()
        userTimezone = location.String()
    }

    now := apptime.Now().In(location)

    response := TimezonedResponse{
        Data:      data,
        Timestamp: now,
        Timezone:  userTimezone,
    }

    c.JSON(200, response)
}

// Middleware to parse user timezone
func TimezoneMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        timezone := c.GetHeader("X-Timezone")
        if timezone != "" {
            if _, err := time.LoadLocation(timezone); err == nil {
                c.Set("user_timezone", timezone)
            }
        }
        c.Next()
    }
}
```

## 🎯 Best Practices

### **1. Always Use UTC for Storage**

```go
// ✅ DO: Store times in UTC in database
type User struct {
    ID        string
    CreatedAt time.Time // Store in UTC
    UpdatedAt time.Time // Store in UTC
}

// ✅ DO: Convert to user timezone for display
func formatForUser(t time.Time, userTimezone string) string {
    location, err := time.LoadLocation(userTimezone)
    if err != nil {
        location = time.UTC
    }

    return t.In(location).Format("2006-01-02 15:04:05 MST")
}
```

### **2. Consistent Time Parsing**

```go
// ✅ DO: Use predefined layouts
const (
    APIDateLayout     = "2006-01-02"
    APITimeLayout     = "15:04:05"
    APIDateTimeLayout = "2006-01-02T15:04:05Z"
)

func parseAPIDate(dateStr string) (time.Time, error) {
    return apptime.Parse(APIDateLayout, dateStr)
}

func parseAPIDateTime(dateTimeStr string) (time.Time, error) {
    return time.Parse(APIDateTimeLayout, dateTimeStr)
}
```

### **3. Time Range Validation**

```go
func validateDateRange(startDate, endDate string) error {
    start, err := apptime.Parse(apptime.DateLayout, startDate)
    if err != nil {
        return fmt.Errorf("invalid start date: %w", err)
    }

    end, err := apptime.Parse(apptime.DateLayout, endDate)
    if err != nil {
        return fmt.Errorf("invalid end date: %w", err)
    }

    if start.After(end) {
        return errors.New("start date cannot be after end date")
    }

    // Limit range to prevent abuse
    if end.Sub(start) > 365*24*time.Hour {
        return errors.New("date range cannot exceed 365 days")
    }

    return nil
}
```

### **4. Testing with Fixed Times**

```go
func TestTimeOperations(t *testing.T) {
    // Set up test timezone
    err := apptime.InitTimezone("UTC")
    require.NoError(t, err)

    // Use fixed time for testing
    fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

    // Test start of day
    startOfDay := apptime.StartOfDay(fixedTime)
    expectedStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
    assert.Equal(t, expectedStart, startOfDay)

    // Test end of day
    endOfDay := apptime.EndOfDay(fixedTime)
    expectedEnd := time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC)
    assert.Equal(t, expectedEnd, endOfDay)
}
```

### **5. Error Handling for Timezones**

```go
func safeLoadLocation(timezone string) *time.Location {
    location, err := time.LoadLocation(timezone)
    if err != nil {
        logger.Warn("Invalid timezone, using UTC",
            zap.String("requested_timezone", timezone),
            zap.Error(err),
        )
        return time.UTC
    }
    return location
}

func getUserTimezoneSafely(user *User) *time.Location {
    if user.Timezone == "" {
        return apptime.GetLocation()
    }

    return safeLoadLocation(user.Timezone)
}
```

## 📊 Common Layout Constants

```go
// Built-in layouts from the package
const (
    DateLayout     = "2006-01-02"              // ISO date format
    TimeLayout     = "15:04:05"               // 24-hour time format
    DateTimeLayout = "2006-01-02 15:04:05"    // Combined date/time
    ISO8601Layout  = "2006-01-02T15:04:05Z07:00" // ISO 8601 format
)

// Custom layouts for specific needs
const (
    DisplayDateLayout = "January 2, 2006"     // Human-readable date
    LogTimeLayout     = "2006-01-02 15:04:05.000" // Millisecond precision
    FileNameLayout    = "20060102_150405"     // Safe for filenames
    APITimestamp      = "2006-01-02T15:04:05.000Z" // API responses
)
```

## 🔧 Migration and Utilities

### **Database Migration Helper**

```go
func migrateTimestampsToTimezone(db *gorm.DB, timezone string) error {
    location, err := time.LoadLocation(timezone)
    if err != nil {
        return err
    }

    // Example: Update all created_at fields
    return db.Exec(`
        UPDATE users
        SET created_at = created_at AT TIME ZONE 'UTC' AT TIME ZONE ?
    `, timezone).Error
}
```

### **Time Comparison Utilities**

```go
func IsSameDay(t1, t2 time.Time) bool {
    local1 := apptime.ToLocal(t1)
    local2 := apptime.ToLocal(t2)

    return local1.Year() == local2.Year() &&
           local1.Month() == local2.Month() &&
           local1.Day() == local2.Day()
}

func DaysBetween(start, end time.Time) int {
    startDay := apptime.StartOfDay(start)
    endDay := apptime.StartOfDay(end)

    return int(endDay.Sub(startDay).Hours() / 24)
}
```

## 🔗 Related Packages

- [`config`](../../config/) - Time configuration
- [`pkg/logger`](../logger/) - Timestamp logging
- [`internal/middleware`](../../internal/middleware/) - Request time tracking

## 📚 Additional Resources

- [Go Time Package Documentation](https://pkg.go.dev/time)
- [Time Zone Database](https://www.iana.org/time-zones)
- [Working with Time Zones in Go](https://golang.org/doc/articles/wiki/)
- [ISO 8601 Date and Time Format](https://en.wikipedia.org/wiki/ISO_8601)
