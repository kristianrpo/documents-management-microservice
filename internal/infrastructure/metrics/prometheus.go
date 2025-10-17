package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics holds all Prometheus metric collectors for the application
type PrometheusMetrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	UploadRequestsTotal     prometheus.Counter
	GetRequestsTotal        prometheus.Counter
	ListRequestsTotal       prometheus.Counter
	DeleteRequestsTotal     prometheus.Counter
	DeleteBulkRequestsTotal prometheus.Counter
	TransferRequestsTotal   prometheus.Counter
	AuthRequestsTotal       prometheus.Counter
	AuthCompletedTotal      *prometheus.CounterVec

	StorageUploadDuration   prometheus.Histogram
	StorageDownloadDuration prometheus.Histogram
	StorageErrorsTotal      *prometheus.CounterVec

	DatabaseQueryDuration *prometheus.HistogramVec
	DatabaseErrorsTotal   *prometheus.CounterVec

	MessagesPublishedTotal *prometheus.CounterVec
	MessagesConsumedTotal  *prometheus.CounterVec
	MessageErrorsTotal     *prometheus.CounterVec
}

// NewPrometheusMetrics creates and registers all Prometheus metrics
func NewPrometheusMetrics(namespace string) *PrometheusMetrics {
	if namespace == "" {
		namespace = "documents_service"
	}

	return &PrometheusMetrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests by method, endpoint, and status code",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		UploadRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "upload_requests_total",
				Help:      "Total number of document upload requests",
			},
		),
		GetRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "get_requests_total",
				Help:      "Total number of individual document get requests (GET /documents/:id)",
			},
		),
		ListRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "list_requests_total",
				Help:      "Total number of document list requests (GET /documents)",
			},
		),
		DeleteRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "delete_requests_total",
				Help:      "Total number of individual document delete requests (DELETE /documents/:id)",
			},
		),
		DeleteBulkRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "delete_bulk_requests_total",
				Help:      "Total number of bulk delete requests (DELETE /documents/user/:id)",
			},
		),
		TransferRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "transfer_requests_total",
				Help:      "Total number of document transfer requests",
			},
		),
		AuthRequestsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_requests_total",
				Help:      "Total number of document authentication requests",
			},
		),
		AuthCompletedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_completed_total",
				Help:      "Total number of authentication completions by result",
			},
			[]string{"result"},
		),

		StorageUploadDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "storage_upload_duration_seconds",
				Help:      "Duration of storage upload operations in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30},
			},
		),
		StorageDownloadDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "storage_download_duration_seconds",
				Help:      "Duration of storage download operations in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30},
			},
		),
		StorageErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "storage_errors_total",
				Help:      "Total number of storage errors by operation",
			},
			[]string{"operation"},
		),

		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "database_query_duration_seconds",
				Help:      "Duration of database queries in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation"},
		),
		DatabaseErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "database_errors_total",
				Help:      "Total number of database errors by operation",
			},
			[]string{"operation"},
		),

		MessagesPublishedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_published_total",
				Help:      "Total number of messages published to RabbitMQ by queue",
			},
			[]string{"queue"},
		),
		MessagesConsumedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_consumed_total",
				Help:      "Total number of messages consumed from RabbitMQ by queue",
			},
			[]string{"queue"},
		),
		MessageErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "message_errors_total",
				Help:      "Total number of message processing errors by queue",
			},
			[]string{"queue", "type"},
		),
	}
}

// RecordHTTPRequest records an HTTP request metric
func (m *PrometheusMetrics) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordDatabaseQuery records a database query metric
func (m *PrometheusMetrics) RecordDatabaseQuery(operation string, duration time.Duration, err error) {
	m.DatabaseQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
	if err != nil {
		m.DatabaseErrorsTotal.WithLabelValues(operation).Inc()
	}
}

// RecordStorageOperation records a storage operation metric
func (m *PrometheusMetrics) RecordStorageOperation(operation string, duration time.Duration, err error) {
	switch operation {
	case "upload":
		m.StorageUploadDuration.Observe(duration.Seconds())
	case "download":
		m.StorageDownloadDuration.Observe(duration.Seconds())
	}

	if err != nil {
		m.StorageErrorsTotal.WithLabelValues(operation).Inc()
	}
}

// RecordMessagePublished records a message publish metric
func (m *PrometheusMetrics) RecordMessagePublished(queue string, err error) {
	if err != nil {
		m.MessageErrorsTotal.WithLabelValues(queue, "publish").Inc()
	} else {
		m.MessagesPublishedTotal.WithLabelValues(queue).Inc()
	}
}

// RecordMessageConsumed records a message consume metric
func (m *PrometheusMetrics) RecordMessageConsumed(queue string, err error) {
	if err != nil {
		m.MessageErrorsTotal.WithLabelValues(queue, "consume").Inc()
	} else {
		m.MessagesConsumedTotal.WithLabelValues(queue).Inc()
	}
}
