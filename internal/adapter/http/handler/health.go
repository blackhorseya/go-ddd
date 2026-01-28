package handler

import (
	"github.com/blackhorseya/go-ddd/internal/adapter/http/response"
	"github.com/gin-gonic/gin"
)

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status string `json:"status"`
}

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
	response.OK(c, HealthStatus{Status: "ok"})
}

// Readiness handles readiness probe.
func (h *HealthHandler) Readiness(c *gin.Context) {
	// TODO: Add dependency checks (database, cache, etc.)
	response.OK(c, HealthStatus{Status: "ok"})
}
