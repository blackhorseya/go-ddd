package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load reads configuration from file and environment variables.
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from config file if path provided
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	// Read from environment variables
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// MustLoad loads configuration and panics on error.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "go-ddd-service")
	v.SetDefault("app.env", "development")

	// HTTP server defaults
	v.SetDefault("server.http.host", "0.0.0.0")
	v.SetDefault("server.http.port", 8080)
	v.SetDefault("server.http.read_timeout", 30*time.Second)
	v.SetDefault("server.http.write_timeout", 30*time.Second)

	// gRPC server defaults
	v.SetDefault("server.grpc.host", "0.0.0.0")
	v.SetDefault("server.grpc.port", 9090)

	// Database defaults
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "app")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}
