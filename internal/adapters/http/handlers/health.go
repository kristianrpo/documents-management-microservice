package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/endpoints"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

// Ping godoc
// @Summary Health check endpoint
// @Description Returns the health status of the service. Use this endpoint to verify that the API is running and responsive.
// @Description This endpoint is useful for:
// @Description - Load balancer health checks
// @Description - Monitoring and alerting systems
// @Description - Kubernetes liveness/readiness probes
// @Tags health
// @Produce json
// @Success 200 {object} endpoints.HealthCheckResponse "Service is healthy and operational"
// @Router /healthz [get]
func (h *HealthHandler) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, endpoints.HealthCheckResponse{OK: true})
}
