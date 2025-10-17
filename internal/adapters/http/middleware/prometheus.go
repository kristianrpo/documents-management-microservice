package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// PrometheusMiddleware creates a Gin middleware that records HTTP metrics
func PrometheusMiddleware(metricsCollector *metrics.PrometheusMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		metricsCollector.HTTPRequestsInFlight.Inc()
		defer metricsCollector.HTTPRequestsInFlight.Dec()

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		metricsCollector.RecordHTTPRequest(method, endpoint, status, duration)
	}
}
