package metrics

import (
	"time"
)

// Client represents a metrics collection interface
// This allows us to swap implementations (Prometheus, StatsD, etc.)
type Client interface {
	// Count tracks a counter metric
	Count(name string, value int64, tags []string, rate float64)
	
	// Gauge tracks a gauge metric (current value)
	Gauge(name string, value float64, tags []string, rate float64)
	
	// Histogram tracks the statistical distribution of a set of values
	Histogram(name string, value float64, tags []string, rate float64)
	
	// Timing tracks a timing metric
	Timing(name string, value time.Duration, tags []string, rate float64)
	
	// Close shuts down the metrics client
	Close() error
}

// NoopClient is a metrics client that does nothing
// Useful for development or when metrics are disabled
type NoopClient struct{}

// NewNoopClient creates a new no-op metrics client
func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

// Count implements Client.Count
func (c *NoopClient) Count(name string, value int64, tags []string, rate float64) {}

// Gauge implements Client.Gauge
func (c *NoopClient) Gauge(name string, value float64, tags []string, rate float64) {}

// Histogram implements Client.Histogram
func (c *NoopClient) Histogram(name string, value float64, tags []string, rate float64) {}

// Timing implements Client.Timing
func (c *NoopClient) Timing(name string, value time.Duration, tags []string, rate float64) {}

// Close implements Client.Close
func (c *NoopClient) Close() error {
	return nil
}
