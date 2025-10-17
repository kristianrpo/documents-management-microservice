package metrics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

// Create a shared instance to avoid duplicate registration
var sharedMetrics = metrics.NewPrometheusMetrics("test_service")

func TestNewPrometheusMetrics(t *testing.T) {
	t.Run("verify metrics structure", func(t *testing.T) {
		m := sharedMetrics

		assert.NotNil(t, m)
		assert.NotNil(t, m.HTTPRequestsTotal)
		assert.NotNil(t, m.HTTPRequestDuration)
		assert.NotNil(t, m.HTTPRequestsInFlight)
		assert.NotNil(t, m.UploadRequestsTotal)
		assert.NotNil(t, m.GetRequestsTotal)
		assert.NotNil(t, m.ListRequestsTotal)
		assert.NotNil(t, m.DeleteRequestsTotal)
		assert.NotNil(t, m.DeleteBulkRequestsTotal)
		assert.NotNil(t, m.TransferRequestsTotal)
		assert.NotNil(t, m.AuthRequestsTotal)
		assert.NotNil(t, m.AuthCompletedTotal)
		assert.NotNil(t, m.StorageUploadDuration)
		assert.NotNil(t, m.StorageDownloadDuration)
		assert.NotNil(t, m.StorageErrorsTotal)
		assert.NotNil(t, m.DatabaseQueryDuration)
		assert.NotNil(t, m.DatabaseErrorsTotal)
		assert.NotNil(t, m.MessagesPublishedTotal)
		assert.NotNil(t, m.MessagesConsumedTotal)
		assert.NotNil(t, m.MessageErrorsTotal)
	})
}

func TestPrometheusMetrics_RecordHTTPRequest(t *testing.T) {
	m := sharedMetrics

	t.Run("record successful request", func(t *testing.T) {
		m.RecordHTTPRequest("GET", "/api/documents", "200", 100*time.Millisecond)
		// No panic means success
	})

	t.Run("record failed request", func(t *testing.T) {
		m.RecordHTTPRequest("POST", "/api/upload", "500", 50*time.Millisecond)
		// No panic means success
	})

	t.Run("record multiple requests", func(t *testing.T) {
		m.RecordHTTPRequest("GET", "/health", "200", 10*time.Millisecond)
		m.RecordHTTPRequest("GET", "/health", "200", 5*time.Millisecond)
		m.RecordHTTPRequest("GET", "/health", "200", 8*time.Millisecond)
		// No panic means success
	})
}

func TestPrometheusMetrics_RecordDatabaseQuery(t *testing.T) {
	m := sharedMetrics

	t.Run("record successful query", func(t *testing.T) {
		m.RecordDatabaseQuery("GetByID", 50*time.Millisecond, nil)
		// No panic means success
	})

	t.Run("record failed query", func(t *testing.T) {
		err := errors.New("connection timeout")
		m.RecordDatabaseQuery("Save", 100*time.Millisecond, err)
		// No panic means success
	})

	t.Run("record various operations", func(t *testing.T) {
		m.RecordDatabaseQuery("ListByOwner", 200*time.Millisecond, nil)
		m.RecordDatabaseQuery("Delete", 30*time.Millisecond, nil)
		m.RecordDatabaseQuery("Update", 80*time.Millisecond, errors.New("constraint violation"))
		// No panic means success
	})
}

func TestPrometheusMetrics_RecordStorageOperation(t *testing.T) {
	m := sharedMetrics

	t.Run("record successful upload", func(t *testing.T) {
		m.RecordStorageOperation("upload", 500*time.Millisecond, nil)
		// No panic means success
	})

	t.Run("record failed upload", func(t *testing.T) {
		err := errors.New("S3 timeout")
		m.RecordStorageOperation("upload", 1*time.Second, err)
		// No panic means success
	})

	t.Run("record successful download", func(t *testing.T) {
		m.RecordStorageOperation("download", 300*time.Millisecond, nil)
		// No panic means success
	})

	t.Run("record failed download", func(t *testing.T) {
		err := errors.New("object not found")
		m.RecordStorageOperation("download", 100*time.Millisecond, err)
		// No panic means success
	})

	t.Run("record delete operation", func(t *testing.T) {
		m.RecordStorageOperation("delete", 50*time.Millisecond, nil)
		// No panic means success
	})

	t.Run("record operation with error", func(t *testing.T) {
		m.RecordStorageOperation("delete", 20*time.Millisecond, errors.New("access denied"))
		// No panic means success
	})
}

//nolint:dupl // Similar test structure for different metrics
func TestPrometheusMetrics_RecordMessagePublished(t *testing.T) {
	m := sharedMetrics

	t.Run("record successful publish", func(t *testing.T) {
		m.RecordMessagePublished("document.authentication.requested", nil)
		// No panic means success
	})

	t.Run("record failed publish", func(t *testing.T) {
		err := errors.New("connection closed")
		m.RecordMessagePublished("document.authentication.requested", err)
		// No panic means success
	})

	t.Run("record multiple publishes", func(t *testing.T) {
		m.RecordMessagePublished("user.transferred", nil)
		m.RecordMessagePublished("user.transferred", nil)
		m.RecordMessagePublished("user.transferred", errors.New("timeout"))
		// No panic means success
	})
}

//nolint:dupl // Similar test structure for different metrics
func TestPrometheusMetrics_RecordMessageConsumed(t *testing.T) {
	m := sharedMetrics

	t.Run("record successful consume", func(t *testing.T) {
		m.RecordMessageConsumed("document.authentication.completed", nil)
		// No panic means success
	})

	t.Run("record failed consume", func(t *testing.T) {
		err := errors.New("parsing error")
		m.RecordMessageConsumed("document.authentication.completed", err)
		// No panic means success
	})

	t.Run("record multiple consumes", func(t *testing.T) {
		m.RecordMessageConsumed("user.transferred", nil)
		m.RecordMessageConsumed("user.transferred", nil)
		m.RecordMessageConsumed("user.transferred", errors.New("invalid format"))
		// No panic means success
	})
}
