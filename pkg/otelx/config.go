package otelx

// Config holds OpenTelemetry configuration.
type Config struct {
	// Enabled controls whether tracing is enabled.
	Enabled bool `mapstructure:"enabled"`

	// ServiceName is the name of the service for tracing.
	ServiceName string `mapstructure:"service_name"`

	// ServiceVersion is the version of the service.
	ServiceVersion string `mapstructure:"service_version"`

	// Environment is the deployment environment (development, staging, production).
	Environment string `mapstructure:"environment"`

	// Exporter specifies the exporter type: "otlp", "stdout", or "noop".
	Exporter string `mapstructure:"exporter"`

	// OTLP contains OTLP exporter configuration.
	OTLP OTLPConfig `mapstructure:"otlp"`

	// SampleRate is the sampling rate (0.0 to 1.0). 1.0 means sample all traces.
	SampleRate float64 `mapstructure:"sample_rate"`
}

// OTLPConfig holds OTLP exporter configuration.
type OTLPConfig struct {
	// Endpoint is the OTLP collector endpoint (e.g., "localhost:4318").
	Endpoint string `mapstructure:"endpoint"`

	// Insecure disables TLS for the connection.
	Insecure bool `mapstructure:"insecure"`

	// Protocol is the transport protocol: "http" or "grpc".
	Protocol string `mapstructure:"protocol"`
}

// DefaultConfig returns a default configuration for development.
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		ServiceName:    "go-ddd",
		ServiceVersion: "0.0.1",
		Environment:    "development",
		Exporter:       "noop",
		SampleRate:     1.0,
		OTLP: OTLPConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
			Protocol: "http",
		},
	}
}
