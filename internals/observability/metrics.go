package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metric
var reqDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Request latency in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path", "status"},
)

func init() {
	prometheus.MustRegister(reqDuration)
}

// WrapH style middleware for Gin

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		reqDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()), // convert int → string
		).Observe(duration)
	}
}
