package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registerOnce sync.Once

	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func initMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(requestsTotal, requestDuration)
	})
}

func MetricsMiddleware() gin.HandlerFunc {
	initMetrics()

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		status := strconv.Itoa(c.Writer.Status())
		labels := prometheus.Labels{
			"method": c.Request.Method,
			"path":   path,
			"status": status,
		}

		requestsTotal.With(labels).Inc()
		requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}
}

func MetricsHandler() gin.HandlerFunc {
	initMetrics()
	return gin.WrapH(promhttp.Handler())
}
