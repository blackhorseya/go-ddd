package http

import "time"

// ServerConfig contains HTTP server configuration.
// This is defined in the adapter layer to avoid dependency on infrastructure layer.
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
