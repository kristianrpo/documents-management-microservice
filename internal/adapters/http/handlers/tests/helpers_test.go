package handlers_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

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
