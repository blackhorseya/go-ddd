package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/blackhorseya/go-ddd/internal/infrastructure/config"
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

	log.Printf("Starting %s in %s mode", cfg.App.Name, cfg.App.Env)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// TODO: Initialize your application components here
	// - Setup dependency injection
	// - Initialize infrastructure (database, redis, etc.)
	// - Start HTTP/gRPC servers

	log.Printf("HTTP server listening on %s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)
	log.Printf("gRPC server listening on %s:%d", cfg.Server.GRPC.Host, cfg.Server.GRPC.Port)

	// Wait for termination signal
	<-signals
	log.Println("Service shutting down...")
}
