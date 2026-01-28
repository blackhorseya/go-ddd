package middleware

import (
	"time"

	"github.com/blackhorseya/go-ddd/pkg/contextx"
	"github.com/gin-gonic/gin"
)

// Logging returns a middleware that logs HTTP requests using contextx.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		ctx := contextx.From(c.Request.Context()).WithFields(
			"status", status,
			"method", method,
			"path", path,
			"query", query,
			"ip", clientIP,
			"latency", latency.String(),
			"user_agent", c.Request.UserAgent(),
		)

		if len(c.Errors) > 0 {
			ctx.Error(c.Errors.String())
			return
		}

		if status >= 500 {
			ctx.Error("server error")
		} else if status >= 400 {
			ctx.Warn("client error")
		} else {
			ctx.Info("request completed")
		}
	}
}
