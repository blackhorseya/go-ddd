package grpc

import "time"

// ServerConfig contains gRPC server configuration.
// This is defined in the adapter layer to avoid dependency on infrastructure layer.
type ServerConfig struct {
	Host            string
	Port            int
	ShutdownTimeout time.Duration // max time to wait for GracefulStop, default 10s
}

// ClientConfig contains gRPC client connection configuration.
// The connection is lazy (no upfront dial). Use per-RPC context deadlines for timeouts.
type ClientConfig struct {
	Target   string // e.g. "localhost:9090" or "dns:///service.namespace:9090"
	Insecure bool   // true = no TLS (for development)
}
