package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"go-starter/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	name string
	db   *sql.DB
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name string, db *sql.DB) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name: name,
		db:   db,
	}
}

// Name returns the name of the health check
func (d *DatabaseHealthChecker) Name() string {
	return d.name
}

// Check performs the health check
func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	// Ping the database
	err := d.db.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		return HealthCheck{
			Name:      d.name,
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Database ping failed: %v", err),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	// Check if response time is acceptable
	if duration > 5*time.Second {
		return HealthCheck{
			Name:      d.name,
			Status:    HealthStatusDegraded,
			Message:   fmt.Sprintf("Database response time is slow: %v", duration),
			Duration:  duration,
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"response_time": duration.String(),
			},
		}
	}

	return HealthCheck{
		Name:      d.name,
		Status:    HealthStatusHealthy,
		Message:   "Database is healthy",
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"response_time": duration.String(),
		},
	}
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	name   string
	client *redis.Client
}

// NewRedisHealthChecker creates a new Redis health checker
func NewRedisHealthChecker(name string, client *redis.Client) *RedisHealthChecker {
	return &RedisHealthChecker{
		name:   name,
		client: client,
	}
}

// Name returns the name of the health check
func (r *RedisHealthChecker) Name() string {
	return r.name
}

// Check performs the health check
func (r *RedisHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	// Ping Redis
	err := r.client.Ping(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		return HealthCheck{
			Name:      r.name,
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Redis ping failed: %v", err),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}

	// Check if response time is acceptable
	if duration > 2*time.Second {
		return HealthCheck{
			Name:      r.name,
			Status:    HealthStatusDegraded,
			Message:   fmt.Sprintf("Redis response time is slow: %v", duration),
			Duration:  duration,
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"response_time": duration.String(),
			},
		}
	}

	return HealthCheck{
		Name:      r.name,
		Status:    HealthStatusHealthy,
		Message:   "Redis is healthy",
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"response_time": duration.String(),
		},
	}
}

