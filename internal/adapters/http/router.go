package http

import (
	"github.com/gin-gonic/gin"
	docs "github.com/kristianrpo/document-management-microservice/docs"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	if cfg.MetricsCollector != nil {
		router.Use(middleware.PrometheusMiddleware(cfg.MetricsCollector))
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes under /api/docs for API Gateway routing
	// Includes health, swagger, and all document endpoints
	apiGroup := router.Group("/api/docs")
	{
		// Health check endpoint
		apiGroup.GET("/healthz", cfg.HealthHandler.Ping)
		
		// Swagger documentation
		apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		
		// User-protected endpoints (require authenticated user with role USER)
		apiGroup.POST("/documents", cfg.JWTMiddleware.Authenticate(), cfg.JWTMiddleware.RequireRole("USER"), cfg.UploadHandler.Upload)
		apiGroup.GET("/documents", cfg.JWTMiddleware.Authenticate(), cfg.JWTMiddleware.RequireRole("USER"), cfg.ListHandler.List)
		apiGroup.GET("/documents/:id", cfg.JWTMiddleware.Authenticate(), cfg.JWTMiddleware.RequireRole("USER"), cfg.GetHandler.GetByID)
		apiGroup.DELETE("/documents/:id", cfg.JWTMiddleware.Authenticate(), cfg.JWTMiddleware.RequireRole("USER"), cfg.DeleteHandler.Delete)
		apiGroup.DELETE("/documents/user/delete-all", cfg.JWTMiddleware.Authenticate(), cfg.JWTMiddleware.RequireRole("USER"), cfg.DeleteAllHandler.DeleteAll)
		apiGroup.POST("/documents/:id/request-authentication", cfg.JWTMiddleware.Authenticate(), cfg.JWTMiddleware.RequireRole("USER"), cfg.RequestAuthHandler.RequestAuthentication)
		apiGroup.GET("/documents/transfer/:id_citizen", cfg.JWTMiddleware.AuthenticateClient(), cfg.JWTMiddleware.RequireClientCredentials(), cfg.TransferHandler.PrepareTransfer)
	}
	
	// Keep root health check for Kubernetes probes compatibility
	router.GET("/healthz", cfg.HealthHandler.Ping)

	return router
}
