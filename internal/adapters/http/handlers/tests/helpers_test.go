package handlers_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	apierrors "github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// createTestMetrics crea métricas con un namespace único para evitar conflicts en tests
func createTestMetrics(t *testing.T) *metrics.PrometheusMetrics {
	namespace := "test_documents_service_" + t.Name()

	return &metrics.PrometheusMetrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests",
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "HTTP requests in flight",
			},
		),
		GetRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "get_requests_total",
				Help:      "Total get requests",
			},
		),
		ListRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "list_requests_total",
				Help:      "Total list requests",
			},
		),
		UploadRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "upload_requests_total",
				Help:      "Total upload requests",
			},
		),
		DeleteRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "delete_requests_total",
				Help:      "Total delete requests",
			},
		),
		DeleteBulkRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "delete_bulk_requests_total",
				Help:      "Total delete bulk requests",
			},
		),
		TransferRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "transfer_requests_total",
				Help:      "Total transfer requests",
			},
		),
		AuthRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_requests_total",
				Help:      "Total auth requests",
			},
		),
		AuthCompletedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_completed_total",
				Help:      "Total auth completions",
			},
			[]string{"result"},
		),
		StorageUploadDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "storage_upload_duration_seconds",
				Help:      "Storage upload duration",
			},
		),
		StorageDownloadDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "storage_download_duration_seconds",
				Help:      "Storage download duration",
			},
		),
		StorageErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "storage_errors_total",
				Help:      "Total storage errors",
			},
			[]string{"operation"},
		),
		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "database_query_duration_seconds",
				Help:      "Database query duration",
			},
			[]string{"operation"},
		),
		DatabaseErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "database_errors_total",
				Help:      "Total database errors",
			},
			[]string{"operation"},
		),
		MessagesPublishedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_published_total",
				Help:      "Total messages published",
			},
			[]string{"queue"},
		),
		MessagesConsumedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_consumed_total",
				Help:      "Total messages consumed",
			},
			[]string{"queue"},
		),
		MessageErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "message_errors_total",
				Help:      "Total message errors",
			},
			[]string{"queue", "type"},
		),
	}
}

// newTestRouter creates a Gin router preconfigured with error handler and optional authenticated user
func newTestRouter(t *testing.T, withAuth bool, idCitizen int64) (*gin.Engine, *apierrors.ErrorHandler, *metrics.PrometheusMetrics) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	errMapper := apierrors.NewErrorMapper()
	errHandler := apierrors.NewErrorHandler(errMapper)
	metricsCollector := createTestMetrics(t)

	if withAuth {
		r.Use(func(c *gin.Context) {
			c.Set(string(middleware.UserContextKey), &middleware.UserClaims{IDCitizen: idCitizen})
			c.Next()
		})
	}

	return r, errHandler, metricsCollector
}

// runWithAuthenticatedRouter creates an authenticated test router (idCitizen=123456),
// registers routes via the setup callback, executes a request with the provided
// method and path, and returns the response recorder for assertions.
func runWithAuthenticatedRouter(t *testing.T, method, path string, setup func(r *gin.Engine)) *httptest.ResponseRecorder {
	t.Helper()
	r, _, _ := newTestRouter(t, true, 123456)

	setup(r)

	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
