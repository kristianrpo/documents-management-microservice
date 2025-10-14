package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

func (h *HealthHandler) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
