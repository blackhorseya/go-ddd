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

	"github.com/blackhorseya/go-ddd/internal/shared/infrastructure/config"
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
	logger := logx.MustNew(&logx.Config{
		Level:     cfg.Log.Level,
		Format:    cfg.Log.Format,
		Output:    cfg.Log.Output,
		AddSource: cfg.Log.AddSource,
	})
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

	// Initialize application via Wire
	app := InitializeApp(cfg)

	// Setup signal handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create cancellable context for graceful shutdown
	runCtx, cancel := context.WithCancel(ctx)

	// Start application
	appErr := make(chan error, 1)
	go func() {
		appErr <- app.Run(runCtx)
	}()

	// Wait for termination signal or server error
	select {
	case sig := <-signals:
		ctx.Info("received signal", "signal", sig.String())
		cancel()
		// Wait for app.Run to finish gracefully
		if err := <-appErr; err != nil {
			ctx.Error("app shutdown error", "error", err)
		}
	case err := <-appErr:
		cancel()
		if err != nil {
			ctx.Error("server error", "error", err)
		}
	}

	ctx.Info("service shutdown complete")
}
