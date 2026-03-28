package interceptor

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

// NewServerTracingHandler returns a gRPC stats.Handler for server-side OpenTelemetry tracing.
func NewServerTracingHandler() stats.Handler {
	return otelgrpc.NewServerHandler()
}

// NewClientTracingHandler returns a gRPC stats.Handler for client-side OpenTelemetry tracing.
func NewClientTracingHandler() stats.Handler {
	return otelgrpc.NewClientHandler()
}
