package metrics

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// DefaultRegistry is the default metrics registry
var DefaultRegistry = NewSimpleRegistry()

// SimpleRegistry implements Registry interface
type SimpleRegistry struct {
	mu      sync.RWMutex
	metrics map[string]interface{}
	config  *Config
}

// NewSimpleRegistry creates a new simple registry
func NewSimpleRegistry() *SimpleRegistry {
	return &SimpleRegistry{
		metrics: make(map[string]interface{}),
		config: &Config{
			Enabled:   true,
			Namespace: "app",
			Buckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05,
				0.1, 0.25, 0.5, 1, 2.5, 5, 10,
			},
		},
	}
}

// SetConfig sets the registry configuration
func (r *SimpleRegistry) SetConfig(config *Config) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = config
}

// NewCounter creates a new counter
func (r *SimpleRegistry) NewCounter(name, help string, labels Labels) Counter {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullName := r.buildMetricName(name)
	counter := &simpleCounter{
		name:   fullName,
		help:   help,
		labels: r.mergeLabels(labels),
		value:  0,
	}

	r.metrics[fullName] = counter
	return counter
}

// NewGauge creates a new gauge
func (r *SimpleRegistry) NewGauge(name, help string, labels Labels) Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullName := r.buildMetricName(name)
	gauge := &simpleGauge{
		name:   fullName,
		help:   help,
		labels: r.mergeLabels(labels),
		value:  0,
	}

	r.metrics[fullName] = gauge
	return gauge
}

// NewHistogram creates a new histogram
func (r *SimpleRegistry) NewHistogram(name, help string, buckets []float64, labels Labels) Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()

	if buckets == nil {
		buckets = r.config.Buckets
	}

	fullName := r.buildMetricName(name)
	histogram := &simpleHistogram{
		name:    fullName,
		help:    help,
		labels:  r.mergeLabels(labels),
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1),
		sum:     0,
		count:   0,
	}

	r.metrics[fullName] = histogram
	return histogram
}

// NewTimer creates a new timer
func (r *SimpleRegistry) NewTimer(histogram Histogram) Timer {
	return &simpleTimer{
		histogram: histogram,
	}
}

// GetMetric retrieves a metric by name
func (r *SimpleRegistry) GetMetric(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metric, exists := r.metrics[name]
	return metric, exists
}

// GetAllMetrics returns all registered metrics
func (r *SimpleRegistry) GetAllMetrics() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var metrics []Metric
	now := time.Now()

	for _, m := range r.metrics {
		switch metric := m.(type) {
		case *simpleCounter:
			metrics = append(metrics, Metric{
				Name:      metric.name,
				Type:      MetricTypeCounter,
				Value:     metric.Get(),
				Labels:    metric.labels,
				Timestamp: now,
				Help:      metric.help,
			})

		case *simpleGauge:
			metrics = append(metrics, Metric{
				Name:      metric.name,
				Type:      MetricTypeGauge,
				Value:     metric.Get(),
				Labels:    metric.labels,
				Timestamp: now,
				Help:      metric.help,
			})

		case *simpleHistogram:
			metrics = append(metrics, Metric{
				Name:      metric.name,
				Type:      MetricTypeHistogram,
				Value:     metric.getSum(),
				Labels:    metric.labels,
				Timestamp: now,
				Help:      metric.help,
			})
		}
	}

	return metrics
}

// Unregister removes a metric
func (r *SimpleRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, name)
}

// Clear removes all metrics
func (r *SimpleRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = make(map[string]interface{})
}

// Helper methods

func (r *SimpleRegistry) buildMetricName(name string) string {
	if r.config.Namespace == "" {
		return name
	}

	if r.config.Subsystem == "" {
		return fmt.Sprintf("%s_%s", r.config.Namespace, name)
	}

	return fmt.Sprintf("%s_%s_%s", r.config.Namespace, r.config.Subsystem, name)
}

func (r *SimpleRegistry) mergeLabels(labels Labels) Labels {
	merged := make(Labels)

	// Add default labels from config
	for k, v := range r.config.Labels {
		merged[k] = v
	}

	// Add provided labels (overrides default)
	for k, v := range labels {
		merged[k] = v
	}

	return merged
}

