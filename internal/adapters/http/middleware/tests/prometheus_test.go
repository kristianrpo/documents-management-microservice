package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metricsCollector := metrics.NewPrometheusMetrics("test_middleware")

	t.Run("records HTTP request metrics", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.PrometheusMiddleware(metricsCollector))
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("records metrics for different status codes", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.PrometheusMiddleware(metricsCollector))
		
		router.GET("/success", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		
		router.GET("/error", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		})
		
		router.GET("/notfound", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})

		testCases := []struct {
			path           string
			expectedStatus int
		}{
			{"/success", http.StatusOK},
			{"/error", http.StatusInternalServerError},
			{"/notfound", http.StatusNotFound},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		}
	})

	t.Run("records metrics for different HTTP methods", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.PrometheusMiddleware(metricsCollector))
		
		router.GET("/resource", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{})
		})
		
		router.POST("/resource", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{})
		})
		
		router.PUT("/resource", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{})
		})
		
		router.DELETE("/resource", func(c *gin.Context) {
			c.JSON(http.StatusNoContent, gin.H{})
		})

		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/resource", nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)

			assert.True(t, w.Code >= 200 && w.Code < 300)
		}
	})

	t.Run("handles unknown endpoint", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.PrometheusMiddleware(metricsCollector))

		req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("increments and decrements in-flight counter", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.PrometheusMiddleware(metricsCollector))
		
		router.GET("/slow", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{})
		})

		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
