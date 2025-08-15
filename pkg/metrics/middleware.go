package metrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPMetricsMiddleware creates middleware for collecting HTTP metrics
func HTTPMetricsMiddleware() gin.HandlerFunc {
	// Create metrics
	requestsTotal := NewCounter("http_requests_total", "Total number of HTTP requests", Labels{})
	requestDuration := NewHistogram("http_request_duration_seconds", "HTTP request duration in seconds", nil, Labels{})
	requestsInFlight := NewGauge("http_requests_in_flight", "Number of HTTP requests currently being processed", Labels{})
	requestSize := NewHistogram("http_request_size_bytes", "HTTP request size in bytes", []float64{100, 1000, 10000, 100000, 1000000}, Labels{})
	responseSize := NewHistogram("http_response_size_bytes", "HTTP response size in bytes", []float64{100, 1000, 10000, 100000, 1000000}, Labels{})

	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Request size
		if c.Request.ContentLength > 0 {
			requestSize.With(Labels{
				"method": c.Request.Method,
				"path":   path,
			}).Observe(float64(c.Request.ContentLength))
		}

		// Increment in-flight requests
		inFlightGauge := requestsInFlight.With(Labels{
			"method": c.Request.Method,
			"path":   path,
		})
		inFlightGauge.Inc()

		// Process request
		c.Next()

		// Decrement in-flight requests
		inFlightGauge.Dec()

		// Record metrics
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		labels := Labels{
			"method": c.Request.Method,
			"path":   path,
			"status": status,
		}

		// Total requests
		requestsTotal.With(labels).Inc()

		// Request duration
		requestDuration.With(labels).Observe(duration.Seconds())

		// Response size
		if c.Writer.Size() > 0 {
			responseSize.With(labels).Observe(float64(c.Writer.Size()))
		}
	}
}

// DatabaseMetricsMiddleware creates middleware for collecting database metrics
func DatabaseMetricsMiddleware() gin.HandlerFunc {
	// Create metrics
	dbQueries := NewCounter("database_queries_total", "Total number of database queries", Labels{})
	dbConnections := NewGauge("database_connections_active", "Number of active database connections", Labels{})
	dbQueryDuration := NewHistogram("database_query_duration_seconds", "Database query duration in seconds", nil, Labels{})

	return func(c *gin.Context) {
		// Store metrics in context for use by database layer
		c.Set("db_queries_counter", dbQueries)
		c.Set("db_connections_gauge", dbConnections)
		c.Set("db_query_duration_histogram", dbQueryDuration)

		c.Next()
	}
}

// CacheMetricsMiddleware creates middleware for collecting cache metrics
func CacheMetricsMiddleware() gin.HandlerFunc {
	// Create metrics
	cacheHits := NewCounter("cache_hits_total", "Total number of cache hits", Labels{})
	cacheMisses := NewCounter("cache_misses_total", "Total number of cache misses", Labels{})
	cacheOperations := NewHistogram("cache_operation_duration_seconds", "Cache operation duration in seconds", nil, Labels{})

	return func(c *gin.Context) {
		// Store metrics in context for use by cache layer
		c.Set("cache_hits_counter", cacheHits)
		c.Set("cache_misses_counter", cacheMisses)
		c.Set("cache_operations_histogram", cacheOperations)

		c.Next()
	}
}

// ErrorMetricsMiddleware creates middleware for collecting error metrics
func ErrorMetricsMiddleware() gin.HandlerFunc {
	errorsTotal := NewCounter("errors_total", "Total number of errors", Labels{})

	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				labels := Labels{
					"type":   fmt.Sprintf("%d", err.Type),
					"method": c.Request.Method,
					"path":   c.FullPath(),
				}

				errorsTotal.With(labels).Inc()
			}
		}

		// Check for HTTP error status codes
		status := c.Writer.Status()
		if status >= 400 {
			labels := Labels{
				"status": strconv.Itoa(status),
				"method": c.Request.Method,
				"path":   c.FullPath(),
			}

			errorsTotal.With(labels).Inc()
		}
	}
}

// QueueMetricsMiddleware creates middleware for collecting queue metrics
func QueueMetricsMiddleware() gin.HandlerFunc {
	jobsQueued := NewCounter("queue_jobs_queued_total", "Total number of jobs queued", Labels{})
	jobsProcessed := NewCounter("queue_jobs_processed_total", "Total number of jobs processed", Labels{})
	jobProcessingDuration := NewHistogram("queue_job_processing_duration_seconds", "Job processing duration in seconds", nil, Labels{})
	queueSize := NewGauge("queue_size", "Current queue size", Labels{})

	return func(c *gin.Context) {
		// Store metrics in context for use by queue layer
		c.Set("jobs_queued_counter", jobsQueued)
		c.Set("jobs_processed_counter", jobsProcessed)
		c.Set("job_processing_duration_histogram", jobProcessingDuration)
		c.Set("queue_size_gauge", queueSize)

		c.Next()
	}
}

// AuthMetricsMiddleware creates middleware for collecting authentication metrics
func AuthMetricsMiddleware() gin.HandlerFunc {
	authAttempts := NewCounter("auth_attempts_total", "Total number of authentication attempts", Labels{})
	authSuccesses := NewCounter("auth_successes_total", "Total number of successful authentications", Labels{})
	authFailures := NewCounter("auth_failures_total", "Total number of failed authentications", Labels{})
	activeUsers := NewGauge("active_users", "Number of currently active users", Labels{})

	return func(c *gin.Context) {
		// Store metrics in context for use by auth layer
		c.Set("auth_attempts_counter", authAttempts)
		c.Set("auth_successes_counter", authSuccesses)
		c.Set("auth_failures_counter", authFailures)
		c.Set("active_users_gauge", activeUsers)

		c.Next()
	}
}

