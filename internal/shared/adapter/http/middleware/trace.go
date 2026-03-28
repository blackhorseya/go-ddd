package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

const (
	// HeaderXTraceID is the header key for trace ID.
	HeaderXTraceID = "X-Trace-ID"
)

// Tracing returns the OpenTelemetry tracing middleware.
// It creates spans for each request and propagates trace context.
func Tracing(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// TraceID returns a middleware that sets the trace ID in the response header.
// This should be used after the Tracing middleware.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Set trace ID in response header after request processing
		traceID := contextx.GetTraceID(c.Request.Context())
		if traceID != "" {
			c.Header(HeaderXTraceID, traceID)
		}
	}
}
