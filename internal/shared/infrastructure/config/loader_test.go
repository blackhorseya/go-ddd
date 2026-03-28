package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	require.NoError(t, err)

	// App defaults
	assert.Equal(t, "go-ddd-service", cfg.App.Name)
	assert.Equal(t, "development", cfg.App.Env)

	// Server defaults
	assert.Equal(t, "0.0.0.0", cfg.Server.HTTP.Host)
	assert.Equal(t, 8080, cfg.Server.HTTP.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.HTTP.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.HTTP.WriteTimeout)
	assert.Equal(t, "0.0.0.0", cfg.Server.GRPC.Host)
	assert.Equal(t, 9090, cfg.Server.GRPC.Port)

	// Log defaults
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)

	// Identity BC defaults
	assert.Equal(t, "postgres", cfg.Identity.Database.Driver)
	assert.Equal(t, "localhost", cfg.Identity.Database.Host)
	assert.Equal(t, 5432, cfg.Identity.Database.Port)
	assert.Equal(t, "postgres", cfg.Identity.Database.User)
	assert.Equal(t, "identity", cfg.Identity.Database.Name)
	assert.Equal(t, "disable", cfg.Identity.Database.SSLMode)
	assert.Equal(t, 25, cfg.Identity.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Identity.Database.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.Identity.Database.ConnMaxLifetime)
}

func TestLoad_FromFile(t *testing.T) {
	content := `
app:
  name: test-service
  env: production
server:
  http:
    port: 9999
identity:
  database:
    host: db.example.com
    name: id_prod
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "test-service", cfg.App.Name)
	assert.Equal(t, "production", cfg.App.Env)
	assert.Equal(t, 9999, cfg.Server.HTTP.Port)
	assert.Equal(t, "db.example.com", cfg.Identity.Database.Host)
	assert.Equal(t, "id_prod", cfg.Identity.Database.Name)

	// Unset fields fall back to defaults
	assert.Equal(t, "0.0.0.0", cfg.Server.HTTP.Host)
	assert.Equal(t, 9090, cfg.Server.GRPC.Port)
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("APP_APP_NAME", "env-service")
	t.Setenv("APP_SERVER_HTTP_PORT", "3000")
	t.Setenv("APP_IDENTITY_DATABASE_HOST", "env-db")

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "env-service", cfg.App.Name)
	assert.Equal(t, 3000, cfg.Server.HTTP.Port)
	assert.Equal(t, "env-db", cfg.Identity.Database.Host)
}

func TestLoad_InvalidFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	assert.Error(t, err)
}

func TestMustLoad_Panics(t *testing.T) {
	assert.Panics(t, func() {
		MustLoad("/nonexistent/config.yaml")
	})
}

func TestAppConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{name: "development", env: "development", want: true},
		{name: "production", env: "production", want: false},
		{name: "staging", env: "staging", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AppConfig{App: App{Env: tt.env}}
			assert.Equal(t, tt.want, cfg.IsDevelopment())
		})
	}
}

func TestAppConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{name: "production", env: "production", want: true},
		{name: "development", env: "development", want: false},
		{name: "staging", env: "staging", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AppConfig{App: App{Env: tt.env}}
			assert.Equal(t, tt.want, cfg.IsProduction())
		})
	}
}
