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

// RouterConfig holds all dependencies required to configure the HTTP router
type RouterConfig struct {
	UploadHandler      *handlers.DocumentUploadHandler
	ListHandler        *handlers.DocumentListHandler
	GetHandler         *handlers.DocumentGetHandler
	DeleteHandler      *handlers.DocumentDeleteHandler
	DeleteAllHandler   *handlers.DocumentDeleteAllHandler
	TransferHandler    *handlers.DocumentTransferHandler
	RequestAuthHandler *handlers.DocumentRequestAuthenticationHandler
	HealthHandler      *handlers.HealthHandler
	MetricsCollector   *metrics.PrometheusMetrics
	// JWT middleware instance (optional). If provided, it will be applied to
	// routes that require authentication (e.g. document upload).
	JWTMiddleware *middleware.JWTAuthMiddleware
}

// NewRouter creates and configures a new HTTP router with all API endpoints
func NewRouter(cfg *RouterConfig) *gin.Engine {
	router := gin.Default()

	if cfg.MetricsCollector != nil {
		router.Use(middleware.PrometheusMiddleware(cfg.MetricsCollector))
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/healthz", cfg.HealthHandler.Ping)

	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/documents", cfg.JWTMiddleware.Authenticate(), cfg.UploadHandler.Upload)
		apiGroup.GET("/documents", cfg.ListHandler.List)
		apiGroup.GET("/documents/:id", cfg.GetHandler.GetByID)
		apiGroup.DELETE("/documents/:id", cfg.DeleteHandler.Delete)
		apiGroup.DELETE("/documents/user/:id_citizen", cfg.DeleteAllHandler.DeleteAll)
		apiGroup.GET("/documents/transfer/:id_citizen", cfg.TransferHandler.PrepareTransfer)
		apiGroup.POST("/documents/:id/request-authentication", cfg.RequestAuthHandler.RequestAuthentication)
	}

	return router
}
