//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/blackhorseya/go-ddd/internal/identity"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/auth"
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

func provideGRPCServer(cfg *config.AppConfig) *grpcserver.Server {
	return grpcserver.NewServer(grpcserver.ServerConfig{
		Host: cfg.Server.GRPC.Host,
		Port: cfg.Server.GRPC.Port,
	}, cfg.IsDevelopment())
}

// InitializeApp builds the application with all dependencies wired up.
func InitializeApp(cfg *config.AppConfig) (*App, error) {
	wire.Build(
		provideHTTPServer,
		provideGRPCServer,
		provideJWTConfig,
		identity.ProviderSet,
		shared.ProviderSet,
		NewApp,
	)
	return nil, nil
}
