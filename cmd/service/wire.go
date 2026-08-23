//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/blackhorseya/go-ddd/internal/identity"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/auth"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence"
	"github.com/blackhorseya/go-ddd/internal/shared"
	grpcserver "github.com/blackhorseya/go-ddd/internal/shared/adapter/grpc"
	httpserver "github.com/blackhorseya/go-ddd/internal/shared/adapter/http"
	"github.com/blackhorseya/go-ddd/internal/shared/infrastructure/config"
)

func provideHTTPServer(cfg *config.AppConfig) *httpserver.Server {
	return httpserver.NewServer(httpserver.ServerConfig{
		Host:         cfg.Server.HTTP.Host,
		Port:         cfg.Server.HTTP.Port,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
	}, cfg.App.Name)
}

func provideJWTConfig(cfg *config.AppConfig) auth.JWTConfig {
	return auth.JWTConfig{
		Secret:          cfg.Auth.JWT.Secret,
		AccessTokenTTL:  cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.JWT.RefreshTokenTTL,
	}
}

// providePersistenceConfig maps the load-time config aggregate onto the Identity
// BC's own persistence config, keeping the BC free of shared config types.
func providePersistenceConfig(cfg *config.AppConfig) persistence.Config {
	return persistence.Config{
		Driver:          cfg.Identity.Database.Driver,
		Host:            cfg.Identity.Database.Host,
		Port:            cfg.Identity.Database.Port,
		User:            cfg.Identity.Database.User,
		Password:        cfg.Identity.Database.Password,
		Name:            cfg.Identity.Database.Name,
		SSLMode:         cfg.Identity.Database.SSLMode,
		MaxOpenConns:    cfg.Identity.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Identity.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Identity.Database.ConnMaxLifetime,
	}
}

func provideGRPCServer(cfg *config.AppConfig) *grpcserver.Server {
	return grpcserver.NewServer(grpcserver.ServerConfig{
		Host: cfg.Server.GRPC.Host,
		Port: cfg.Server.GRPC.Port,
	}, cfg.IsDevelopment())
}

// InitializeApp builds the application with all dependencies wired up.
// The returned cleanup releases resources held by providers, such as the database pool.
func InitializeApp(cfg *config.AppConfig) (*App, func(), error) {
	wire.Build(
		provideHTTPServer,
		provideGRPCServer,
		provideJWTConfig,
		providePersistenceConfig,
		identity.ProviderSet,
		shared.ProviderSet,
		NewApp,
	)
	return nil, nil, nil
}
