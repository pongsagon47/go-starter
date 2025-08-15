package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"go-starter/pkg/logger"

	"go.uber.org/zap"
)

// SimpleMonitor implements Monitor interface
type SimpleMonitor struct {
	registry     Registry
	mu           sync.RWMutex
	healthChecks map[string]HealthChecker
	exporters    []Exporter

	// System metrics collection
	systemMetricsEnabled bool
	systemMetricsTicker  *time.Ticker
	systemMetricsStop    chan struct{}

	// Metrics
	systemMetrics      *SystemMetrics
	applicationMetrics *ApplicationMetrics

	logger *zap.Logger
}

// NewSimpleMonitor creates a new simple monitor
func NewSimpleMonitor() *SimpleMonitor {
	return &SimpleMonitor{
		registry:     NewSimpleRegistry(),
		healthChecks: make(map[string]HealthChecker),
		exporters:    make([]Exporter, 0),
		logger:       logger.Logger,
	}
}

// Registry returns the metrics registry
func (m *SimpleMonitor) Registry() Registry {
	return m.registry
}

// StartSystemMetricsCollection starts collecting system metrics
func (m *SimpleMonitor) StartSystemMetricsCollection(interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.systemMetricsEnabled {
		return nil // Already started
	}

	m.systemMetricsEnabled = true
	m.systemMetricsTicker = time.NewTicker(interval)
	m.systemMetricsStop = make(chan struct{})

	go m.systemMetricsLoop()

	m.logger.Info("Started system metrics collection",
		zap.Duration("interval", interval),
	)

	return nil
}

// StopSystemMetricsCollection stops collecting system metrics
func (m *SimpleMonitor) StopSystemMetricsCollection() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.systemMetricsEnabled {
		return
	}

	m.systemMetricsEnabled = false
	if m.systemMetricsTicker != nil {
		m.systemMetricsTicker.Stop()
	}

	if m.systemMetricsStop != nil {
		close(m.systemMetricsStop)
	}

	m.logger.Info("Stopped system metrics collection")
}

// GetSystemMetrics returns current system metrics
func (m *SimpleMonitor) GetSystemMetrics() (*SystemMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.systemMetrics == nil {
		// Collect once if not already collecting
		metrics := m.collectSystemMetrics()
		return &metrics, nil
	}

	return m.systemMetrics, nil
}

// GetApplicationMetrics returns current application metrics
func (m *SimpleMonitor) GetApplicationMetrics() (*ApplicationMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := m.collectApplicationMetrics()
	return &metrics, nil
}

// RegisterHealthCheck registers a health check
func (m *SimpleMonitor) RegisterHealthCheck(checker HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.healthChecks[checker.Name()] = checker

	m.logger.Info("Registered health check",
		zap.String("name", checker.Name()),
	)
}

// UnregisterHealthCheck unregisters a health check
func (m *SimpleMonitor) UnregisterHealthCheck(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.healthChecks, name)

	m.logger.Info("Unregistered health check",
		zap.String("name", name),
	)
}

// GetHealthChecks returns all health check results
func (m *SimpleMonitor) GetHealthChecks(ctx context.Context) ([]HealthCheck, error) {
	m.mu.RLock()
	checkers := make([]HealthChecker, 0, len(m.healthChecks))
	for _, checker := range m.healthChecks {
		checkers = append(checkers, checker)
	}
	m.mu.RUnlock()

	results := make([]HealthCheck, 0, len(checkers))

	for _, checker := range checkers {
		result := checker.Check(ctx)
		results = append(results, result)
	}

	return results, nil
}

// GetOverallHealth returns overall system health
func (m *SimpleMonitor) GetOverallHealth(ctx context.Context) (HealthStatus, error) {
	checks, err := m.GetHealthChecks(ctx)
	if err != nil {
		return HealthStatusUnhealthy, err
	}

	if len(checks) == 0 {
		return HealthStatusHealthy, nil
	}

	healthyCount := 0
	degradedCount := 0

	for _, check := range checks {
		switch check.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			return HealthStatusUnhealthy, nil
		}
	}

	if degradedCount > 0 {
		return HealthStatusDegraded, nil
	}

	return HealthStatusHealthy, nil
}

// AddExporter adds a metrics exporter
func (m *SimpleMonitor) AddExporter(exporter Exporter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.exporters = append(m.exporters, exporter)

	m.logger.Info("Added metrics exporter")
}

// RemoveExporter removes a metrics exporter
func (m *SimpleMonitor) RemoveExporter(exporter Exporter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, exp := range m.exporters {
		if exp == exporter {
			m.exporters = append(m.exporters[:i], m.exporters[i+1:]...)
			break
		}
	}

	m.logger.Info("Removed metrics exporter")
}

// Export exports metrics to all registered exporters
func (m *SimpleMonitor) Export(ctx context.Context) error {
	metrics := m.registry.GetAllMetrics()

	m.mu.RLock()
	exporters := make([]Exporter, len(m.exporters))
	copy(exporters, m.exporters)
	m.mu.RUnlock()

	for _, exporter := range exporters {
		if err := exporter.Export(ctx, metrics); err != nil {
			m.logger.Error("Failed to export metrics",
				zap.Error(err),
			)
		}
	}

	return nil
}

