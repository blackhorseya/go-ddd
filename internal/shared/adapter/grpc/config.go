package grpc

import "time"

// ServerConfig contains gRPC server configuration.
// This is defined in the adapter layer to avoid dependency on infrastructure layer.
type ServerConfig struct {
	Host string
	Port int
}

// ClientConfig contains gRPC client connection configuration.
type ClientConfig struct {
	Target   string        // e.g. "localhost:9090" or "dns:///service.namespace:9090"
	Insecure bool          // true = no TLS (for development)
	Timeout  time.Duration // connection establishment timeout, default 5s
}
