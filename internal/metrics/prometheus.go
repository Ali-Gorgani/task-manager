package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts the total number of HTTP requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestLatencyHistogram measures the latency of HTTP requests
	RequestLatencyHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_latency_histogram",
			Help:    "Histogram of HTTP request latencies",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// TasksCount tracks the current number of tasks
	TasksCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "tasks_count",
			Help: "Current number of tasks in the system",
		},
	)
)

// PrometheusMiddleware is a Gin middleware that collects metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get endpoint path (use route pattern, not actual path with IDs)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Record metrics
		RequestsTotal.WithLabelValues(
			c.Request.Method,
			endpoint,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		RequestLatencyHistogram.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(duration)
	}
}

// UpdateTasksCount updates the tasks count metric
func UpdateTasksCount(count int) {
	TasksCount.Set(float64(count))
}
