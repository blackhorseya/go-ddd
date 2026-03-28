package config

import (
	"time"
)

// Config holds all configuration for the service.
type Config struct {
	App      App       `mapstructure:"app"`
	Server   Server    `mapstructure:"server"`
	Database Database  `mapstructure:"database"`
	Redis    Redis     `mapstructure:"redis"`
	Log      LogConfig `mapstructure:"log"`
}

// LogConfig contains logging configuration.
// This is defined in infrastructure layer to avoid dependency on pkg/logx.
type LogConfig struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	Output    string `mapstructure:"output"`
	AddSource bool   `mapstructure:"add_source"`
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

// Database contains database configuration.
type Database struct {
	Driver          string        `mapstructure:"driver"` // postgres, mysql
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

// Redis contains Redis configuration.
type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
