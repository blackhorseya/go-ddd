package middleware

import (
	"time"

	"github.com/blackhorseya/go-ddd/pkg/logx"
	"github.com/gin-gonic/gin"
)

// Logging returns a middleware that logs HTTP requests.
func Logging(logger *logx.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		reqLogger := logger.With(
			"status", status,
			"method", method,
			"path", path,
			"query", query,
			"ip", clientIP,
			"latency", latency.String(),
			"user_agent", c.Request.UserAgent(),
		)

		if len(c.Errors) > 0 {
			reqLogger.Error(c.Errors.String())
			return
		}

		if status >= 500 {
			reqLogger.Error("server error")
		} else if status >= 400 {
			reqLogger.Warn("client error")
		} else {
			reqLogger.Info("request completed")
		}
	}
}
