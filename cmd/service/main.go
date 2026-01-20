package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/blackhorseya/go-ddd/internal/infrastructure/config"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
	"github.com/blackhorseya/go-ddd/pkg/logx"
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

	ctx.Info("service starting",
		"http_host", cfg.Server.HTTP.Host,
		"http_port", cfg.Server.HTTP.Port,
		"grpc_host", cfg.Server.GRPC.Host,
		"grpc_port", cfg.Server.GRPC.Port,
	)

	// Setup signal handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// TODO: Initialize your application components here
	// - Setup dependency injection
	// - Initialize infrastructure (database, redis, etc.)
	// - Start HTTP/gRPC servers

	// Wait for termination signal
	<-signals
	ctx.Info("service shutting down")
}