// HTTPServiceHealthChecker checks external HTTP service
type HTTPServiceHealthChecker struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPServiceHealthChecker creates a new HTTP service health checker
func NewHTTPServiceHealthChecker(name, url string) *HTTPServiceHealthChecker {
	return &HTTPServiceHealthChecker{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the name of the health check
func (h *HTTPServiceHealthChecker) Name() string {
	return h.name
}

// Check performs the health check
func (h *HTTPServiceHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return HealthCheck{
			Name:      h.name,
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Failed to create request: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	resp, err := h.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return HealthCheck{
			Name:      h.name,
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("HTTP request failed: %v", err),
			Duration:  duration,
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status := HealthStatusHealthy
		message := "Service is healthy"

		// Check if response time is acceptable
		if duration > 5*time.Second {
			status = HealthStatusDegraded
			message = fmt.Sprintf("Service response time is slow: %v", duration)
		}

		return HealthCheck{
			Name:      h.name,
			Status:    status,
			Message:   message,
			Duration:  duration,
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"status_code":   fmt.Sprintf("%d", resp.StatusCode),
				"response_time": duration.String(),
			},
		}
	}

	return HealthCheck{
		Name:      h.name,
		Status:    HealthStatusUnhealthy,
		Message:   fmt.Sprintf("Service returned status code: %d", resp.StatusCode),
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		},
	}
}

// MemoryHealthChecker checks memory usage
type MemoryHealthChecker struct {
	name              string
	warningThreshold  float64 // Memory usage percentage to trigger warning
	criticalThreshold float64 // Memory usage percentage to trigger critical
}

// NewMemoryHealthChecker creates a new memory health checker
func NewMemoryHealthChecker(name string, warningThreshold, criticalThreshold float64) *MemoryHealthChecker {
	return &MemoryHealthChecker{
		name:              name,
		warningThreshold:  warningThreshold,
		criticalThreshold: criticalThreshold,
	}
}

// Name returns the name of the health check
func (m *MemoryHealthChecker) Name() string {
	return m.name
}

// Check performs the health check
func (m *MemoryHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	systemMetrics, err := GetSystemMetrics()
	if err != nil {
		return HealthCheck{
			Name:      m.name,
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("Failed to get system metrics: %v", err),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}

	memoryUsage := systemMetrics.MemoryUsage
	duration := time.Since(start)

	status := HealthStatusHealthy
	message := fmt.Sprintf("Memory usage: %.2f%%", memoryUsage)

	if memoryUsage >= m.criticalThreshold {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Critical memory usage: %.2f%% (threshold: %.2f%%)", memoryUsage, m.criticalThreshold)
	} else if memoryUsage >= m.warningThreshold {
		status = HealthStatusDegraded
		message = fmt.Sprintf("High memory usage: %.2f%% (threshold: %.2f%%)", memoryUsage, m.warningThreshold)
	}

	return HealthCheck{
		Name:      m.name,
		Status:    status,
		Message:   message,
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"memory_usage":       fmt.Sprintf("%.2f%%", memoryUsage),
			"warning_threshold":  fmt.Sprintf("%.2f%%", m.warningThreshold),
			"critical_threshold": fmt.Sprintf("%.2f%%", m.criticalThreshold),
			"memory_total":       fmt.Sprintf("%d", systemMetrics.MemoryTotal),
			"memory_available":   fmt.Sprintf("%d", systemMetrics.MemoryAvailable),
		},
	}
}

// CustomHealthChecker allows custom health check logic
type CustomHealthChecker struct {
	name    string
	checkFn func(ctx context.Context) (HealthStatus, string, map[string]string)
}

// NewCustomHealthChecker creates a new custom health checker
func NewCustomHealthChecker(name string, checkFn func(ctx context.Context) (HealthStatus, string, map[string]string)) *CustomHealthChecker {
	return &CustomHealthChecker{
		name:    name,
		checkFn: checkFn,
	}
}

// Name returns the name of the health check
func (c *CustomHealthChecker) Name() string {
	return c.name
}

// Check performs the health check
func (c *CustomHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	status, message, metadata := c.checkFn(ctx)
	duration := time.Since(start)

	return HealthCheck{
		Name:      c.name,
		Status:    status,
		Message:   message,
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}

// Common health check functions

// CheckDatabaseConnections checks database connection pool
func CheckDatabaseConnections(db *sql.DB) (HealthStatus, string, map[string]string) {
	stats := db.Stats()

	metadata := map[string]string{
		"open_connections":    fmt.Sprintf("%d", stats.OpenConnections),
		"in_use":              fmt.Sprintf("%d", stats.InUse),
		"idle":                fmt.Sprintf("%d", stats.Idle),
		"wait_count":          fmt.Sprintf("%d", stats.WaitCount),
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     fmt.Sprintf("%d", stats.MaxIdleClosed),
		"max_lifetime_closed": fmt.Sprintf("%d", stats.MaxLifetimeClosed),
	}

	// Check for potential issues
	if stats.WaitCount > 0 {
		return HealthStatusDegraded,
			fmt.Sprintf("Database connections are waiting: %d waits", stats.WaitCount),
			metadata
	}

	if stats.OpenConnections > 80 { // Assuming max 100 connections
		return HealthStatusDegraded,
			fmt.Sprintf("High number of open connections: %d", stats.OpenConnections),
			metadata
	}

	return HealthStatusHealthy, "Database connections are healthy", metadata
}

// CheckDiskSpace checks available disk space
func CheckDiskSpace(path string, warningThreshold, criticalThreshold float64) (HealthStatus, string, map[string]string) {
	// This would need platform-specific implementation
	// For now, return healthy as placeholder

	metadata := map[string]string{
		"path":               path,
		"warning_threshold":  fmt.Sprintf("%.2f%%", warningThreshold),
		"critical_threshold": fmt.Sprintf("%.2f%%", criticalThreshold),
	}

	return HealthStatusHealthy, "Disk space check not implemented", metadata
}

// RegisterCommonHealthChecks registers common health checks
func RegisterCommonHealthChecks(db *sql.DB, redisClient *redis.Client) {
	// Database health check
	if db != nil {
		dbChecker := NewDatabaseHealthChecker("database", db)
		RegisterHealthCheck(dbChecker)

		// Database connections health check
		dbConnChecker := NewCustomHealthChecker("database_connections", func(ctx context.Context) (HealthStatus, string, map[string]string) {
			return CheckDatabaseConnections(db)
		})
		RegisterHealthCheck(dbConnChecker)
	}

	// Redis health check
	if redisClient != nil {
		redisChecker := NewRedisHealthChecker("redis", redisClient)
		RegisterHealthCheck(redisChecker)
	}

	// Memory health check
	memoryChecker := NewMemoryHealthChecker("memory", 80.0, 95.0) // 80% warning, 95% critical
	RegisterHealthCheck(memoryChecker)

	logger.Info("Registered common health checks")
}