// Simple implementations

type simpleCounter struct {
	mu     sync.RWMutex
	name   string
	help   string
	labels Labels
	value  float64
}

func (c *simpleCounter) Inc() {
	c.Add(1)
}

func (c *simpleCounter) Add(value float64) {
	if value < 0 {
		return // Counters can't decrease
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
}

func (c *simpleCounter) Get() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

func (c *simpleCounter) With(labels Labels) Counter {
	mergedLabels := make(Labels)
	for k, v := range c.labels {
		mergedLabels[k] = v
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}

	return &simpleCounter{
		name:   c.name,
		help:   c.help,
		labels: mergedLabels,
		value:  0,
	}
}

type simpleGauge struct {
	mu     sync.RWMutex
	name   string
	help   string
	labels Labels
	value  float64
}

func (g *simpleGauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

func (g *simpleGauge) Inc() {
	g.Add(1)
}

func (g *simpleGauge) Dec() {
	g.Sub(1)
}

func (g *simpleGauge) Add(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += value
}

func (g *simpleGauge) Sub(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value -= value
}

func (g *simpleGauge) Get() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

func (g *simpleGauge) With(labels Labels) Gauge {
	mergedLabels := make(Labels)
	for k, v := range g.labels {
		mergedLabels[k] = v
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}

	return &simpleGauge{
		name:   g.name,
		help:   g.help,
		labels: mergedLabels,
		value:  0,
	}
}

type simpleHistogram struct {
	mu      sync.RWMutex
	name    string
	help    string
	labels  Labels
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
}

func (h *simpleHistogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Find the appropriate bucket
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			break
		}
	}

	// Always increment the +Inf bucket
	h.counts[len(h.counts)-1]++
}

func (h *simpleHistogram) ObserveDuration(start time.Time) {
	h.Observe(time.Since(start).Seconds())
}

func (h *simpleHistogram) With(labels Labels) Histogram {
	mergedLabels := make(Labels)
	for k, v := range h.labels {
		mergedLabels[k] = v
	}
	for k, v := range labels {
		mergedLabels[k] = v
	}

	return &simpleHistogram{
		name:    h.name,
		help:    h.help,
		labels:  mergedLabels,
		buckets: h.buckets,
		counts:  make([]int64, len(h.buckets)+1),
		sum:     0,
		count:   0,
	}
}

func (h *simpleHistogram) getSum() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum
}

func (h *simpleHistogram) getCount() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

type simpleTimer struct {
	histogram Histogram
}

func (t *simpleTimer) Start() func() {
	start := time.Now()
	return func() {
		t.histogram.ObserveDuration(start)
	}
}

func (t *simpleTimer) ObserveDuration(start time.Time) {
	t.histogram.ObserveDuration(start)
}

// Convenience functions for the default registry

// NewCounter creates a counter in the default registry
func NewCounter(name, help string, labels Labels) Counter {
	return DefaultRegistry.NewCounter(name, help, labels)
}

// NewGauge creates a gauge in the default registry
func NewGauge(name, help string, labels Labels) Gauge {
	return DefaultRegistry.NewGauge(name, help, labels)
}

// NewHistogram creates a histogram in the default registry
func NewHistogram(name, help string, buckets []float64, labels Labels) Histogram {
	return DefaultRegistry.NewHistogram(name, help, buckets, labels)
}

// NewTimer creates a timer in the default registry
func NewTimer(histogram Histogram) Timer {
	return DefaultRegistry.NewTimer(histogram)
}

// GetAllMetrics returns all metrics from the default registry
func GetAllMetrics() []Metric {
	return DefaultRegistry.GetAllMetrics()
}

// SetDefaultConfig sets the default registry configuration
func SetDefaultConfig(config *Config) {
	DefaultRegistry.SetConfig(config)
}

// Clear clears all metrics from the default registry
func Clear() {
	DefaultRegistry.Clear()
}

// GetMetricsSorted returns all metrics sorted by name
func GetMetricsSorted() []Metric {
	metrics := GetAllMetrics()
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics
}
