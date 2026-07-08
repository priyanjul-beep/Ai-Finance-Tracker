// Package monitoring exposes Prometheus metrics and health-check handlers.
package monitoring

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus counters/histograms used in the app.
type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	AIRequestsTotal     *prometheus.CounterVec
	AIRequestDuration   *prometheus.HistogramVec
	CacheHits           *prometheus.CounterVec
	ActiveUsers         prometheus.Gauge
}

// NewMetrics registers and returns all application metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "finance",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "path", "status"}),

		HTTPRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "finance",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path"}),

		AIRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "finance",
			Name:      "ai_requests_total",
			Help:      "Total AI provider calls.",
		}, []string{"provider", "operation", "status"}),

		AIRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "finance",
			Name:      "ai_request_duration_seconds",
			Help:      "AI provider call latency.",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		}, []string{"provider", "operation"}),

		CacheHits: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "finance",
			Name:      "cache_operations_total",
			Help:      "Cache hit/miss counter.",
		}, []string{"operation", "result"}),

		ActiveUsers: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "finance",
			Name:      "active_users",
			Help:      "Number of users with active sessions.",
		}),
	}
}

// PrometheusHandler returns the /metrics HTTP handler.
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// GinMiddleware records HTTP metrics for every request.
func GinMiddleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		status := http.StatusText(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// HealthResponse is the JSON body returned by health-check endpoints.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// HealthHandler returns 200 when the service is healthy.
func HealthHandler(checks map[string]func() error) gin.HandlerFunc {
	return func(c *gin.Context) {
		results := make(map[string]string, len(checks))
		allOK := true
		for name, fn := range checks {
			if err := fn(); err != nil {
				results[name] = "down: " + err.Error()
				allOK = false
			} else {
				results[name] = "up"
			}
		}

		status := "healthy"
		code := http.StatusOK
		if !allOK {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		c.JSON(code, HealthResponse{
			Status:    status,
			Timestamp: time.Now(),
			Checks:    results,
		})
	}
}
