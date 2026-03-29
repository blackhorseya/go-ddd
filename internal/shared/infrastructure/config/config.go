package config

import (
	"time"
)

// AppConfig holds all configuration for the service.
// This is the single load-time aggregate — only cmd/service/ should import this package.
type AppConfig struct {
	App      App            `mapstructure:"app"`
	Server   Server         `mapstructure:"server"`
	Auth     Auth           `mapstructure:"auth"`
	Log      LogConfig      `mapstructure:"log"`
	OTel     OTelConfig     `mapstructure:"otel"`
	Identity IdentityConfig `mapstructure:"identity"`
}

// Auth contains authentication configuration.
type Auth struct {
	JWT JWT `mapstructure:"jwt"`
}

// JWT contains JWT token configuration.
type JWT struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

// IdentityConfig holds configuration for the Identity bounded context.
type IdentityConfig struct {
	Database IdentityDatabase `mapstructure:"database"`
}

// IdentityDatabase holds database configuration specific to the Identity BC.
type IdentityDatabase struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// LogConfig contains logging configuration.
// This is defined in infrastructure layer to avoid dependency on pkg/logx.
type LogConfig struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	Output    string `mapstructure:"output"`
	AddSource bool   `mapstructure:"add_source"`
}

// OTelConfig contains OpenTelemetry configuration.
// This is defined in infrastructure layer to avoid dependency on pkg/otelx.
type OTelConfig struct {
	Enabled    bool       `mapstructure:"enabled"`
	Exporter   string     `mapstructure:"exporter"`
	SampleRate float64    `mapstructure:"sample_rate"`
	OTLP       OTLPConfig `mapstructure:"otlp"`
}

// OTLPConfig contains OTLP exporter configuration.
type OTLPConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Insecure bool   `mapstructure:"insecure"`
	Protocol string `mapstructure:"protocol"`
}

// App contains application-level configuration.
type App struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"` // development, staging, production
}

// Server contains HTTP/gRPC server configuration.
type Server struct {
	HTTP HTTP `mapstructure:"http"`
	GRPC GRPC `mapstructure:"grpc"`
}

// HTTP contains HTTP server configuration.
type HTTP struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// GRPC contains gRPC server configuration.
type GRPC struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// IsDevelopment returns true if running in development environment.
func (c *AppConfig) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production environment.
func (c *AppConfig) IsProduction() bool {
	return c.App.Env == "production"
}