// BusinessMetricsMiddleware creates middleware for collecting business metrics
func BusinessMetricsMiddleware() gin.HandlerFunc {
	userRegistrations := NewCounter("user_registrations_total", "Total number of user registrations", Labels{})
	userLogins := NewCounter("user_logins_total", "Total number of user logins", Labels{})
	orders := NewCounter("orders_total", "Total number of orders", Labels{})
	revenue := NewGauge("revenue_total", "Total revenue", Labels{})

	return func(c *gin.Context) {
		// Store metrics in context for use by business logic
		c.Set("user_registrations_counter", userRegistrations)
		c.Set("user_logins_counter", userLogins)
		c.Set("orders_counter", orders)
		c.Set("revenue_gauge", revenue)

		c.Next()
	}
}

// MetricsFromContext extracts metrics from gin context
func MetricsFromContext(c *gin.Context) map[string]interface{} {
	metrics := make(map[string]interface{})

	// HTTP metrics
	if counter, exists := c.Get("http_requests_counter"); exists {
		metrics["http_requests_counter"] = counter
	}

	// Database metrics
	if counter, exists := c.Get("db_queries_counter"); exists {
		metrics["db_queries_counter"] = counter
	}
	if gauge, exists := c.Get("db_connections_gauge"); exists {
		metrics["db_connections_gauge"] = gauge
	}
	if histogram, exists := c.Get("db_query_duration_histogram"); exists {
		metrics["db_query_duration_histogram"] = histogram
	}

	// Cache metrics
	if counter, exists := c.Get("cache_hits_counter"); exists {
		metrics["cache_hits_counter"] = counter
	}
	if counter, exists := c.Get("cache_misses_counter"); exists {
		metrics["cache_misses_counter"] = counter
	}

	// Queue metrics
	if counter, exists := c.Get("jobs_queued_counter"); exists {
		metrics["jobs_queued_counter"] = counter
	}
	if counter, exists := c.Get("jobs_processed_counter"); exists {
		metrics["jobs_processed_counter"] = counter
	}

	// Auth metrics
	if counter, exists := c.Get("auth_attempts_counter"); exists {
		metrics["auth_attempts_counter"] = counter
	}
	if counter, exists := c.Get("auth_successes_counter"); exists {
		metrics["auth_successes_counter"] = counter
	}
	if counter, exists := c.Get("auth_failures_counter"); exists {
		metrics["auth_failures_counter"] = counter
	}

	return metrics
}

// RecordUserRegistration records a user registration metric
func RecordUserRegistration(c *gin.Context, userType string) {
	if counter, exists := c.Get("user_registrations_counter"); exists {
		if userRegCounter, ok := counter.(Counter); ok {
			userRegCounter.With(Labels{"user_type": userType}).Inc()
		}
	}
}

// RecordUserLogin records a user login metric
func RecordUserLogin(c *gin.Context, userType string, success bool) {
	if counter, exists := c.Get("user_logins_counter"); exists {
		if loginCounter, ok := counter.(Counter); ok {
			status := "success"
			if !success {
				status = "failure"
			}
			loginCounter.With(Labels{
				"user_type": userType,
				"status":    status,
			}).Inc()
		}
	}
}

// RecordOrder records an order metric
func RecordOrder(c *gin.Context, orderType string, amount float64) {
	if counter, exists := c.Get("orders_counter"); exists {
		if orderCounter, ok := counter.(Counter); ok {
			orderCounter.With(Labels{"order_type": orderType}).Inc()
		}
	}

	if gauge, exists := c.Get("revenue_gauge"); exists {
		if revenueGauge, ok := gauge.(Gauge); ok {
			revenueGauge.Add(amount)
		}
	}
}

// RecordCacheOperation records a cache operation metric
func RecordCacheOperation(c *gin.Context, operation string, hit bool, duration time.Duration) {
	// Record hit/miss
	if hit {
		if counter, exists := c.Get("cache_hits_counter"); exists {
			if cacheCounter, ok := counter.(Counter); ok {
				cacheCounter.With(Labels{"operation": operation}).Inc()
			}
		}
	} else {
		if counter, exists := c.Get("cache_misses_counter"); exists {
			if cacheCounter, ok := counter.(Counter); ok {
				cacheCounter.With(Labels{"operation": operation}).Inc()
			}
		}
	}

	// Record duration
	if histogram, exists := c.Get("cache_operations_histogram"); exists {
		if cacheHistogram, ok := histogram.(Histogram); ok {
			cacheHistogram.With(Labels{"operation": operation}).Observe(duration.Seconds())
		}
	}
}

// RecordDatabaseQuery records a database query metric
func RecordDatabaseQuery(c *gin.Context, query string, duration time.Duration) {
	if counter, exists := c.Get("db_queries_counter"); exists {
		if dbCounter, ok := counter.(Counter); ok {
			dbCounter.With(Labels{"query_type": query}).Inc()
		}
	}

	if histogram, exists := c.Get("db_query_duration_histogram"); exists {
		if dbHistogram, ok := histogram.(Histogram); ok {
			dbHistogram.With(Labels{"query_type": query}).Observe(duration.Seconds())
		}
	}
}
