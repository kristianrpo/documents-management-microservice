package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
)

func NewRouter(
	uploadHandler *handlers.DocumentUploadHandler,
	listHandler *handlers.DocumentListHandler,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/healthz", healthHandler.Ping)

	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/documents", uploadHandler.Upload)
		apiGroup.GET("/documents", listHandler.List)
	}

	return router
}
