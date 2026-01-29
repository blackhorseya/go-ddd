// Package main is the entry point of the service.
//
//	@title			Go DDD Service API
//	@version		1.2.0
//	@description	Go DDD 範本專案 API，實作 Clean Architecture 與 Domain-Driven Design 原則
//
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				輸入 Bearer token，格式: "Bearer {token}"
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/blackhorseya/go-ddd/internal/adapter/http"
	"github.com/blackhorseya/go-ddd/internal/infrastructure/config"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
	"github.com/blackhorseya/go-ddd/pkg/logx"
	"github.com/blackhorseya/go-ddd/pkg/otelx"
)

// 版本資訊，由 GoReleaser ldflags 注入
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	logger := logx.MustNew(&cfg.Log)
	logger.SetAsDefault()

	// Create base context with service info
	ctx := contextx.Background().
		WithService(cfg.App.Name).
		WithEnvironment(cfg.App.Env)

	// Initialize OpenTelemetry tracing
	otelCfg := otelx.DefaultConfig()
	otelCfg.ServiceName = cfg.App.Name
	otelCfg.Environment = cfg.App.Env
	tp, err := otelx.Setup(ctx, otelCfg)
	if err != nil {
		log.Fatalf("failed to setup tracing: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			ctx.Error("failed to shutdown tracer provider", "error", err)
		}
	}()

	ctx.Info("service starting",
		"version", Version,
		"commit", Commit,
		"build_date", Date,
		"http_host", cfg.Server.HTTP.Host,
		"http_port", cfg.Server.HTTP.Port,
		"grpc_host", cfg.Server.GRPC.Host,
		"grpc_port", cfg.Server.GRPC.Port,
	)

	// Setup signal handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create cancellable context for graceful shutdown
	runCtx, cancel := context.WithCancel(ctx)

	// Initialize HTTP server
	server := httpserver.NewServer(httpserver.ServerConfig{
		Host:         cfg.Server.HTTP.Host,
		Port:         cfg.Server.HTTP.Port,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
	}, cfg.App.Name)

	// Start HTTP server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := server.Run(runCtx); err != nil {
			errCh <- err
		}
	}()

	// Wait for termination signal or server error
	select {
	case sig := <-signals:
		ctx.Info("received signal", "signal", sig.String())
	case err := <-errCh:
		ctx.Error("server error", "error", err)
	}

	// Trigger graceful shutdown
	cancel()
	ctx.Info("service shutdown complete")
}
