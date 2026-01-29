package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/blackhorseya/go-ddd/internal/adapter/http/response"
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
//
//	@Summary		Liveness probe
//	@Description	檢查服務是否存活
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	response.Response{data=HealthStatus}
//	@Router			/healthz [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	response.OK(c, HealthStatus{Status: "ok"})
}

// Readiness handles readiness probe.
//
//	@Summary		Readiness probe
//	@Description	檢查服務是否準備好接收流量
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	response.Response{data=HealthStatus}
//	@Router			/readyz [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	// TODO: Add dependency checks (database, cache, etc.)
	response.OK(c, HealthStatus{Status: "ok"})
}
