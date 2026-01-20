package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Register registers health check routes.
func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/healthz", h.Liveness)
	r.GET("/readyz", h.Readiness)
}

// Liveness handles liveness probe.
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness handles readiness probe.
func (h *HealthHandler) Readiness(c *gin.Context) {
	// TODO: Add dependency checks (database, cache, etc.)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
