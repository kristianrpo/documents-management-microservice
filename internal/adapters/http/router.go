package http

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
)

// NewRouter creates and configures a new HTTP router with all API endpoints
func NewRouter(
	uploadHandler *handlers.DocumentUploadHandler,
	listHandler *handlers.DocumentListHandler,
	getHandler *handlers.DocumentGetHandler,
	deleteHandler *handlers.DocumentDeleteHandler,
	deleteAllHandler *handlers.DocumentDeleteAllHandler,
	transferHandler *handlers.DocumentTransferHandler,
	requestAuthHandler *handlers.DocumentRequestAuthenticationHandler,
	healthHandler *handlers.HealthHandler,
	metricsCollector *metrics.PrometheusMetrics,
) *gin.Engine {
	router := gin.Default()

	if metricsCollector != nil {
		router.Use(middleware.PrometheusMiddleware(metricsCollector))
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/healthz", healthHandler.Ping)

	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/documents", uploadHandler.Upload)
		apiGroup.GET("/documents", listHandler.List)
		apiGroup.GET("/documents/:id", getHandler.GetByID)
		apiGroup.DELETE("/documents/:id", deleteHandler.Delete)
		apiGroup.DELETE("/documents/user/:id_citizen", deleteAllHandler.DeleteAll)
		apiGroup.GET("/documents/transfer/:id_citizen", transferHandler.PrepareTransfer)
		apiGroup.POST("/documents/:id/request-authentication", requestAuthHandler.RequestAuthentication)
	}

	return router
}