// Close closes the monitor
func (m *SimpleMonitor) Close() error {
	m.StopSystemMetricsCollection()

	m.mu.RLock()
	exporters := make([]Exporter, len(m.exporters))
	copy(exporters, m.exporters)
	m.mu.RUnlock()

	for _, exporter := range exporters {
		if err := exporter.Close(); err != nil {
			m.logger.Error("Failed to close exporter",
				zap.Error(err),
			)
		}
	}

	return nil
}

// Private methods

func (m *SimpleMonitor) systemMetricsLoop() {
	for {
		select {
		case <-m.systemMetricsTicker.C:
			metrics := m.collectSystemMetrics()

			m.mu.Lock()
			m.systemMetrics = &metrics
			m.mu.Unlock()

		case <-m.systemMetricsStop:
			return
		}
	}
}

func (m *SimpleMonitor) collectSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemMetrics{
		MemoryUsage:     float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		MemoryTotal:     int64(memStats.Sys),
		MemoryAvailable: int64(memStats.Sys - memStats.Alloc),
		Goroutines:      runtime.NumGoroutine(),
		Timestamp:       time.Now(),

		// Note: Some metrics require platform-specific implementations
		// This is a basic implementation for demonstration
		CPUUsage:            0, // Would need platform-specific code
		DiskUsage:           0, // Would need platform-specific code
		DiskTotal:           0, // Would need platform-specific code
		DiskAvailable:       0, // Would need platform-specific code
		NetworkBytesIn:      0, // Would need platform-specific code
		NetworkBytesOut:     0, // Would need platform-specific code
		LoadAverage1m:       0, // Would need platform-specific code
		LoadAverage5m:       0, // Would need platform-specific code
		LoadAverage15m:      0, // Would need platform-specific code
		OpenFileDescriptors: 0, // Would need platform-specific code
	}
}

func (m *SimpleMonitor) collectApplicationMetrics() ApplicationMetrics {
	allMetrics := m.registry.GetAllMetrics()

	metrics := ApplicationMetrics{
		CustomMetrics: make(map[string]float64),
		Timestamp:     time.Now(),
	}

	// Extract application metrics from registry
	for _, metric := range allMetrics {
		switch metric.Name {
		case "http_requests_total":
			metrics.HTTPRequestsTotal = int64(metric.Value)
		case "http_request_duration":
			metrics.HTTPRequestDuration = metric.Value
		case "http_requests_in_flight":
			metrics.HTTPRequestsInFlight = int64(metric.Value)
		case "database_connections":
			metrics.DatabaseConnections = int64(metric.Value)
		case "cache_hits":
			metrics.CacheHits = int64(metric.Value)
		case "cache_misses":
			metrics.CacheMisses = int64(metric.Value)
		case "queue_size":
			metrics.QueueSize = int64(metric.Value)
		case "queue_processing_time":
			metrics.QueueProcessingTime = metric.Value
		case "errors_total":
			metrics.ErrorsTotal = int64(metric.Value)
		case "active_users":
			metrics.ActiveUsers = int64(metric.Value)
		default:
			metrics.CustomMetrics[metric.Name] = metric.Value
		}
	}

	return metrics
}

// Default monitor instance
var DefaultMonitor = NewSimpleMonitor()

// Convenience functions for the default monitor

// StartSystemMetrics starts system metrics collection on the default monitor
func StartSystemMetrics(interval time.Duration) error {
	return DefaultMonitor.StartSystemMetricsCollection(interval)
}

// StopSystemMetrics stops system metrics collection on the default monitor
func StopSystemMetrics() {
	DefaultMonitor.StopSystemMetricsCollection()
}

// GetSystemMetrics returns system metrics from the default monitor
func GetSystemMetrics() (*SystemMetrics, error) {
	return DefaultMonitor.GetSystemMetrics()
}

// GetApplicationMetrics returns application metrics from the default monitor
func GetApplicationMetrics() (*ApplicationMetrics, error) {
	return DefaultMonitor.GetApplicationMetrics()
}

// RegisterHealthCheck registers a health check on the default monitor
func RegisterHealthCheck(checker HealthChecker) {
	DefaultMonitor.RegisterHealthCheck(checker)
}

// GetHealthChecks returns health checks from the default monitor
func GetHealthChecks(ctx context.Context) ([]HealthCheck, error) {
	return DefaultMonitor.GetHealthChecks(ctx)
}

// GetOverallHealth returns overall health from the default monitor
func GetOverallHealth(ctx context.Context) (HealthStatus, error) {
	return DefaultMonitor.GetOverallHealth(ctx)
}

// AddExporter adds an exporter to the default monitor
func AddExporter(exporter Exporter) {
	DefaultMonitor.AddExporter(exporter)
}

// Export exports metrics from the default monitor
func Export(ctx context.Context) error {
	return DefaultMonitor.Export(ctx)
}
