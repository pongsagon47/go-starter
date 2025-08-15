package metrics

import (
	"context"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Labels represents metric labels
type Labels map[string]string

// Metric represents a single metric measurement
type Metric struct {
	Name      string     `json:"name"`
	Type      MetricType `json:"type"`
	Value     float64    `json:"value"`
	Labels    Labels     `json:"labels,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	Help      string     `json:"help,omitempty"`
}

// Counter represents a counter metric (monotonically increasing)
type Counter interface {
	// Inc increments the counter by 1
	Inc()

	// Add adds the given value to the counter
	Add(value float64)

	// Get returns the current value
	Get() float64

	// With returns a new counter with additional labels
	With(labels Labels) Counter
}

// Gauge represents a gauge metric (can go up and down)
type Gauge interface {
	// Set sets the gauge to the given value
	Set(value float64)

	// Inc increments the gauge by 1
	Inc()

	// Dec decrements the gauge by 1
	Dec()

	// Add adds the given value to the gauge
	Add(value float64)

	// Sub subtracts the given value from the gauge
	Sub(value float64)

	// Get returns the current value
	Get() float64

	// With returns a new gauge with additional labels
	With(labels Labels) Gauge
}

// Histogram represents a histogram metric
type Histogram interface {
	// Observe adds a single observation
	Observe(value float64)

	// ObserveDuration observes a duration
	ObserveDuration(start time.Time)

	// With returns a new histogram with additional labels
	With(labels Labels) Histogram
}

// Timer helps measure durations
type Timer interface {
	// Start starts the timer and returns a function to stop it
	Start() func()

	// ObserveDuration observes the duration since the given time
	ObserveDuration(start time.Time)
}

// Registry manages metrics
type Registry interface {
	// NewCounter creates a new counter
	NewCounter(name, help string, labels Labels) Counter

	// NewGauge creates a new gauge
	NewGauge(name, help string, labels Labels) Gauge

	// NewHistogram creates a new histogram
	NewHistogram(name, help string, buckets []float64, labels Labels) Histogram

	// NewTimer creates a new timer for the given histogram
	NewTimer(histogram Histogram) Timer

	// GetMetric retrieves a metric by name
	GetMetric(name string) (interface{}, bool)

	// GetAllMetrics returns all registered metrics
	GetAllMetrics() []Metric

	// Unregister removes a metric
	Unregister(name string)

	// Clear removes all metrics
	Clear()
}

// Collector collects metrics for export
type Collector interface {
	// Collect gathers all current metric values
	Collect() ([]Metric, error)

	// Describe returns metric descriptions
	Describe() []MetricDescription
}

// MetricDescription describes a metric
type MetricDescription struct {
	Name   string     `json:"name"`
	Type   MetricType `json:"type"`
	Help   string     `json:"help"`
	Labels []string   `json:"labels"`
}

// Exporter exports metrics to external systems
type Exporter interface {
	// Export sends metrics to the external system
	Export(ctx context.Context, metrics []Metric) error

	// Close closes the exporter
	Close() error
}

// Config holds metrics configuration
type Config struct {
	Enabled       bool          `json:"enabled"`
	Namespace     string        `json:"namespace"`
	Subsystem     string        `json:"subsystem"`
	Labels        Labels        `json:"labels"`
	Buckets       []float64     `json:"buckets"` // Default histogram buckets
	FlushInterval time.Duration `json:"flush_interval"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPUUsage            float64   `json:"cpu_usage"`
	MemoryUsage         float64   `json:"memory_usage"`
	MemoryTotal         int64     `json:"memory_total"`
	MemoryAvailable     int64     `json:"memory_available"`
	DiskUsage           float64   `json:"disk_usage"`
	DiskTotal           int64     `json:"disk_total"`
	DiskAvailable       int64     `json:"disk_available"`
	NetworkBytesIn      int64     `json:"network_bytes_in"`
	NetworkBytesOut     int64     `json:"network_bytes_out"`
	LoadAverage1m       float64   `json:"load_average_1m"`
	LoadAverage5m       float64   `json:"load_average_5m"`
	LoadAverage15m      float64   `json:"load_average_15m"`
	OpenFileDescriptors int       `json:"open_file_descriptors"`
	Goroutines          int       `json:"goroutines"`
	Timestamp           time.Time `json:"timestamp"`
}

// ApplicationMetrics represents application-level metrics
type ApplicationMetrics struct {
	HTTPRequestsTotal    int64              `json:"http_requests_total"`
	HTTPRequestDuration  float64            `json:"http_request_duration"`
	HTTPRequestsInFlight int64              `json:"http_requests_in_flight"`
	DatabaseConnections  int64              `json:"database_connections"`
	CacheHits            int64              `json:"cache_hits"`
	CacheMisses          int64              `json:"cache_misses"`
	QueueSize            int64              `json:"queue_size"`
	QueueProcessingTime  float64            `json:"queue_processing_time"`
	ErrorsTotal          int64              `json:"errors_total"`
	ActiveUsers          int64              `json:"active_users"`
	CustomMetrics        map[string]float64 `json:"custom_metrics"`
	Timestamp            time.Time          `json:"timestamp"`
}

// HealthStatus represents service health
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check
type HealthCheck struct {
	Name      string            `json:"name"`
	Status    HealthStatus      `json:"status"`
	Message   string            `json:"message,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// HealthChecker performs health checks
type HealthChecker interface {
	// Check performs the health check
	Check(ctx context.Context) HealthCheck

	// Name returns the name of the health check
	Name() string
}

// Monitor provides monitoring capabilities
type Monitor interface {
	// Registry returns the metrics registry
	Registry() Registry

	// StartSystemMetricsCollection starts collecting system metrics
	StartSystemMetricsCollection(interval time.Duration) error

	// StopSystemMetricsCollection stops collecting system metrics
	StopSystemMetricsCollection()

	// GetSystemMetrics returns current system metrics
	GetSystemMetrics() (*SystemMetrics, error)

	// GetApplicationMetrics returns current application metrics
	GetApplicationMetrics() (*ApplicationMetrics, error)

	// RegisterHealthCheck registers a health check
	RegisterHealthCheck(checker HealthChecker)

	// UnregisterHealthCheck unregisters a health check
	UnregisterHealthCheck(name string)

	// GetHealthChecks returns all health check results
	GetHealthChecks(ctx context.Context) ([]HealthCheck, error)

	// GetOverallHealth returns overall system health
	GetOverallHealth(ctx context.Context) (HealthStatus, error)

	// AddExporter adds a metrics exporter
	AddExporter(exporter Exporter)

	// RemoveExporter removes a metrics exporter
	RemoveExporter(exporter Exporter)

	// Export exports metrics to all registered exporters
	Export(ctx context.Context) error

	// Close closes the monitor
	Close() error
}
