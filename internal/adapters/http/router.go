package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
)

func NewRouter(uploadHandler *handlers.DocumentUploadHandler, healthHandler *handlers.HealthHandler) *gin.Engine {
	router := gin.Default()

	router.GET("/healthz", healthHandler.Ping)

	apiGroup := router.Group("/api/v1")
	{
		apiGroup.POST("/documents", uploadHandler.Upload)
	}

	return router
}
